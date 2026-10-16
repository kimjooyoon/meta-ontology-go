// Package originpack preserves bounded raw origin bytes in a checked envelope.
// It is intended for detached storage, never the compiler's execution path.
package originpack

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

const MaxRawBytes = 128 << 10
const headerBytes = 45
const magic = "GOOOOR1\n"

// Pack uses gzip only when it reduces payload size. SHA-256 covers raw bytes.
func Pack(raw []byte) ([]byte, error) {
	if len(raw) > MaxRawBytes {
		return nil, fmt.Errorf("ORIGIN_RAW_BOUND")
	}
	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(raw); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	mode, payload := byte(0), raw
	if compressed.Len() < len(raw) {
		mode, payload = 1, compressed.Bytes()
	}
	result := make([]byte, headerBytes+len(payload))
	copy(result, magic)
	result[8] = mode
	binary.LittleEndian.PutUint32(result[9:13], uint32(len(raw)))
	digest := sha256.Sum256(raw)
	copy(result[13:headerBytes], digest[:])
	copy(result[headerBytes:], payload)
	return result, nil
}

// Unpack restores exact raw bytes; it never parses or reformats their contents.
func Unpack(packed []byte) ([]byte, error) {
	if len(packed) < headerBytes || len(packed) > MaxRawBytes+headerBytes || string(packed[:8]) != magic {
		return nil, fmt.Errorf("ORIGIN_ENVELOPE_INVALID")
	}
	size := binary.LittleEndian.Uint32(packed[9:13])
	if size > MaxRawBytes {
		return nil, fmt.Errorf("ORIGIN_RAW_BOUND")
	}
	var raw []byte
	switch packed[8] {
	case 0:
		raw = bytes.Clone(packed[headerBytes:])
	case 1:
		var err error
		raw, err = inflate(packed[headerBytes:], int64(size))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("ORIGIN_ENCODING_UNAVAILABLE")
	}
	digest := sha256.Sum256(raw)
	if int64(len(raw)) != int64(size) || !bytes.Equal(digest[:], packed[13:headerBytes]) {
		return nil, fmt.Errorf("ORIGIN_RAW_MISMATCH")
	}
	return raw, nil
}

func inflate(payload []byte, size int64) ([]byte, error) {
	input := bytes.NewReader(payload)
	reader, err := gzip.NewReader(input)
	if err != nil {
		return nil, fmt.Errorf("ORIGIN_GZIP_INVALID: %w", err)
	}
	defer reader.Close()
	reader.Multistream(false)
	raw, err := io.ReadAll(io.LimitReader(reader, size+1))
	if err != nil || int64(len(raw)) != size || input.Len() != 0 {
		return nil, fmt.Errorf("ORIGIN_GZIP_PAYLOAD_INVALID")
	}
	return raw, nil
}
