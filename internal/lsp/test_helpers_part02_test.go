package lsp

import (
	"bytes"
	"io"
)

type chunkReader struct {
	data []byte
	size int
	read int
}

func (reader *chunkReader) Read(buffer []byte) (int, error) {
	if reader.read == len(reader.data) {
		return 0, io.EOF
	}
	size := reader.size
	if size > len(buffer) {
		size = len(buffer)
	}
	if size > len(reader.data)-reader.read {
		size = len(reader.data) - reader.read
	}
	copy(buffer, reader.data[reader.read:reader.read+size])
	reader.read += size
	return size, nil
}

type shortWriter struct {
	output *bytes.Buffer
	size   int
}

func (writer *shortWriter) Write(data []byte) (int, error) {
	size := writer.size
	if size > len(data) {
		size = len(data)
	}
	_, _ = writer.output.Write(data[:size])
	return size, nil
}
