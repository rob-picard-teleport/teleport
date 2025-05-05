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

package vnet

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/gravitational/trace"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
)

// remoteAppProvider implements appProvider when the client application is
// available over gRPC.
type remoteAppProvider struct {
	clt *clientApplicationServiceClient
}

func newRemoteAppProvider(clt *clientApplicationServiceClient) *remoteAppProvider {
	return &remoteAppProvider{
		clt: clt,
	}
}

// ResolveAppInfo implements [appProvider.ResolveAppInfo].
func (p *remoteAppProvider) ResolveAppInfo(ctx context.Context, fqdn string) (*vnetv1.AppInfo, error) {
	appInfo, err := p.clt.ResolveAppInfo(ctx, fqdn)
	// Avoid wrapping errNoTCPHandler, no need to collect a stack trace.
	if errors.Is(err, errNoTCPHandler) {
		return nil, errNoTCPHandler
	}
	return appInfo, trace.Wrap(err)
}

// ReissueAppCert issues a new cert for the target app. Signatures made with the
// returned [tls.Certificate] happen over gRPC as the key never leaves the
// client application process.
func (p *remoteAppProvider) ReissueAppCert(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16) (tls.Certificate, error) {
	cert, err := p.clt.ReissueAppCert(ctx, appInfo, targetPort)
	if err != nil {
		return tls.Certificate{}, trace.Wrap(err, "reissuing certificate for app %s", appInfo.GetAppKey().GetName())
	}
	x509Cert, err := x509.ParseCertificate(cert)
	if err != nil {
		return tls.Certificate{}, trace.Wrap(err, "parsing x509 certificate for app")
	}
	appKey := appInfo.GetAppKey()
	signer := &remoteSigner{
		pub: x509Cert.PublicKey,
		sendRequest: func(ctx context.Context, req *vnetv1.SignRequest) ([]byte, error) {
			return p.clt.SignForApp(ctx, &vnetv1.SignForAppRequest{
				AppKey:     appKey,
				TargetPort: uint32(targetPort),
				Sign:       req,
			})
		},
	}
	return tls.Certificate{
		Certificate: [][]byte{cert},
		PrivateKey:  signer,
	}, nil
}

// OnNewConnection reports a new TCP connection to the target app.
func (p *remoteAppProvider) OnNewConnection(ctx context.Context, appKey *vnetv1.AppKey) error {
	if err := p.clt.OnNewConnection(ctx, appKey); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// OnInvalidLocalPort reports a failed connection to an invalid local port for
// the target app.
func (p *remoteAppProvider) OnInvalidLocalPort(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16) {
	if err := p.clt.OnInvalidLocalPort(ctx, appInfo, targetPort); err != nil {
		log.ErrorContext(ctx, "Could not notify client application about invalid local port",
			"error", err,
			"app_name", appInfo.GetAppKey().GetName(),
			"target_port", targetPort,
		)
	}
}
