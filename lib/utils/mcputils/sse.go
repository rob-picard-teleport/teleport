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

package mcputils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gravitational/trace"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gravitational/teleport/lib/defaults"
)

// SSERequestWriter posts requests to the remote server.
type SSERequestWriter struct {
	client      *http.Client
	endpointURL *url.URL
}

// NewSSERequestWriter creates a new SSERequestWriter.
func NewSSERequestWriter(client *http.Client, endpointURL *url.URL) *SSERequestWriter {
	return &SSERequestWriter{
		client:      client,
		endpointURL: endpointURL,
	}
}

// GetSessionID returns the session ID tracked by the remote server.
func (w *SSERequestWriter) GetSessionID() string {
	return w.endpointURL.Query().Get("sessionId")
}

// WriteMessage posts the request to the remote server.
func (w *SSERequestWriter) WriteMessage(ctx context.Context, msg mcp.JSONRPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return trace.Wrap(err, "marshalling message")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.endpointURL.String(), bytes.NewReader(data))
	if err != nil {
		return trace.Wrap(err, "building SSE POST request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return trace.Wrap(err, "sending SSE request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return trace.BadParameter("SSE request returned %s", resp.Status)
	}
	return nil
}

// SSEResponseReader implements TransportReader for reading SSE stream from the
// MCP server.
type SSEResponseReader struct {
	io.Closer
	br *bufio.Reader
}

// NewSSEResponseReader creates a new SSEResponseReader. Input reader is usually the
// http body used for SSE stream.
func NewSSEResponseReader(reader io.ReadCloser) *SSEResponseReader {
	return &SSEResponseReader{
		Closer: reader,
		br:     bufio.NewReader(reader),
	}
}

// ReadEndpoint reads the endpoint event and crafts the endpoint URL.
// This is usually the first event after connecting to SSE server.
func (r *SSEResponseReader) ReadEndpoint(ctx context.Context, baseURL *url.URL) (*url.URL, error) {
	event, err := readSSEEvent(ctx, r.br)
	if err != nil {
		return nil, trace.Wrap(err, "reading SSE server message")
	}
	if event.EventType != SSEEventEndpoint {
		return nil, trace.BadParameter("expecting endpoint event, got %s", event.EventType)
	}

	endpointURI, err := baseURL.Parse(event.Data)
	if err != nil {
		return nil, trace.Wrap(err, "parsing endpoint data")
	}
	return endpointURI, nil
}

// ReadMessage reads the next SSE message event from SSE stream.
func (r *SSEResponseReader) ReadMessage(ctx context.Context) (string, error) {
	event, err := readSSEEvent(ctx, r.br)
	if err != nil {
		return "", trace.Wrap(err)
	}
	if event.EventType != SSEEventMessage {
		return "", newReaderParseError(trace.BadParameter("unexpected event type %s", event.EventType))
	}
	return event.Data, nil
}

// Type returns "sse".
func (r *SSEResponseReader) Type() string {
	return "sse"
}

func ConnectSSEServer(ctx context.Context, baseURL *url.URL) (*SSEResponseReader, *SSERequestWriter, error) {
	client, err := defaults.HTTPClient()
	if err != nil {
		return nil, nil, trace.Wrap(err, "making HTTP client")
	}

	connectReq, err := makeSSEConnectionRequest(ctx, baseURL.String())
	if err != nil {
		return nil, nil, trace.Wrap(err, "making SSE connection request")
	}

	resp, err := client.Do(connectReq)
	if err != nil {
		return nil, nil, trace.Wrap(err, "sending SSE request")
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, nil, trace.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader := NewSSEResponseReader(resp.Body)
	endpointURL, err := reader.ReadEndpoint(ctx, baseURL)
	if err != nil {
		defer reader.Close()
		return nil, nil, trace.Wrap(err, "reading SSE server endpoint")
	}
	requestWriter := NewSSERequestWriter(client, endpointURL)
	return reader, requestWriter, nil
}

func makeSSEConnectionRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, trace.Wrap(err, "building SSE request")
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	return req, nil
}

type SSEEventType string

const (
	SSEEventEndpoint SSEEventType = "endpoint"
	SSEEventMessage  SSEEventType = "message"
)

func toSSEEventType(event string) (SSEEventType, error) {
	switch event {
	case string(SSEEventEndpoint):
		return SSEEventEndpoint, nil
	case "", string(SSEEventMessage):
		return SSEEventMessage, nil
	default:
		return "", trace.BadParameter("unknown SSE event: %s", event)
	}
}

type SSEEvent struct {
	EventType SSEEventType
	Data      string
}

func readSSEEvent(ctx context.Context, br *bufio.Reader) (*SSEEvent, error) {
	var event, data string
	for {
		if ctx.Err() != nil {
			return nil, trace.Wrap(ctx.Err())
		}

		line, err := br.ReadString('\n')
		if err != nil {
			return nil, trace.Wrap(err)
		}

		// Remove only newline markers
		line = strings.TrimRight(line, "\r\n")

		// Empty line means end of event
		if line == "" {
			if data != "" {
				eventType, err := toSSEEventType(event)
				if err != nil {
					return nil, newReaderParseError(err)
				}
				return &SSEEvent{EventType: eventType, Data: data}, nil
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
}
