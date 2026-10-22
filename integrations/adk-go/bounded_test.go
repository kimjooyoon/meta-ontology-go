package adkcheck

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// A plain reader has no WriterTo method, so io.Copy would choose a promoted
// bytes.Buffer.ReadFrom if limitedBuffer accidentally exposed one.
type plainReader struct{ reader io.Reader }

func (r plainReader) Read(p []byte) (int, error) { return r.reader.Read(p) }

func TestLimitedBufferCopyBoundary(t *testing.T) {
	for _, size := range []int{maxReportBytes, maxReportBytes + 1} {
		var target limitedBuffer
		_, err := io.Copy(&target, plainReader{bytes.NewReader(bytes.Repeat([]byte{'x'}, size))})
		if size == maxReportBytes && err != nil {
			t.Fatalf("exact limit rejected: %v", err)
		}
		if size > maxReportBytes && !errors.Is(err, errReportTooLarge) {
			t.Fatalf("oversized output was accepted: %v", err)
		}
		if len(target.Bytes()) > maxReportBytes {
			t.Fatalf("buffer exceeded limit: %d", len(target.Bytes()))
		}
	}
}
