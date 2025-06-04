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

package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gravitational/teleport"
	apidefaults "github.com/gravitational/teleport/api/defaults"
	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/utils/mcputils"
)

func shouldEmitEvent(method mcp.MCPMethod) bool {
	switch method {
	case mcp.MethodPing,
		mcp.MethodResourcesList,
		mcp.MethodResourcesTemplatesList,
		mcp.MethodPromptsList,
		mcp.MethodToolsList:
		return false
	default:
		return true
	}
}

type auditor struct {
	// cfg reuses a valid ServerConfig (because i'm lazy).
	cfg ServerConfig
}

func (a *auditor) emitStartEvent(ctx context.Context, session *SessionCtx) {
	a.emitEvent(ctx, &apievents.MCPSessionStart{
		Metadata: apievents.Metadata{
			Type:        events.MCPSessionStartEvent,
			Code:        events.MCPSessionStartCode,
			ClusterName: session.Identity.RouteToApp.ClusterName,
		},
		ServerMetadata:     a.makeSessionServerMetadata(),
		SessionMetadata:    a.makeSessionMetadata(session),
		UserMetadata:       session.Identity.GetUserMetadata(),
		ConnectionMetadata: a.makeSessionConnectionMetadata(session),
		AppMetadata:        a.makeSessionAppMetadata(session),
	})
}

func (a *auditor) emitEndEvent(ctx context.Context, session *SessionCtx) {
	a.emitEvent(ctx, &apievents.MCPSessionEnd{
		Metadata: apievents.Metadata{
			Type:        events.MCPSessionEndEvent,
			Code:        events.MCPSessionEndCode,
			ClusterName: session.Identity.RouteToApp.ClusterName,
		},
		ServerMetadata:     a.makeSessionServerMetadata(),
		SessionMetadata:    a.makeSessionMetadata(session),
		UserMetadata:       session.Identity.GetUserMetadata(),
		ConnectionMetadata: a.makeSessionConnectionMetadata(session),
		AppMetadata:        a.makeSessionAppMetadata(session),
	})
}

func (a *auditor) emitNotificationEvent(ctx context.Context, session *SessionCtx, msg *mcputils.JSONRPCNotification) {
	if !shouldEmitEvent(msg.Method) {
		return
	}
	a.emitEvent(ctx, &apievents.MCPSessionNotification{
		Metadata: apievents.Metadata{
			Type:        events.MCPSessionNotificationEvent,
			Code:        events.MCPSessionNotificationCode,
			ClusterName: session.Identity.RouteToApp.ClusterName,
		},
		SessionMetadata: a.makeSessionMetadata(session),
		UserMetadata:    session.Identity.GetUserMetadata(),
		AppMetadata:     a.makeSessionAppMetadata(session),
		Message: apievents.MCPJSONRPCMessage{
			JSONRPC: msg.JSONRPC,
			Method:  string(msg.Method),
			Params:  msg.Params.GetEventParams(),
		},
	})
}

func (a *auditor) emitRequestEvent(ctx context.Context, session *SessionCtx, msg *mcputils.JSONRPCRequest, err error) {
	if !shouldEmitEvent(msg.Method) && err == nil {
		return
	}
	event := &apievents.MCPSessionRequest{
		Metadata: apievents.Metadata{
			Type:        events.MCPSessionRequestEvent,
			Code:        events.MCPSessionRequestCode,
			ClusterName: session.Identity.RouteToApp.ClusterName,
		},
		SessionMetadata: a.makeSessionMetadata(session),
		UserMetadata:    session.Identity.GetUserMetadata(),
		AppMetadata:     a.makeSessionAppMetadata(session),
		Status: apievents.Status{
			Success: true,
		},
		Message: apievents.MCPJSONRPCMessage{
			JSONRPC: msg.JSONRPC,
			Method:  string(msg.Method),
			ID:      msg.ID.String(),
			Params:  msg.Params.GetEventParams(),
		},
	}

	if err != nil {
		event.Metadata.Code = events.MCPSessionRequestFailureCode
		event.Status.Success = false
		event.Status.Error = err.Error()
	}
	a.emitEvent(ctx, event)
}

func (a *auditor) emitEvent(ctx context.Context, event apievents.AuditEvent) {
	if err := a.cfg.Emitter.EmitAuditEvent(ctx, event); err != nil {
		a.cfg.Log.DebugContext(ctx, "Failed to emit audit event", "error", err)
	}
}

func (a *auditor) makeSessionServerMetadata() apievents.ServerMetadata {
	return apievents.ServerMetadata{
		ServerVersion:   teleport.Version,
		ServerID:        a.cfg.HostID,
		ServerNamespace: apidefaults.Namespace,
	}
}

func (a *auditor) makeSessionConnectionMetadata(session *SessionCtx) apievents.ConnectionMetadata {
	return apievents.ConnectionMetadata{
		RemoteAddr: session.Identity.LoginIP,
	}
}

func (a *auditor) makeSessionAppMetadata(session *SessionCtx) apievents.AppMetadata {
	return apievents.AppMetadata{
		AppURI:  session.App.GetURI(),
		AppName: session.App.GetName(),
	}
}

func (a *auditor) makeSessionMetadata(session *SessionCtx) apievents.SessionMetadata {
	return apievents.SessionMetadata{
		SessionID:        session.sessionID,
		WithMFA:          session.Identity.MFAVerified,
		PrivateKeyPolicy: string(session.Identity.PrivateKeyPolicy),
	}
}
