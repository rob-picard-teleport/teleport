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
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/gravitational/trace"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/gravitational/teleport"
	logutils "github.com/gravitational/teleport/lib/utils/log"
)

// StderrTraceLogWriter implements io.Writer and logs the content at TRACE
// level. Used for tracing stderr.
type StderrTraceLogWriter struct {
	ctx context.Context
	log *slog.Logger
}

// NewStderrTraceLogWriter returns a new StderrTraceLogWriter.
func NewStderrTraceLogWriter(ctx context.Context, log *slog.Logger) *StderrTraceLogWriter {
	return &StderrTraceLogWriter{
		ctx: ctx,
		log: cmp.Or(log, slog.Default()),
	}
}

// Write implements io.Writer and logs the given input p at trace level.
// Note that the input p may contain arbitrary-length data, which can span
// multiple lines or include partial lines.
func (l *StderrTraceLogWriter) Write(p []byte) (int, error) {
	l.log.Log(l.ctx, logutils.TraceLevel, "Trace stderr", "data", p)
	return len(p), nil
}

// StdioMessageWriter writes a JSONRPC message in stdio transport.
type StdioMessageWriter struct {
	w io.Writer
}

// NewStdioMessageWriter returns a MessageWriter using stdio transport.
func NewStdioMessageWriter(w io.Writer) *StdioMessageWriter {
	return &StdioMessageWriter{
		w: w,
	}
}

// WriteMessage writes a JSONRPC message in stdio transport.
func (w *StdioMessageWriter) WriteMessage(_ context.Context, resp mcp.JSONRPCMessage) error {
	bytes, err := json.Marshal(resp)
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = fmt.Fprintf(w.w, "%s\n", string(bytes))
	return trace.Wrap(err)
}

// HandleParseErrorFunc handles parse errors.
type HandleParseErrorFunc func(context.Context, *mcp.JSONRPCError) error
type HandleRequestFunc func(context.Context, *JSONRPCRequest) error
type HandleResponseFunc func(context.Context, *JSONRPCResponse) error
type HandleNotificationFunc func(context.Context, *JSONRPCNotification) error

// ReplyParseError returns a HandleParseErrorFunc that forwards the error to
// provided writer.
func ReplyParseError(w MessageWriter) HandleParseErrorFunc {
	return func(ctx context.Context, parseError *mcp.JSONRPCError) error {
		return trace.Wrap(w.WriteMessage(ctx, parseError))
	}
}

// LogAndIgnoreParseError returns a HandleParseErrorFunc that logs the parse
// error.
func LogAndIgnoreParseError(log *slog.Logger) HandleParseErrorFunc {
	return func(ctx context.Context, parseError *mcp.JSONRPCError) error {
		log.DebugContext(ctx, "Ignore parse error", "error", parseError)
		return nil
	}
}

// TransportReader defines an interface for reading next raw/unmarshalled
// message from the MCP transport.
type TransportReader interface {
	// Type is the transport type for logging purpose.
	Type() string
	// ReadMessage reads the next raw message.
	ReadMessage(context.Context) (string, error)
	// Close closes the transport.
	Close() error
}

// StdioReader implements TransportReader for stdio transport
type StdioReader struct {
	io.Closer
	br *bufio.Reader
}

// NewStdioReader creates a new StdioReader. Input reader can be either stdin or
// stdout.
func NewStdioReader(reader io.ReadCloser) *StdioReader {
	return &StdioReader{
		Closer: reader,
		br:     bufio.NewReader(reader),
	}
}

// ReadMessage reads the next line.
func (r *StdioReader) ReadMessage(context.Context) (string, error) {
	line, err := r.br.ReadString('\n')
	if err != nil {
		return "", trace.Wrap(err)
	}
	return line, nil
}

// Type returns "stdio".
func (r *StdioReader) Type() string {
	return "stdio"
}

// MessageReaderConfig is the config for MessageReader.
type MessageReaderConfig struct {
	// Transport is the input to the read the message from. Transport will be
	// closed when reader finishes.
	Transport TransportReader
	// Logger is the slog.Logger.
	Logger *slog.Logger
	// ParentContext is the parent's context. Used for logging during tear down.
	ParentContext context.Context

	// OnClose is an optional callback when reader finishes.
	OnClose func()
	// OnParseError specifies the handler for handling parse error. Any error
	// returned by the handler stops this message reader.
	OnParseError HandleParseErrorFunc
	// OnRequest specifies the handler for handling request. Any error by the
	// handler stops this message reader.
	OnRequest HandleRequestFunc
	// OnResponse specifies the handler for handling response. Any error by the
	// handler stops this message reader.
	OnResponse HandleResponseFunc
	// OnNotification specifies the handler for handling notification. Any error
	// returned by the handler stops this message reader.
	OnNotification HandleNotificationFunc
}

// CheckAndSetDefaults checks values and sets defaults.
func (c *MessageReaderConfig) CheckAndSetDefaults() error {
	if c.Transport == nil {
		return trace.BadParameter("missing parameter Transport")
	}
	if c.OnParseError == nil {
		return trace.BadParameter("missing parameter OnParseError")
	}
	if c.OnNotification == nil {
		return trace.BadParameter("missing parameter OnNotification")
	}
	if c.OnRequest == nil && c.OnResponse == nil {
		return trace.BadParameter("one of OnRequest or OnResponse must be set")
	}
	if c.ParentContext == nil {
		return trace.BadParameter("missing parameter ParentContext")
	}
	if c.Logger == nil {
		c.Logger = slog.With(teleport.ComponentKey, "mcp")
	}
	return nil
}

// MessageReader reads messages with provided transport and config.
type MessageReader struct {
	cfg MessageReaderConfig
}

// NewMessageReader creates a new MessageReader. Must call "Start" to
// start the processing.
func NewMessageReader(cfg MessageReaderConfig) (*MessageReader, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return &MessageReader{
		cfg: cfg,
	}, nil
}

// Run starts reading requests from provided reader. Run blocks until an
// error happens from the provided reader or any of the handler.
func (r *MessageReader) Run(ctx context.Context) {
	r.cfg.Logger.InfoContext(ctx, "Start processing messages", "transport", r.cfg.Transport.Type())

	finished := make(chan struct{})
	go func() {
		r.startProcess(ctx)
		close(finished)
	}()

	select {
	case <-finished:
	case <-ctx.Done():
	}

	r.cfg.Logger.InfoContext(r.cfg.ParentContext, "Finished processing messages", "transport", r.cfg.Transport.Type())
	if err := r.cfg.Transport.Close(); err != nil && !IsOKCloseError(err) {
		r.cfg.Logger.ErrorContext(r.cfg.ParentContext, "Failed to close transport", "error", err)
	}
	if r.cfg.OnClose != nil {
		r.cfg.OnClose()
	}
}

func (r *MessageReader) startProcess(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		if err := r.processNextMessage(ctx); err != nil {
			if !IsOKCloseError(err) {
				r.cfg.Logger.ErrorContext(ctx, "Failed to process line", "error", err)
			}
			return
		}
	}
}

func (r *MessageReader) processNextMessage(ctx context.Context) error {
	rawMessage, err := r.cfg.Transport.ReadMessage(ctx)
	switch {
	case isReaderParseError(err):
		rpcError := mcp.NewJSONRPCError(mcp.NewRequestId(nil), mcp.PARSE_ERROR, err.Error(), nil)
		if err := r.cfg.OnParseError(ctx, &rpcError); err != nil {
			return trace.Wrap(err, "handling reader parse error")
		}
	case err != nil:
		return trace.Wrap(err, "reading next data")
	}

	r.cfg.Logger.Log(ctx, logutils.TraceLevel, "Trace read", "raw", rawMessage)

	var base baseJSONRPCMessage
	if parseError := json.Unmarshal([]byte(rawMessage), &base); parseError != nil {
		rpcError := mcp.NewJSONRPCError(mcp.NewRequestId(nil), mcp.PARSE_ERROR, parseError.Error(), nil)
		if err := r.cfg.OnParseError(ctx, &rpcError); err != nil {
			return trace.Wrap(err, "handling JSON unmarshal error")
		}
	}

	switch {
	case base.isNotification():
		return trace.Wrap(r.cfg.OnNotification(ctx, base.makeNotification()), "handling notification")
	case base.isRequest():
		if r.cfg.OnRequest != nil {
			return trace.Wrap(r.cfg.OnRequest(ctx, base.makeRequest()), "handling request")
		}
		// Should not happen. Log something just in case.
		r.cfg.Logger.DebugContext(ctx, "Skipping request", "id", base.ID)
		return nil
	case base.isResponse():
		if r.cfg.OnResponse != nil {
			return trace.Wrap(r.cfg.OnResponse(ctx, base.makeResponse()), "handling response")
		}
		// Should not happen. Log something just in case.
		r.cfg.Logger.DebugContext(ctx, "Skipping response", "id", base.ID)
		return nil
	default:
		rpcError := mcp.NewJSONRPCError(base.ID, mcp.PARSE_ERROR, "unknown message type", rawMessage)
		return trace.Wrap(
			r.cfg.OnParseError(ctx, &rpcError),
			"handling unknown message type error",
		)
	}
}
