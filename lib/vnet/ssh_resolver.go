// Teleport
// Copyright (C) 2024 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package vnet

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"

	"github.com/gravitational/teleport"
	proxyclient "github.com/gravitational/teleport/api/client/proxy"
	tracessh "github.com/gravitational/teleport/api/observability/tracing/ssh"
	"github.com/gravitational/teleport/api/utils/grpc/interceptors"
	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
	"github.com/gravitational/teleport/lib/cryptosuites"
)

type sshProvider interface {
	ResolveSSHInfo(ctx context.Context, fqdn string) (*vnetv1.SshInfo, error)
	TeleportClientTLSConfig(ctx context.Context, profileName, clusterName string) (*tls.Config, error)
	UserSSHConfig(ctx context.Context, sshInfo *vnetv1.SshInfo, username string) (*ssh.ClientConfig, error)
}

type sshResolver struct {
	sshProvider sshProvider
	log         *slog.Logger
	clock       clockwork.Clock
	hostSigner  ssh.Signer
}

func newSSHResolver(sshProvider sshProvider, clock clockwork.Clock) *sshResolver {
	hostKey, err := cryptosuites.GenerateKeyWithAlgorithm(cryptosuites.Ed25519)
	if err != nil {
		panic(err)
	}
	hostSigner, err := ssh.NewSignerFromSigner(hostKey)
	if err != nil {
		panic(err)
	}
	return &sshResolver{
		sshProvider: sshProvider,
		log:         log.With(teleport.ComponentKey, "VNet.SSHResolver"),
		clock:       clock,
		hostSigner:  hostSigner,
	}
}

func (r sshResolver) resolveTCPHandler(ctx context.Context, fqdn string) (*tcpHandlerSpec, error) {
	sshInfo, err := r.sshProvider.ResolveSSHInfo(ctx, fqdn)
	if err != nil {
		return nil, err
	}
	sshHandler := r.newSSHHandler(ctx, sshInfo)
	return &tcpHandlerSpec{
		ipv4CIDRRange: sshInfo.Ipv4CidrRange,
		tcpHandler:    sshHandler,
	}, nil
}

func (r *sshResolver) newSSHHandler(ctx context.Context, sshInfo *vnetv1.SshInfo) *sshHandler {
	return &sshHandler{
		sshInfo:     sshInfo,
		sshProvider: r.sshProvider,
		hostSigner:  r.hostSigner,
	}
}

type sshHandler struct {
	sshInfo     *vnetv1.SshInfo
	sshProvider sshProvider
	hostSigner  ssh.Signer

	fg              singleflight.Group
	sshClientConfig sync.Map
}

func (h *sshHandler) handleTCPConnector(ctx context.Context, localPort uint16, connector func() (net.Conn, error)) error {
	targetTCPConn, err := h.dialTargetTCP(ctx)
	if err != nil {
		return trace.Wrap(err, "dialing SSH host %s", h.sshInfo.SshKey.Hostname)
	}
	defer targetTCPConn.Close()

	localTCPConn, err := connector()
	if err != nil {
		return trace.Wrap(err, "unwrapping local VNet TCP conn")
	}
	defer localTCPConn.Close()

	var targetClient *ssh.Client
	var preAuthConn ssh.ServerPreAuthConn
	serverConfig := &ssh.ServerConfig{
		PreAuthConnCallback: func(conn ssh.ServerPreAuthConn) {
			preAuthConn = conn
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			targetClient, err = h.dialTargetSSH(ctx, targetTCPConn, conn.User())
			if err != nil {
				err = trace.Wrap(err, "dialing target node")
				go func() { preAuthConn.SendAuthBanner(err.Error()) }()
				time.Sleep(500 * time.Millisecond)
				return nil, err
			}
			return nil, nil
		},
	}
	serverConfig.AddHostKey(h.hostSigner)
	serverConn, chans, requests, err := ssh.NewServerConn(localTCPConn, serverConfig)
	if err != nil {
		if targetClient != nil {
			targetClient.Close()
		}
		return trace.Wrap(err, "accepting incoming SSH conn")
	}
	defer func() {
		serverConn.Close()
		targetClient.Close()
	}()

	return trace.Wrap(proxySSHConnection(ctx, targetClient, chans, requests), "proxying SSH connection")
}

func (h *sshHandler) dialTargetTCP(ctx context.Context) (net.Conn, error) {
	proxyClientConfig := proxyclient.ClientConfig{
		ProxyAddress: h.sshInfo.DialOptions.WebProxyAddr,
		TLSConfigFunc: func(cluster string) (*tls.Config, error) {
			return h.sshProvider.TeleportClientTLSConfig(ctx, h.sshInfo.SshKey.Profile, cluster)
		},
		UnaryInterceptors:  []grpc.UnaryClientInterceptor{interceptors.GRPCClientUnaryErrorInterceptor},
		StreamInterceptors: []grpc.StreamClientInterceptor{interceptors.GRPCClientStreamErrorInterceptor},
		// This empty SSH client config should never be used, we dial to the
		// proxy over TLS.
		SSHConfig:               &ssh.ClientConfig{},
		InsecureSkipVerify:      h.sshInfo.DialOptions.InsecureSkipVerify,
		ALPNConnUpgradeRequired: h.sshInfo.DialOptions.AlpnConnUpgradeRequired,
	}
	pclt, err := proxyclient.NewClient(ctx, proxyClientConfig)
	if err != nil {
		return nil, trace.Wrap(err, "creating proxy client")
	}
	target := h.sshInfo.GetSshKey().GetHostname() + ":0"
	log.DebugContext(ctx, "Dialing target host",
		"target", target,
	)
	targetConn, _, err := pclt.DialHost(ctx, target, h.sshInfo.Cluster, nil /*keyRing*/)
	return targetConn, trace.Wrap(err)
}

func (h *sshHandler) dialTargetSSH(ctx context.Context, tcpConn net.Conn, username string) (*ssh.Client, error) {
	sshClientConfig, err := h.userSSHConfig(ctx, username)
	if err != nil {
		return nil, trace.Wrap(err, "getting user SSH client config")
	}
	sshconn, chans, reqs, err := tracessh.NewClientConn(ctx, tcpConn, h.sshInfo.SshKey.Hostname, sshClientConfig)
	if err != nil {
		log.InfoContext(ctx, "Error dialing target SSH node, retrying with a fresh user cert", "error", err)
		sshClient, err := h.retryDialTargetSSH(ctx, username)
		return sshClient, trace.Wrap(err)
	}
	log.DebugContext(ctx, "Dialed target SSH node", "target", h.sshInfo.SshKey.Hostname)
	return ssh.NewClient(sshconn, chans, reqs), nil
}

func (h *sshHandler) retryDialTargetSSH(ctx context.Context, username string) (*ssh.Client, error) {
	h.sshClientConfig.Delete(username)
	sshClientConfig, err := h.userSSHConfig(ctx, username)
	if err != nil {
		return nil, trace.Wrap(err, "getting fresh SSH client config")
	}
	// We need a fresh TCP connection to the target.
	tcpConn, err := h.dialTargetTCP(ctx)
	if err != nil {
		return nil, trace.Wrap(err, "redialing target with fresh SSH cert")
	}
	sshconn, chans, reqs, err := tracessh.NewClientConn(ctx, tcpConn, h.sshInfo.SshKey.Hostname, sshClientConfig)
	if err != nil {
		return nil, trace.Wrap(err, "dialing target SSH node with fresh user cert")
	}
	return ssh.NewClient(sshconn, chans, reqs), nil
}

func (h *sshHandler) userSSHConfig(ctx context.Context, username string) (*ssh.ClientConfig, error) {
	if c, ok := h.sshClientConfig.Load(username); ok {
		return c.(*ssh.ClientConfig), nil
	}
	_, err, _ := h.fg.Do(username, func() (any, error) {
		if c, ok := h.sshClientConfig.Load(username); ok {
			return c.(*ssh.ClientConfig), nil
		}
		c, err := h.sshProvider.UserSSHConfig(ctx, h.sshInfo, username)
		if err != nil {
			return nil, trace.Wrap(err, "getting user SSH client config")
		}
		h.sshClientConfig.Store(username, c)
		return nil, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	c, ok := h.sshClientConfig.Load(username)
	if !ok {
	}
	return c.(*ssh.ClientConfig), nil
}
