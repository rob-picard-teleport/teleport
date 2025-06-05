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

package processhealthv1

import (
	"context"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"

	processhealthv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/processhealth/v1"
	"github.com/gravitational/teleport/api/types"
	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/services"
	usagereporter "github.com/gravitational/teleport/lib/usagereporter/teleport"
)

// ServiceConfig holds configuration options for the ProcessHealth gRPC service.
type ServiceConfig struct {
	// Authorizer is the authorizer to use.
	Authorizer authz.Authorizer

	// Backend is the backend for storing ProcessHealth.
	Backend services.ProcessHealth

	// Cache is the cache for storing ProcessHealth.
	Cache Reader

	// Clock is used to control time - mainly used for testing.
	Clock clockwork.Clock

	// UsageReporter is the reporter for sending usage without it be related to an API call.
	UsageReporter func() usagereporter.UsageReporter

	// Emitter is the event emitter.
	Emitter apievents.Emitter
}

// CheckAndSetDefaults checks the ServiceConfig fields and returns an error if
// a required param is not provided.
// Authorizer, Cache and Backend are required params
func (s *ServiceConfig) CheckAndSetDefaults() error {
	if s.Authorizer == nil {
		return trace.BadParameter("authorizer is required")
	}
	if s.Backend == nil {
		return trace.BadParameter("backend is required")
	}
	if s.Cache == nil {
		return trace.BadParameter("cache is required")
	}
	if s.UsageReporter == nil {
		return trace.BadParameter("usage reporter is required")
	}
	if s.Emitter == nil {
		return trace.BadParameter("emitter is required")
	}
	if s.Clock == nil {
		s.Clock = clockwork.NewRealClock()
	}

	return nil
}

// Reader contains the methods defined for cache access.
type Reader interface {
	ListProcessHealths(ctx context.Context, pageSize int64, nextToken string) ([]*processhealthv1.ProcessHealth, string, error)
	GetProcessHealth(ctx context.Context, name string) (*processhealthv1.ProcessHealth, error)
}

// Service implements the teleport.ProcessHealth.v1.ProcessHealthService RPC service.
type Service struct {
	processhealthv1.UnimplementedProcessHealthServiceServer

	authorizer    authz.Authorizer
	backend       services.ProcessHealth
	cache         Reader
	clock         clockwork.Clock
	usageReporter func() usagereporter.UsageReporter
	emitter       apievents.Emitter
}

// NewService returns a new ProcessHealth gRPC service.
func NewService(cfg ServiceConfig) (*Service, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	return &Service{
		authorizer:    cfg.Authorizer,
		backend:       cfg.Backend,
		cache:         cfg.Cache,
		clock:         cfg.Clock,
		usageReporter: cfg.UsageReporter,
		emitter:       cfg.Emitter,
	}, nil
}

// ListProcessHealths returns a list of user tasks.
func (s *Service) ListProcessHealths(ctx context.Context, req *processhealthv1.ListProcessHealthsRequest) (*processhealthv1.ListProcessHealthsResponse, error) {
	authCtx, err := s.authorizer.Authorize(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := authCtx.CheckAccessToKind(types.KindProcessHealth, types.VerbRead, types.VerbList); err != nil {
		return nil, trace.Wrap(err)
	}

	rsp, nextToken, err := s.cache.ListProcessHealths(ctx, int64(req.PageSize), req.PageToken)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &processhealthv1.ListProcessHealthsResponse{
		ProcessHealths: rsp,
		NextPageToken:  nextToken,
	}, nil
}

// GetProcessHealth returns user task resource.
func (s *Service) GetProcessHealth(ctx context.Context, req *processhealthv1.GetProcessHealthRequest) (*processhealthv1.ProcessHealth, error) {
	authCtx, err := s.authorizer.Authorize(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := authCtx.CheckAccessToKind(types.KindProcessHealth, types.VerbRead); err != nil {
		return nil, trace.Wrap(err)
	}

	rsp, err := s.cache.GetProcessHealth(ctx, req.GetName())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return rsp, nil

}

// UpsertProcessHealth upserts user task resource.
func (s *Service) UpsertProcessHealth(ctx context.Context, req *processhealthv1.UpsertProcessHealthRequest) (*processhealthv1.ProcessHealth, error) {
	authCtx, err := s.authorizer.Authorize(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := authCtx.CheckAccessToKind(types.KindProcessHealth, types.VerbUpdate, types.VerbCreate); err != nil {
		return nil, trace.Wrap(err)
	}

	rsp, err := s.backend.UpsertProcessHealth(ctx, req.ProcessHealth)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return rsp, nil
}
