package lsp

import (
	"bufio"
	"bytes"
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

// ReadMessage reads one LSP payload, including its Content-Length framing.
func ReadMessage(input io.Reader) ([]byte, error) {
	// A one-byte buffer avoids consuming the next frame when callers invoke
	// this convenience function repeatedly on the same stream. bufio enforces
	// a larger minimum, so constrain the underlying reader instead.
	return readFrame(bufio.NewReader(singleByteReader{input: input}))
}

type singleByteReader struct {
	input io.Reader
}

func (r singleByteReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return r.input.Read(p[:1])
}

func readFrame(input *bufio.Reader) ([]byte, error) {
	length, err := readHeaders(input)
	if err != nil {
		return nil, err
	}
	if length > maxMessageSize {
		return nil, ErrMessageTooLarge
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(input, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func readHeaders(input *bufio.Reader) (int, error) {
	length := -1
	for {
		line, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) == 0 && length < 0 {
				return 0, io.EOF
			}
			return 0, io.ErrUnexpectedEOF
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" {
			if length < 0 {
				return 0, ErrMissingLength
			}
			return length, nil
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(name) == "" {
			return 0, ErrMalformedHeader
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			if length >= 0 {
				return 0, ErrMalformedHeader
			}
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(value))
			if parseErr != nil || parsed < 0 {
				return 0, fmt.Errorf("%w: invalid Content-Length", ErrMalformedHeader)
			}
			length = parsed
		}
	}
}

// WriteMessage writes one JSON payload using LSP Content-Length framing.
func WriteMessage(output io.Writer, payload []byte) error {
	if len(payload) > maxMessageSize {
		return ErrMessageTooLarge
	}
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	return writeAll(output, bytes.Join([][]byte{[]byte(header), payload}, nil))
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
