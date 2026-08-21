package lsp

import (
	"bufio"
	"context"
	"errors"
	"io"
)

func (server *Server) ServeContext(ctx context.Context, input io.Reader, output io.Writer) error {
	reader := bufio.NewReader(input)
	loop := newRequestLoop(server, output)
	for {
		if err := ctx.Err(); err != nil {
			loop.cancelAll()
			return err
		}
		if err := loop.drain(); err != nil {
			loop.cancelAll()
			return err
		}
		payload, err := readFrameContext(ctx, reader, input)
		if errors.Is(err, io.EOF) {
			loop.cancelAll()
			return loop.wait(ctx)
		}
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				loop.cancelAll()
				return ctxErr
			}
			loop.cancelAll()
			return err
		}
		if request, valid := decodeRequest(payload); valid {
			if request.Method == "$/cancelRequest" {
				server.cancelRequest(request)
				continue
			}
			if server.canRunAsync(request) {
				if err := loop.start(ctx, request, payload); err != nil {
					return err
				}
				continue
			}
		}
		response, notifications, dispatchErr := server.dispatch(ctx, payload)
		if dispatchErr != nil {
			loop.cancelAll()
			return dispatchErr
		}
		if response != nil {
			if err := writeResponse(output, response); err != nil {
				return err
			}
		}
		for _, notification := range notifications {
			if err := WriteMessage(output, notification); err != nil {
				return err
			}
		}
		shutdown, exited := server.lifecycleState()
		if exited {
			loop.cancelAll()
			if err := loop.wait(ctx); err != nil {
				return err
			}
			if shutdown {
				return nil
			}
			return ErrExitWithoutShutdown
		}
	}
}
