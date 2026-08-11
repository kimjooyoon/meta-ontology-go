package lsp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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

type singleByteReader struct{ input io.Reader }

func (reader singleByteReader) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}
	return reader.input.Read(buffer[:1])
}

func readFrame(reader *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) && len(line) == 0 {
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" {
			if length < 0 {
				return nil, ErrMissingLength
			}
			break
		}
		name, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(name) == "" {
			return nil, ErrMalformedHeader
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(value))
			if parseErr != nil || parsed < 0 {
				return nil, ErrMalformedHeader
			}
			length = parsed
		}
	}
	if length > maxMessageSize {
		return nil, ErrMessageTooLarge
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func WriteMessage(output io.Writer, payload []byte) error {
	if len(payload) > maxMessageSize {
		return ErrMessageTooLarge
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	if err := writeAll(output, []byte(header)); err != nil {
		return err
	}
	return writeAll(output, payload)
}

func writeAll(output io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := output.Write(data)
		if written > 0 {
			data = data[written:]
		}
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
