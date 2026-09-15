package lsp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func readFrame(reader *bufio.Reader) ([]byte, error) {
	length := -1
	headerStarted := false
	for {
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			if !headerStarted && len(line) == 0 {
				return nil, io.EOF
			}
			return nil, io.ErrUnexpectedEOF
		}
		if err != nil {
			return nil, err
		}
		headerStarted = true
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
			if length >= 0 {
				return nil, ErrMalformedHeader
			}
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
