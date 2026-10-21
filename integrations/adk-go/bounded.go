package adkcheck

import (
	"bytes"
	"errors"
)

const maxReportBytes = 1 << 20

var errReportTooLarge = errors.New("Gooo check output exceeds 1 MiB")

type limitedBuffer struct {
	bytes.Buffer
	err error
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > maxReportBytes {
		b.err = errReportTooLarge
		return 0, b.err
	}
	return b.Buffer.Write(p)
}
