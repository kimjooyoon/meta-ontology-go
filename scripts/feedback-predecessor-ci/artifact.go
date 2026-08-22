package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"path"
)

const receiptLimit = 1 << 20

func decodeReceipt(archive []byte) archivedReceipt {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return archivedReceipt{}
	}
	for _, file := range reader.File {
		if path.Base(file.Name) != "artifact-feedback-resolution-receipt.json" ||
			file.UncompressedSize64 > receiptLimit {
			continue
		}
		input, err := file.Open()
		if err != nil {
			return archivedReceipt{}
		}
		data, readErr := io.ReadAll(io.LimitReader(input, receiptLimit+1))
		closeErr := input.Close()
		if readErr != nil || closeErr != nil || len(data) > receiptLimit {
			return archivedReceipt{}
		}
		var receipt archivedReceipt
		if json.Unmarshal(data, &receipt) != nil {
			return archivedReceipt{}
		}
		return receipt
	}
	return archivedReceipt{}
}
