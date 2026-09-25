package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"
)

const payloadLimit = 1 << 20

func decodeArchive(archive []byte, expectedBase string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	var result []byte
	for _, file := range reader.File {
		if path.Base(file.Name) != expectedBase {
			continue
		}
		if result != nil || file.UncompressedSize64 > payloadLimit {
			return nil, fmt.Errorf("artifact payload cardinality or size invalid")
		}
		input, err := file.Open()
		if err != nil {
			return nil, err
		}
		result, err = io.ReadAll(io.LimitReader(input, payloadLimit+1))
		closeErr := input.Close()
		if err != nil || closeErr != nil || len(result) > payloadLimit {
			return nil, fmt.Errorf("artifact payload read failed")
		}
	}
	if result == nil {
		return nil, fmt.Errorf("artifact payload missing")
	}
	return result, nil
}

func verifiedPayload(archive []byte, filename string) ([]byte, error) {
	payload, err := decodeArchive(archive, filename)
	if err != nil {
		return nil, err
	}
	checksum, err := decodeArchive(archive, filename+".sha256")
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:]) + "  " + filename + "\n"
	if string(checksum) != expected {
		return nil, fmt.Errorf("artifact checksum mismatch")
	}
	return payload, nil
}
