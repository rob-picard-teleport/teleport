// Copyright 2025 Gravitational, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processhealth

import (
	"context"

	"github.com/gravitational/trace"

	processhealthv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/processhealth/v1"
)

// Client is a client for the Process Health API.
type Client struct {
	grpcClient processhealthv1.ProcessHealthServiceClient
}

// NewClient creates a new Process Health client.
func NewClient(grpcClient processhealthv1.ProcessHealthServiceClient) *Client {
	return &Client{
		grpcClient: grpcClient,
	}
}

// ListProcessHealths returns a list of Process Healths.
func (c *Client) ListProcessHealths(ctx context.Context, pageSize int64, nextToken string) ([]*processhealthv1.ProcessHealth, string, error) {
	resp, err := c.grpcClient.ListProcessHealths(ctx, &processhealthv1.ListProcessHealthsRequest{
		PageSize:  int32(pageSize),
		PageToken: nextToken,
	})
	if err != nil {
		return nil, "", trace.Wrap(err)
	}

	return resp.ProcessHealths, resp.NextPageToken, nil
}
func (c *Client) GetProcessHealth(ctx context.Context, name string) (*processhealthv1.ProcessHealth, error) {
	rsp, err := c.grpcClient.GetProcessHealth(ctx, &processhealthv1.GetProcessHealthRequest{
		Name: name,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return rsp, nil
}

// UpsertProcessHealth upserts a Process Health.
func (c *Client) UpsertProcessHealth(ctx context.Context, req *processhealthv1.ProcessHealth) (*processhealthv1.ProcessHealth, error) {
	rsp, err := c.grpcClient.UpsertProcessHealth(ctx, &processhealthv1.UpsertProcessHealthRequest{
		ProcessHealth: req,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return rsp, nil
}
