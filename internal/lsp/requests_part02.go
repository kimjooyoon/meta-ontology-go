package lsp

import (
	"context"
)

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
func (server *Server) cancelRequestsForURI(uri string) {
	for _, state := range server.inflight {
		if state.uri != uri {
			continue
		}
		state.canceled = true
		state.cancel()
	}
}
