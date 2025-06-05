/*
 * Teleport
 * Copyright (C) 2025  Gravitational, Inc.
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

package services

import (
	"context"

	processhealthv1 "github.com/gravitational/teleport/api/gen/proto/go/teleport/processhealth/v1"
)

// ProcessHealth is the interface for managing user tasks resources.
type ProcessHealth interface {
	// UpsertProcessHealth creates or updates the user tasks resource.
	UpsertProcessHealth(context.Context, *processhealthv1.ProcessHealth) (*processhealthv1.ProcessHealth, error)
	// GetProcessHealth returns the user tasks resource by name.
	GetProcessHealth(ctx context.Context, name string) (*processhealthv1.ProcessHealth, error)
	// ListProcessHealth returns the user tasks resources.
	ListProcessHealths(ctx context.Context, pageSize int64, nextToken string) ([]*processhealthv1.ProcessHealth, string, error)
}

// MarshalProcessHealth marshals the ProcessHealth object into a JSON byte array.
func MarshalProcessHealth(object *processhealthv1.ProcessHealth, opts ...MarshalOption) ([]byte, error) {
	return MarshalProtoResource(object, opts...)
}

// UnmarshalProcessHealth unmarshals the ProcessHealth object from a JSON byte array.
func UnmarshalProcessHealth(data []byte, opts ...MarshalOption) (*processhealthv1.ProcessHealth, error) {
	return UnmarshalProtoResource[*processhealthv1.ProcessHealth](data, opts...)
}
