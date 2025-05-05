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
	"crypto/x509"

	"github.com/gravitational/trace"
	"golang.org/x/crypto/ssh"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
)

type remoteSSHProvider struct {
	clt *clientApplicationServiceClient
}

func newRemoteSSHProvider(clt *clientApplicationServiceClient) *remoteSSHProvider {
	return &remoteSSHProvider{
		clt: clt,
	}
}

func (p *remoteSSHProvider) ResolveSSHInfo(ctx context.Context, fqdn string) (*vnetv1.SshInfo, error) {
	return p.clt.ResolveSSHInfo(ctx, fqdn)
}

func (p *remoteSSHProvider) TeleportClientTLSConfig(ctx context.Context, profile, clusterName string) (*tls.Config, error) {
	cert, err := p.clt.UserMTLSCert(ctx, profile)
	if err != nil {
		return nil, trace.Wrap(err, "fetching user mTLS cert for profile %s", profile)
	}
	x509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, trace.Wrap(err, "parsing x509 certificate for user")
	}
	signer := &remoteSigner{
		pub: x509Cert.PublicKey,
		sendRequest: func(ctx context.Context, req *vnetv1.SignRequest) ([]byte, error) {
			return p.clt.SignForUserMTLS(ctx, &vnetv1.SignForUserMTLSRequest{
				Profile: profile,
				Sign:    req,
			})
		},
	}
	tlsCert := tls.Certificate{
		Certificate: [][]byte{cert},
		PrivateKey:  signer,
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ServerName:         x509Cert.Issuer.CommonName,
		InsecureSkipVerify: true,
	}, nil
}

func (p *remoteSSHProvider) UserSSHConfig(ctx context.Context, sshInfo *vnetv1.SshInfo, username string) (*ssh.ClientConfig, error) {
	cert, err := p.clt.ReissueSSHCert(ctx, sshInfo, username)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	sshPub, _, _, _, err := ssh.ParseAuthorizedKey(cert)
	if err != nil {
		return nil, trace.Wrap(err, "parsing SSH certificate")
	}
	sshCert, ok := sshPub.(*ssh.Certificate)
	if !ok {
		return nil, trace.BadParameter("expected ssh.Certificate, got %T", sshCert)
	}
	signer := &remoteSigner{
		pub: sshCert.Key.(ssh.CryptoPublicKey).CryptoPublicKey(),
		sendRequest: func(ctx context.Context, req *vnetv1.SignRequest) ([]byte, error) {
			return p.clt.SignForSSH(ctx, &vnetv1.SignForSshRequest{
				SshKey: sshInfo.GetSshKey(),
				User:   username,
				Sign:   req,
			})
		},
	}
	sshSigner, err := ssh.NewSignerFromSigner(signer)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	sshSigner, err = ssh.NewCertSigner(sshCert, sshSigner)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &ssh.ClientConfig{
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(sshSigner)},
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
