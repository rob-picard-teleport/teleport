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

package cache

import (
	"context"

	"github.com/gravitational/trace"
	"google.golang.org/protobuf/proto"

	headerv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/header/v1"
	processhealthv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/processhealth/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/services"
)

type processHealthIndex string

const processHealthNameIndex processHealthIndex = "name"

func newProcessHealthCollection(upstream services.ProcessHealth, w types.WatchKind) (*collection[*processhealthv1.ProcessHealth, processHealthIndex], error) {
	if upstream == nil {
		return nil, trace.BadParameter("missing parameter ProcessHealths")
	}

	return &collection[*processhealthv1.ProcessHealth, processHealthIndex]{
		store: newStore(
			proto.CloneOf[*processhealthv1.ProcessHealth],
			map[processHealthIndex]func(*processhealthv1.ProcessHealth) string{
				processHealthNameIndex: func(r *processhealthv1.ProcessHealth) string {
					return r.GetMetadata().GetName()
				},
			}),
		fetcher: func(ctx context.Context, loadSecrets bool) ([]*processhealthv1.ProcessHealth, error) {
			var resources []*processhealthv1.ProcessHealth
			var nextToken string
			for {
				var page []*processhealthv1.ProcessHealth
				var err error
				page, nextToken, err = upstream.ListProcessHealths(ctx, 0 /* page size */, nextToken)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				resources = append(resources, page...)

				if nextToken == "" {
					break
				}
			}
			return resources, nil
		},
		headerTransform: func(hdr *types.ResourceHeader) *processhealthv1.ProcessHealth {
			return &processhealthv1.ProcessHealth{
				Kind:    hdr.Kind,
				Version: hdr.Version,
				Metadata: &headerv1.Metadata{
					Name: hdr.Metadata.Name,
				},
			}
		},
		watch: w,
	}, nil
}

// ListProcessHealth returns a list of ProcessHealth resources.
func (c *Cache) ListProcessHealths(ctx context.Context, pageSize int64, pageToken string) ([]*processhealthv1.ProcessHealth, string, error) {
	ctx, span := c.Tracer.Start(ctx, "cache/ListProcessHealths")
	defer span.End()

	lister := genericLister[*processhealthv1.ProcessHealth, processHealthIndex]{
		cache:      c,
		collection: c.collections.processHealth,
		index:      processHealthNameIndex,
		upstreamList: func(ctx context.Context, i int, s string) ([]*processhealthv1.ProcessHealth, string, error) {
			out, next, err := c.Config.ProcessHealth.ListProcessHealths(ctx, pageSize, pageToken)
			return out, next, trace.Wrap(err)
		},
		nextToken: func(t *processhealthv1.ProcessHealth) string {
			return t.GetMetadata().Name
		},
	}
	out, next, err := lister.list(ctx, int(pageSize), pageToken)
	return out, next, trace.Wrap(err)
}

// GetProcessHealth returns the specified ProcessHealth resource.
func (c *Cache) GetProcessHealth(ctx context.Context, name string) (*processhealthv1.ProcessHealth, error) {
	ctx, span := c.Tracer.Start(ctx, "cache/GetProcessHealth")
	defer span.End()

	getter := genericGetter[*processhealthv1.ProcessHealth, processHealthIndex]{
		cache:       c,
		collection:  c.collections.processHealth,
		index:       processHealthNameIndex,
		upstreamGet: c.Config.ProcessHealth.GetProcessHealth,
	}
	out, err := getter.get(ctx, name)
	return out, trace.Wrap(err)
}
