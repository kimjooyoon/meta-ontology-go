package lsp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

type nonCloseableReader struct {
	started  chan struct{}
	release  chan struct{}
	finished chan struct{}
	start    sync.Once
	finish   sync.Once
}

func newNonCloseableReader() *nonCloseableReader {
	return &nonCloseableReader{
		started: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{}),
	}
}

func (reader *nonCloseableReader) Read([]byte) (int, error) {
	reader.start.Do(func() { close(reader.started) })
	<-reader.release
	reader.finish.Do(func() { close(reader.finished) })
	return 0, io.EOF
}

func TestReadMessageHandlesPartialInput(t *testing.T) {
	input := []byte("Content-Length: 11\r\n\r\n{\"ok\":true}")
	payload, err := ReadMessage(&chunkReader{data: input, size: 1})
	if err != nil || string(payload) != `{"ok":true}` {
		t.Fatalf("ReadMessage() = %q, error = %v", payload, err)
	}
}

func TestReadMessageRejectsMalformedAndPartialFrames(t *testing.T) {
	if _, err := ReadMessage(strings.NewReader("Content-Length 4\r\n\r\ntest")); !errors.Is(err, ErrMalformedHeader) {
		t.Fatalf("malformed header error = %v", err)
	}
	if _, err := ReadMessage(strings.NewReader("Content-Length: 5\r\n\r\nabc")); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("partial payload error = %v", err)
	}
}

func TestReadMessageRejectsTruncatedHeaders(t *testing.T) {
	for _, input := range []string{
		"Content-Length: 4\r\n",
		"Content-Length: 4",
	} {
		if _, err := ReadMessage(strings.NewReader(input)); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Fatalf("truncated header %q error = %v, want io.ErrUnexpectedEOF", input, err)
		}
	}
}

func TestReadMessageKeepsSequentialFrames(t *testing.T) {
	var input bytes.Buffer
	writeFrameForTest(t, &input, []byte(`{"text":"😀"}`))
	writeFrameForTest(t, &input, []byte(`{"text":"second"}`))
	first, firstErr := ReadMessage(&input)
	second, secondErr := ReadMessage(&input)
	if firstErr != nil || secondErr != nil || string(first) != `{"text":"😀"}` || string(second) != `{"text":"second"}` {
		t.Fatalf("frames = %q, %q; errors = %v, %v", first, second, firstErr, secondErr)
	}
}

func TestWriteMessageHandlesShortWrites(t *testing.T) {
	var output bytes.Buffer
	if err := WriteMessage(&shortWriter{output: &output, size: 2}, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("WriteMessage() error = %v", err)
	}
	payload, err := readFrame(bufio.NewReader(&output))
	if err != nil || string(payload) != `{"ok":true}` {
		t.Fatalf("framed output = %q, error = %v", payload, err)
	}
}
