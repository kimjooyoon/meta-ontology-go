package lsp

import (
	"bufio"
	"context"
	"errors"
	"io"
)

const maxMessageSize = 16 << 20

var (
	ErrMalformedHeader = errors.New("lsp: malformed header")
	ErrMissingLength   = errors.New("lsp: missing Content-Length header")
	ErrMessageTooLarge = errors.New("lsp: message exceeds maximum size")
)

// ReadMessage reads one Content-Length-framed JSON-RPC payload.
func ReadMessage(input io.Reader) ([]byte, error) {
	return readFrame(bufio.NewReader(singleByteReader{input: input}))
}

type frameResult struct {
	payload []byte
	err     error
}

func readFrameContext(ctx context.Context, reader *bufio.Reader, input io.Reader) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	results := make(chan frameResult, 1)
	go func() {
		payload, err := readFrame(reader)
		results <- frameResult{payload: payload, err: err}
	}()
	select {
	case result := <-results:
		return result.payload, result.err
	case <-ctx.Done():
		select {
		case result := <-results:
			return result.payload, result.err
		default:
		}
		if closer, interruptible := input.(io.Closer); interruptible {
			_ = closer.Close()
		}
		return nil, ctx.Err()
	}
}

type singleByteReader struct{ input io.Reader }

func (reader singleByteReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	return reader.input.Read(buffer[:1])
}
