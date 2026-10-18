package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	packageReceiptMaxBytes     = 32 << 20
	packageReceiptGzipMaxBytes = 2 << 20
)

func validatePackageReceiptCompression(compression string, jsonMode bool) error {
	if compression != "" && (compression != "gzip" || !jsonMode) {
		return fmt.Errorf("--receipt-compression requires --json and the value gzip")
	}
	return nil
}

// Encoding changes the transport only. The JSON encoder and receipt fields are
// identical to plain output, including unresolved observation counters.
func writePackageReceiptJSON(output io.Writer, value any, compression string) error {
	if err := validatePackageReceiptCompression(compression, true); err != nil {
		return err
	}
	var compressed *gzip.Writer
	if compression == "gzip" {
		var err error
		compressed, err = gzip.NewWriterLevel(&packageReceiptLimitWriter{output: output, limit: packageReceiptGzipMaxBytes}, gzip.BestSpeed)
		if err != nil {
			return err
		}
		output = &packageReceiptLimitWriter{output: compressed, limit: packageReceiptMaxBytes}
	}
	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(value)
	if compressed != nil {
		if closeErr := compressed.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

type packageReceiptLimitWriter struct {
	output  io.Writer
	limit   int
	written int
}

func (w *packageReceiptLimitWriter) Write(p []byte) (int, error) {
	if len(p) > w.limit-w.written {
		return 0, fmt.Errorf("receipt output exceeds %d bytes", w.limit)
	}
	n, err := w.output.Write(p)
	w.written += n
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}

func readPackageReceipt(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("receipt must be a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > packageReceiptMaxBytes {
		return nil, inputLimitError(packageReceiptMaxBytes)
	}
	var header [2]byte
	if _, err := file.ReadAt(header[:], 0); err != nil && err != io.EOF {
		return nil, err
	}
	wireLimit := packageReceiptMaxBytes
	if header == [2]byte{0x1f, 0x8b} {
		wireLimit = packageReceiptGzipMaxBytes
	}
	if info.Size() > int64(wireLimit) {
		return nil, inputLimitError(int64(wireLimit))
	}
	raw, err := io.ReadAll(io.LimitReader(file, int64(wireLimit)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > wireLimit {
		return nil, inputLimitError(int64(wireLimit))
	}
	return decodePackageReceiptTransport(raw, packageReceiptMaxBytes)
}

// Hashes downstream always name these original JSON bytes. Transport filenames,
// compression headers and compressed bytes never replace a source identity.
func decodePackageReceiptTransport(raw []byte, limit int) ([]byte, error) {
	if len(raw) == 0 || len(raw) > limit {
		return nil, inputLimitError(int64(limit))
	}
	if !bytes.HasPrefix(raw, []byte{0x1f, 0x8b}) {
		return raw, nil
	}
	if len(raw) > packageReceiptGzipMaxBytes {
		return nil, fmt.Errorf("compressed package receipt exceeds %d bytes", packageReceiptGzipMaxBytes)
	}
	input := bytes.NewReader(raw)
	reader, err := gzip.NewReader(input)
	if err != nil {
		return nil, fmt.Errorf("read compressed package receipt: %w", err)
	}
	reader.Multistream(false)
	decoded, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	closeErr := reader.Close()
	if err != nil {
		return nil, fmt.Errorf("decode compressed package receipt: %w", err)
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(decoded) == 0 || len(decoded) > limit {
		return nil, inputLimitError(int64(limit))
	}
	if input.Len() != 0 {
		return nil, fmt.Errorf("compressed package receipt has trailing data or multiple members")
	}
	return decoded, nil
}
