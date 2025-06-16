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
	"net/url"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/utils"
	logutils "github.com/gravitational/teleport/lib/utils/log"
	"github.com/gravitational/teleport/lib/utils/mcputils"
)

func (s *Server) handleStdioToSSE(ctx context.Context, sessionCtx SessionCtx) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	baseURL, err := makeSSEBaseURI(sessionCtx.App)
	if err != nil {
		return trace.Wrap(err, "parsing SSE URI")
	}
	session, err := s.makeSessionHandler(ctx, &sessionCtx)
	if err != nil {
		return trace.Wrap(err)
	}

	session.logger.DebugContext(s.cfg.ParentContext, "Started handling stdio to SSE session", "base_uri", logutils.StringerAttr(baseURL))
	defer session.logger.DebugContext(s.cfg.ParentContext, "Completed handling stdio to SSE session")

	serverTransportReader, serverRequestWriter, err := mcputils.ConnectSSEServer(ctx, baseURL)
	if err != nil {
		return trace.Wrap(err)
	}

	if externalSessionID := serverRequestWriter.GetSessionID(); externalSessionID != "" {
		session.externalSessionID = externalSessionID
		session.logger.DebugContext(s.cfg.ParentContext, "Found external session ID", "session_id", externalSessionID)
	}

	clientResponseWriter := mcputils.NewStdioMessageWriter(utils.NewSyncWriter(sessionCtx.ClientConn))
	stdoutLogger := session.logger.With("sse", "stdout")
	serverResponseReader, err := mcputils.NewMessageReader(mcputils.MessageReaderConfig{
		Transport:      serverTransportReader,
		Logger:         stdoutLogger,
		ParentContext:  s.cfg.ParentContext,
		OnClose:        cancel,
		OnParseError:   mcputils.LogAndIgnoreParseError(stdoutLogger),
		OnNotification: session.onServerNotification(clientResponseWriter),
		OnResponse:     session.onServerResponse(clientResponseWriter),
	})
	if err != nil {
		return trace.Wrap(err)
	}
	go serverResponseReader.Run(ctx)

	clientRequestReader, err := makeStdioClientRequestReader(session, clientResponseWriter, serverRequestWriter, cancel)
	if err != nil {
		return trace.Wrap(err)
	}

	session.emitStartEvent(session.parentCtx)
	defer session.emitEndEvent(session.parentCtx)
	clientRequestReader.Run(ctx)
	return nil
}

func makeSSEBaseURI(app types.Application) (*url.URL, error) {
	baseURL, err := url.Parse(app.GetURI())
	if err != nil {
		return nil, trace.Wrap(err, "parsing SSE URI")
	}
	transportType := types.GetMCPServerTransportType(app.GetURI())
	switch transportType {
	case types.MCPTransportSSEHTTP:
		baseURL.Scheme = "http"
	case types.MCPTransportSSEHTTPS:
		baseURL.Scheme = "https"
	default:
		return nil, trace.BadParameter("unknown transport type: %v", transportType)
	}
	return baseURL, nil
}
