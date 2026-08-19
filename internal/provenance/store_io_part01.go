package provenance

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func appendPayload(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create provenance directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open provenance store: %w", err)
	}
	if err := writeFullAt(file, payload, faultLedgerAppendWrite); err != nil {
		_ = file.Close()
		return fmt.Errorf("append provenance evidence: %w", err)
	}
	if err := syncFile(file, faultLedgerAppendSync); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync provenance store: %w", err)
	}
	if err := closeFile(file, faultLedgerAppendClose); err != nil {
		return fmt.Errorf("close provenance store: %w", err)
	}
	return nil
}
func writeFullAt(file *os.File, payload []byte, point storageFaultPoint) error {
	if fault, ok := takeStorageFault(point); ok {
		partial := fault.partial
		if partial < 0 {
			partial = 0
		}
		if partial > len(payload) {
			partial = len(payload)
		}
		if partial > 0 {
			written, err := file.Write(payload[:partial])
			if err != nil {
				return err
			}
			if written != partial {
				return io.ErrShortWrite
			}
		}
		return fault.err
	}
	for len(payload) > 0 {
		written, err := file.Write(payload)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}
func syncFile(file *os.File, point storageFaultPoint) error {
	if err := failStorageOperation(point); err != nil {
		return err
	}
	return file.Sync()
}
