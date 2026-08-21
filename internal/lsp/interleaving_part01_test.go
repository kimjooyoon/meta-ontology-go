package lsp

import (
	"bytes"
	"io"
	"sync"
)

type diagnosticsBuffer struct {
	mu     sync.Mutex
	data   bytes.Buffer
	count  int
	first  chan struct{}
	second chan struct{}
}

func newDiagnosticsBuffer() *diagnosticsBuffer {
	return &diagnosticsBuffer{first: make(chan struct{}), second: make(chan struct{})}
}
func (buffer *diagnosticsBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	_, _ = buffer.data.Write(data)
	if bytes.Contains(data, []byte(`"method":"textDocument/publishDiagnostics"`)) {
		buffer.count++
		if buffer.count == 1 {
			close(buffer.first)
		}
		if buffer.count == 2 {
			close(buffer.second)
		}
	}
	return len(data), nil
}
func (buffer *diagnosticsBuffer) Bytes() []byte {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return append([]byte(nil), buffer.data.Bytes()...)
}

type trackedPipeReader struct {
	reader  *io.PipeReader
	started chan struct{}
	once    sync.Once
}
type nonInterruptingCloser struct {
	started chan struct{}
	release chan struct{}
	closed  chan struct{}
	start   sync.Once
	close   sync.Once
}

func newNonInterruptingCloser() *nonInterruptingCloser {
	return &nonInterruptingCloser{
		started: make(chan struct{}), release: make(chan struct{}), closed: make(chan struct{}),
	}
}
func (reader *nonInterruptingCloser) Read([]byte) (int, error) {
	reader.start.Do(func() { close(reader.started) })
	<-reader.release
	return 0, io.EOF
}
func (reader *nonInterruptingCloser) Close() error {
	reader.close.Do(func() { close(reader.closed) })
	return nil
}
func (reader *trackedPipeReader) Read(data []byte) (int, error) {
	reader.once.Do(func() { close(reader.started) })
	return reader.reader.Read(data)
}
func (reader *trackedPipeReader) Close() error {
	return reader.reader.Close()
}
