package lsp

import (
	"context"
	"encoding/json"
	"io"
)

type cancelParams struct {
	ID json.RawMessage `json:"id"`
}

type inFlightRequest struct {
	cancel   context.CancelFunc
	canceled bool
}

type requestResult struct {
	key           string
	ctx           context.Context
	response      *responseEnvelope
	notifications [][]byte
	err           error
}

type requestLoop struct {
	server  *Server
	output  io.Writer
	results chan requestResult
	pending int
}

func newRequestLoop(server *Server, output io.Writer) *requestLoop {
	return &requestLoop{server: server, output: output, results: make(chan requestResult, 16)}
}

func (loop *requestLoop) start(parent context.Context, request requestEnvelope, payload []byte) error {
	key := string(request.ID)
	if _, exists := loop.server.inflight[key]; exists {
		return writeResponse(loop.output, errorResponse(request.ID, invalidRequest, "duplicate request ID"))
	}
	requestCtx, cancel := context.WithCancel(parent)
	loop.server.inflight[key] = &inFlightRequest{cancel: cancel}
	loop.pending++
	go func() {
		response, notifications, err := loop.server.dispatch(requestCtx, payload)
		loop.results <- requestResult{key: key, ctx: requestCtx, response: response, notifications: notifications, err: err}
	}()
	return nil
}

func (loop *requestLoop) drain() error {
	for {
		select {
		case result := <-loop.results:
			loop.pending--
			if err := loop.finish(result); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (loop *requestLoop) wait(ctx context.Context) error {
	for loop.pending > 0 {
		select {
		case result := <-loop.results:
			loop.pending--
			if err := loop.finish(result); err != nil {
				return err
			}
		case <-ctx.Done():
			loop.cancelAll()
			return ctx.Err()
		}
	}
	return nil
}

func (loop *requestLoop) finish(result requestResult) error {
	state, exists := loop.server.inflight[result.key]
	delete(loop.server.inflight, result.key)
	if !exists || state.canceled || result.ctx.Err() != nil {
		return nil
	}
	if result.err != nil {
		return result.err
	}
	if result.response != nil {
		if err := writeResponse(loop.output, result.response); err != nil {
			return err
		}
	}
	for _, notification := range result.notifications {
		if err := WriteMessage(loop.output, notification); err != nil {
			return err
		}
	}
	return nil
}

func (loop *requestLoop) cancelAll() {
	for _, state := range loop.server.inflight {
		state.canceled = true
		state.cancel()
	}
}

func (server *Server) cancelRequest(request requestEnvelope) {
	var params cancelParams
	if decodeParams(request.Params, &params) != nil || len(params.ID) == 0 {
		return
	}
	if state, exists := server.inflight[string(params.ID)]; exists {
		state.canceled = true
		state.cancel()
	}
}

func (server *Server) canRunAsync(request requestEnvelope) bool {
	if request.ID == nil {
		return false
	}
	if _, ok := server.parser.(ContextParser); !ok {
		return false
	}
	if _, isSyntaxParser := server.parser.(SyntaxParser); isSyntaxParser {
		return false
	}
	switch request.Method {
	case "textDocument/hover", "textDocument/completion", "textDocument/definition", "textDocument/semanticTokens/full":
		return true
	default:
		return false
	}
}

func decodeRequest(payload []byte) (requestEnvelope, bool) {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return requestEnvelope{}, false
	}
	return request, true
}
