package adkcheck

import (
	"bytes"
	"errors"
)

const maxReportBytes = 1 << 20

var errReportTooLarge = errors.New("Gooo check output exceeds 1 MiB")

type limitedBuffer struct {
	buf bytes.Buffer
	err error
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > maxReportBytes-b.buf.Len() {
		b.err = errReportTooLarge
		return 0, b.err
	}
	return b.buf.Write(p)
}

func (b *limitedBuffer) Bytes() []byte  { return b.buf.Bytes() }
func (b *limitedBuffer) String() string { return b.buf.String() }
