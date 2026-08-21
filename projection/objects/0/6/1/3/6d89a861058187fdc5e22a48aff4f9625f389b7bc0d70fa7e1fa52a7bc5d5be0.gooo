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
	uri      string
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
	loop.server.inflight[key] = &inFlightRequest{cancel: cancel, uri: requestDocumentURI(request)}
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
