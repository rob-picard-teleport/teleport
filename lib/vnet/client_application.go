package vnet

import (
	"context"
	"crypto/tls"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
	"github.com/gravitational/teleport/lib/auth/authclient"
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
}

// ClusterClient is an interface defining the subset of [client.ClusterClient]
// methods used by via [ClientApplication].
type ClusterClient interface {
	CurrentCluster() authclient.ClientI
	ClusterName() string
	RootClusterName() string
}
