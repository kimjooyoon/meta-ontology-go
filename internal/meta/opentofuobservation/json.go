package opentofuobservation

import (
	"encoding/json"
	"fmt"
	"io"
)

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytesReader(raw))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func bytesReader(raw []byte) *byteReader { return &byteReader{raw: raw} }

type byteReader struct {
	raw []byte
	pos int
}

func (reader *byteReader) Read(p []byte) (int, error) {
	if reader.pos == len(reader.raw) {
		return 0, io.EOF
	}
	n := copy(p, reader.raw[reader.pos:])
	reader.pos += n
	return n, nil
}
