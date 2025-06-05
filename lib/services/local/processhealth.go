/*
 * Teleport
 * Copyright (C) 2024  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package local

import (
	"context"

	"github.com/gravitational/trace"

	processhealthv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/processhealth/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/backend"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local/generic"
)

type ProcessHealthService struct {
	service *generic.ServiceWrapper[*processhealthv1.ProcessHealth]
}

const processHealthKey = "process_health"

// NewProcessHealthService creates a new ProcessHealthService.
func NewProcessHealthService(b backend.Backend) (*ProcessHealthService, error) {
	service, err := generic.NewServiceWrapper(
		generic.ServiceConfig[*processhealthv1.ProcessHealth]{
			Backend:       b,
			ResourceKind:  types.KindProcessHealth,
			BackendPrefix: backend.NewKey(processHealthKey),
			MarshalFunc:   services.MarshalProtoResource[*processhealthv1.ProcessHealth],
			UnmarshalFunc: services.UnmarshalProtoResource[*processhealthv1.ProcessHealth],
		})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &ProcessHealthService{service: service}, nil
}

func (s *ProcessHealthService) ListProcessHealths(ctx context.Context, pagesize int64, lastKey string) ([]*processhealthv1.ProcessHealth, string, error) {
	r, nextToken, err := s.service.ListResources(ctx, int(pagesize), lastKey)
	return r, nextToken, trace.Wrap(err)
}

func (s *ProcessHealthService) GetProcessHealth(ctx context.Context, name string) (*processhealthv1.ProcessHealth, error) {
	r, err := s.service.GetResource(ctx, name)
	return r, trace.Wrap(err)
}

func (s *ProcessHealthService) UpsertProcessHealth(ctx context.Context, userTask *processhealthv1.ProcessHealth) (*processhealthv1.ProcessHealth, error) {
	r, err := s.service.UpsertResource(ctx, userTask)
	return r, trace.Wrap(err)
}
