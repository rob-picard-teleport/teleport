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
	"crypto"
	"crypto/tls"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
	"github.com/gravitational/teleport/lib/auth/authclient"
	"github.com/gravitational/teleport/lib/client"
)

// ClientApplication is the common interface implemented by each VNet client
// application: Connect and tsh. It provides methods to list user profiles, get
// cluster clients, issue app certificates, and report metrics and errors -
// anything that uses the user's credentials or a Teleport client.
// The name "client application" refers to a user-facing client application, in
// constrast to the MacOS daemon or Windows service.
type ClientApplication interface {
	// ListProfiles lists the names of all profiles saved for the user.
	ListProfiles() ([]string, error)

	// GetCachedClient returns a [*client.ClusterClient] for the given profile
	// and leaf cluster. [leafClusterName] may be empty when requesting a client
	// for the root cluster. Returned clients are expected to be cached, as this
	// may be called frequently.
	GetCachedClient(ctx context.Context, profileName, leafClusterName string) (ClusterClient, error)

	// ReissueAppCert issues a new cert for the target app.
	ReissueAppCert(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16) (tls.Certificate, error)

	// GetDialOptions returns ALPN dial options for the profile.
	GetDialOptions(ctx context.Context, profileName string) (*vnetv1.DialOptions, error)

	// OnNewConnection gets called whenever a new connection is about to be
	// established through VNet. By the time OnNewConnection is called, VNet has
	// already verified that the user holds a valid cert for the app.
	//
	// The connection won't be established until OnNewConnection returns.
	// Returning an error prevents the connection from being made.
	OnNewConnection(ctx context.Context, appKey *vnetv1.AppKey) error

	// OnInvalidLocalPort gets called before VNet refuses to handle a connection
	// to a multi-port TCP app because the provided port does not match any of
	// the TCP ports in the app spec.
	OnInvalidLocalPort(ctx context.Context, appInfo *vnetv1.AppInfo, targetPort uint16)

	TeleportClientTLSConfig(ctx context.Context, profileName, clusterName string) (*tls.Config, error)
	SessionSSHCert(ctx context.Context, sshInfo *vnetv1.SshInfo, username string) ([]byte, crypto.Signer, error)
}

// ClusterClient is an interface defining the subset of [client.ClusterClient]
// methods used by via [ClientApplication].
type ClusterClient interface {
	CurrentCluster() authclient.ClientI
	ClusterName() string
	RootClusterName() string
	SessionSSHCert(ctx context.Context, user string, target client.NodeDetails) ([]byte, crypto.Signer, error)
}
