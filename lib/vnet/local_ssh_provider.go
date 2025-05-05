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
	"errors"
	"strings"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
)

type localSSHProvider struct {
	ClientApplication
	clusterConfigCache *ClusterConfigCache
}

func newLocalSSHProvider(clientApp ClientApplication, clock clockwork.Clock) *localSSHProvider {
	return &localSSHProvider{
		ClientApplication:  clientApp,
		clusterConfigCache: NewClusterConfigCache(clock),
	}
}

func (p *localSSHProvider) ResolveSSHInfo(ctx context.Context, fqdn string) (*vnetv1.SshInfo, error) {
	profileNames, err := p.ClientApplication.ListProfiles()
	if err != nil {
		return nil, trace.Wrap(err, "listing profiles")
	}
	for _, profileName := range profileNames {
		clusterClient, err := p.clusterClientForFQDN(ctx, profileName, fqdn)
		if err != nil {
			if errors.Is(err, errNoMatch) {
				continue
			}
			// The user might be logged out from this one cluster (and retryWithRelogin isn't working). Log
			// the error but don't return it so that DNS resolution will be forwarded upstream instead of
			// failing.
			log.ErrorContext(ctx, "Failed to get teleport client.", "error", err)
			continue
		}

		leafClusterName := ""
		clusterName := clusterClient.ClusterName()
		if clusterName != "" && clusterName != clusterClient.RootClusterName() {
			leafClusterName = clusterName
		}

		return p.resolveSSHInfoForCluster(ctx, clusterClient, profileName, leafClusterName, fqdn)
	}
	return nil, errNoTCPHandler
}

func (p *localSSHProvider) clusterClientForFQDN(ctx context.Context, profileName, fqdn string) (ClusterClient, error) {
	rootClient, err := p.ClientApplication.GetCachedClient(ctx, profileName, "")
	if err != nil {
		log.ErrorContext(ctx, "Failed to get root cluster client, ssh nodes in this cluster will not be resolved.", "profile", profileName, "error", err)
		return nil, errNoMatch
	}
	rootClusterName := rootClient.ClusterName()

	if isDescendantSubdomain(fqdn, rootClusterName) {
		return rootClient, nil
	}

	leafClusters, err := getLeafClusters(ctx, rootClient)
	if err != nil {
		// Good chance we're here because the user is not logged in to the profile.
		log.ErrorContext(ctx, "Failed to list leaf clusters, ssh nodes in this cluster will not be resolved.", "profile", profileName, "error", err)
		return nil, errNoMatch
	}

	// Prefix with an empty string to represent the root cluster.
	allClusters := append([]string{""}, leafClusters...)
	for _, leafClusterName := range allClusters {
		if !isDescendantSubdomain(fqdn, leafClusterName+"."+rootClusterName) {
			continue
		}
		clusterClient, err := p.ClientApplication.GetCachedClient(ctx, profileName, leafClusterName)
		return clusterClient, trace.Wrap(err)
	}
	return nil, errNoMatch
}

func (p *localSSHProvider) resolveSSHInfoForCluster(
	ctx context.Context,
	clusterClient ClusterClient,
	profileName, leafClusterName, fqdn string,
) (*vnetv1.SshInfo, error) {
	target := stripSSHSuffix(fqdn, leafClusterName, clusterClient.RootClusterName())
	log := log.With("profile", profileName, "leaf_cluster", leafClusterName, "fqdn", fqdn, "target", target)
	log.DebugContext(ctx, "Resolving SSH info")
	clusterConfig, err := p.clusterConfigCache.GetClusterConfig(ctx, clusterClient)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get cluster VNet config for matching SSH node", "error", err)
		return nil, trace.Wrap(err, "getting cached cluster VNet config for matching SSH node")
	}
	dialOpts, err := p.ClientApplication.GetDialOptions(ctx, profileName)
	if err != nil {
		log.ErrorContext(ctx, "Failed to get cluster dial options", "error", err)
		return nil, trace.Wrap(err, "getting dial options for matching SSH node")
	}
	return &vnetv1.SshInfo{
		SshKey: &vnetv1.SshKey{
			Profile:     profileName,
			LeafCluster: leafClusterName,
			Hostname:    target,
		},
		Cluster:       clusterClient.ClusterName(),
		Ipv4CidrRange: clusterConfig.IPv4CIDRRange,
		DialOptions:   dialOpts,
	}, nil
}

func stripSSHSuffix(s, leafClusterName, rootClusterName string) string {
	stripped := strings.TrimSuffix(s, fullyQualify(rootClusterName))
	stripped = strings.TrimSuffix(stripped, fullyQualify(leafClusterName))
	return strings.TrimSuffix(stripped, ".")
}
