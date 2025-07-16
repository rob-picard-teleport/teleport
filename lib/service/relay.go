// Teleport
// Copyright (C) 2025 Gravitational, Inc.
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

package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gravitational/trace"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"

	"github.com/gravitational/teleport"
	headerv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/header/v1"
	presencev1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/presence/v1"
	relayv1alpha "github.com/gravitational/teleport/api/gen/proto/go/teleport/relay/v1alpha"
	transportv1pb "github.com/gravitational/teleport/api/gen/proto/go/teleport/transport/v1"
	apitypes "github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/grpc/interceptors"
	"github.com/gravitational/teleport/lib/agentless"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/auth/authclient"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/proxy"
	"github.com/gravitational/teleport/lib/relayapi"
	"github.com/gravitational/teleport/lib/relaytunnel"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/srv"
	"github.com/gravitational/teleport/lib/srv/transport/transportv1"
	"github.com/gravitational/teleport/lib/utils"
)

func (process *TeleportProcess) initRelay() {
	process.RegisterWithAuthServer(apitypes.RoleRelay, RelayIdentityEvent)
	process.RegisterCriticalFunc("relay.run", process.runRelayService)
}

func (process *TeleportProcess) runRelayService() (retErr error) {
	log := process.logger.With(teleport.ComponentKey, teleport.Component(teleport.ComponentRelay, process.id))

	defer func() {
		if err := process.closeImportedDescriptors(teleport.ComponentRelay); err != nil {
			log.WarnContext(process.ExitContext(), "Failed closing imported file descriptors.", "error", err)
		}
	}()

	conn, err := process.WaitForConnector(RelayIdentityEvent, log)
	if conn == nil {
		return trace.Wrap(err)
	}
	defer conn.Close()

	accessPoint, err := process.newLocalCacheForRelay(conn.Client, []string{teleport.ComponentRelay})
	if err != nil {
		return err
	}
	defer accessPoint.Close()

	asyncEmitter, err := process.NewAsyncEmitter(conn.Client)
	if err != nil {
		return trace.Wrap(err)
	}
	defer asyncEmitter.Close()

	lockWatcher, err := services.NewLockWatcher(process.ExitContext(), services.LockWatcherConfig{
		ResourceWatcherConfig: services.ResourceWatcherConfig{
			Component: teleport.ComponentRelay,
			Logger:    log.With("subcomponent", "lock_watcher"),
			Client:    conn.Client,
		},
	})
	if err != nil {
		return trace.Wrap(err)
	}
	defer lockWatcher.Close()

	authorizer, err := authz.NewAuthorizer(authz.AuthorizerOpts{
		ClusterName:   conn.clusterName,
		AccessPoint:   accessPoint,
		LockWatcher:   lockWatcher,
		Logger:        log.With("subcomponent", "authorizer"),
		PermitCaching: process.Config.CachePolicy.Enabled,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	connMonitor, err := srv.NewConnectionMonitor(srv.ConnectionMonitorConfig{
		AccessPoint:    accessPoint,
		LockWatcher:    lockWatcher,
		Clock:          process.Clock,
		ServerID:       conn.hostID,
		Emitter:        asyncEmitter,
		EmitterContext: process.ExitContext(),
		Logger:         log.With("subcomponent", "conn_monitor"),
	})
	if err != nil {
		return trace.Wrap(err)
	}

	nodeWatcher, err := services.NewNodeWatcher(process.ExitContext(), services.NodeWatcherConfig{
		ResourceWatcherConfig: services.ResourceWatcherConfig{
			Component:    teleport.ComponentRelay,
			Logger:       log.With("subcomponent", "node_watcher"),
			Client:       conn.Client,
			MaxStaleness: time.Minute,
		},
		NodesGetter: accessPoint,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	defer nodeWatcher.Close()

	tunnelServer, err := relaytunnel.NewServer(relaytunnel.ServerConfig{
		Log: log.With("subcomponent", "tunnel_server"),
		GetCertificate: func(ctx context.Context) (*tls.Certificate, error) {
			return conn.serverGetCertificate()
		},
		GetPool: func(ctx context.Context) (*x509.CertPool, error) {
			pool, _, err := authclient.ClientCertPool(ctx, accessPoint, conn.clusterName, apitypes.HostCA)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			return pool, nil
		},
		Ciphersuites: process.Config.CipherSuites,
	})
	if err != nil {
		return trace.Wrap(err)
	}
	defer tunnelServer.Close()

	apiTLSConfig, err := conn.ServerTLSConfig(process.Config.CipherSuites)
	if err != nil {
		return trace.Wrap(err)
	}
	apiTLSConfig.NextProtos = []string{"h2"}
	apiTLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	apiTLSConfig.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
		pool, hauci, _, err := authclient.DefaultClientCertPool(chi.Context(), accessPoint, conn.clusterName)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		utils.RefreshTLSConfigTickets(apiTLSConfig)
		c := apiTLSConfig.Clone()

		c.ClientCAs = pool
		c.VerifyPeerCertificate = (&auth.HostAndUserCAPoolInfo{
			Pool:    pool,
			CATypes: hauci,
		}).VerifyPeerCertificate
		return c, nil
	}

	apiCreds, err := auth.NewTransportCredentials(auth.TransportCredentialsConfig{
		TransportCredentials: credentials.NewTLS(apiTLSConfig),
		UserGetter: &authz.Middleware{
			ClusterName: conn.clusterName,
		},
		Authorizer:        authorizer,
		GetAuthPreference: accessPoint.GetAuthPreference,
	})
	if err != nil {
		return trace.Wrap(err)
	}

	apiListener, err := process.importOrCreateListener(ListenerRelayAPI, process.Config.Relay.APIListenAddr)
	if err != nil {
		return trace.Wrap(err)
	}
	defer apiListener.Close()

	tunnelListener, err := process.importOrCreateListener(ListenerRelayTunnel, process.Config.Relay.TunnelListenAddr)
	if err != nil {
		return trace.Wrap(err)
	}
	defer tunnelListener.Close()
	go tunnelServer.ServeTLSTunnelListener(tunnelListener)

	relayRouter, err := proxy.NewRelayRouter(conn.clusterName, tunnelServer.Dial, accessPoint, nodeWatcher)
	if err != nil {
		return err
	}

	transportService, err := transportv1.NewService(transportv1.ServerConfig{
		FIPS:   process.Config.FIPS,
		Logger: log.With("subcomponent", "transport_service"),
		Dialer: relayRouter,
		SignerFn: func(*authz.Context, string) agentless.SignerCreator {
			return func(context.Context, agentless.LocalAccessPoint, agentless.CertGenerator) (ssh.Signer, error) {
				// we should never attempt to connect to an agentless server
				return nil, trace.Errorf("connections to agentless servers are not supported (this is a bug)")
			}
		},
		ConnectionMonitor: connMonitor,
		LocalAddr:         apiListener.Addr(),
	})
	if err != nil {
		return trace.Wrap(err)
	}

	apiServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.GRPCServerUnaryErrorInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.GRPCServerStreamErrorInterceptor,
		),
		grpc.Creds(apiCreds),
	)
	defer apiServer.Stop()

	transportv1pb.RegisterTransportServiceServer(apiServer, transportService)
	relayv1alpha.RegisterDiscoveryServiceServer(apiServer, &relayapi.StaticDiscoverServiceServer{
		RelayGroup:            process.Config.Relay.RelayGroup,
		TunnelPublicAddr:      process.Config.Relay.TunnelPublicAddr,
		TargetConnectionCount: process.Config.Relay.TargetConnectionCount,
	})

	go apiServer.Serve(apiListener)

	nonce := uuid.NewString()
	var relayServer atomic.Pointer[presencev1.RelayServer]
	relayServer.Store(&presencev1.RelayServer{
		Kind:    apitypes.KindRelayServer,
		SubKind: "",
		Version: apitypes.V1,
		Metadata: &headerv1.Metadata{
			Name: process.Config.HostUUID,
		},
		Spec: &presencev1.RelayServer_Spec{
			Hostname:   process.Config.Hostname,
			RelayGroup: process.Config.Relay.RelayGroup,

			Nonce: nonce,
		},
	})

	hb, err := srv.NewRelayServerHeartbeat(srv.HeartbeatV2Config[*presencev1.RelayServer]{
		InventoryHandle: process.inventoryHandle,
		GetResource: func(context.Context) (*presencev1.RelayServer, error) {
			return relayServer.Load(), nil
		},

		// there's no fallback announce mode, the relay service only works with
		// clusters recent enough to support relay heartbeats through the ICS
		Announcer: nil,

		OnHeartbeat: process.OnHeartbeat(teleport.ComponentRelay),
	}, log.With("subcomponent", "relay_heartbeat"))
	if err != nil {
		return trace.Wrap(err)
	}
	go hb.Run()
	defer hb.Close()

	if err := process.closeImportedDescriptors(teleport.ComponentRelay); err != nil {
		log.WarnContext(process.ExitContext(), "Failed closing imported file descriptors", "error", err)
	}

	process.BroadcastEvent(Event{Name: RelayReady})
	log.InfoContext(process.ExitContext(), "The relay service has successfully started", "nonce", nonce)

	exitEvent, _ := process.WaitForEvent(process.ExitContext(), TeleportExitEvent)
	ctx, _ := exitEvent.Payload.(context.Context)
	if ctx == nil {
		// if we're here it's because we got an ungraceful exit event or
		// WaitForEvent errored out because of the ungraceful shutdown; either
		// way, process.ExitContext() is a done context and all operations
		// should get canceled immediately
		log.InfoContext(ctx, "Stopping the relay service ungracefully")
		ctx = process.ExitContext()
	} else {
		log.InfoContext(ctx, "Stopping the relay service")
	}

	{
		r := proto.CloneOf(relayServer.Load())
		r.GetSpec().Terminating = true
		relayServer.Store(r)
	}

	tunnelServer.SetTerminating()

	if delay := process.Config.Relay.ShutdownDelay; delay > 0 {
		log.InfoContext(ctx, "Waiting for the shutdown delay", "delay", delay.String())
		select {
		case <-ctx.Done():
		case <-time.After(delay):
		}
	}

	log.DebugContext(ctx, "Stopping servers")
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// TODO(espadolini): let connections continue (for a time?) before
		// abruptly terminating them right after the shutdown delay
		tunnelServer.Close()
		return nil
	})
	eg.Go(func() error {
		defer context.AfterFunc(egCtx, apiServer.Stop)()
		apiServer.GracefulStop()
		return nil
	})
	warnOnErr(egCtx, eg.Wait(), log)

	warnOnErr(ctx, hb.Close(), log)
	warnOnErr(ctx, conn.Close(), log)

	log.InfoContext(ctx, "The relay service has stopped")

	return nil
}
