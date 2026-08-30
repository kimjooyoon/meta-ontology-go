package governancesnapshot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func LoadSnapshot(path, root string) (LoadedSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LoadedSnapshot{}, err
	}
	var snapshot Snapshot
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return LoadedSnapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return LoadedSnapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	loaded := LoadedSnapshot{Snapshot: snapshot, Payloads: map[string][]byte{}}
	for _, request := range snapshot.Requests {
		if request.State != "PRESENT" {
			continue
		}
		if filepath.IsAbs(request.PayloadPath) || filepath.Clean(request.PayloadPath) != request.PayloadPath || strings.HasPrefix(request.PayloadPath, ".."+string(filepath.Separator)) {
			return LoadedSnapshot{}, fmt.Errorf("snapshot payload path is not relative: %q", request.PayloadPath)
		}
		payload, err := os.ReadFile(filepath.Join(root, request.PayloadPath))
		if err != nil {
			return LoadedSnapshot{}, fmt.Errorf("read snapshot payload %s: %w", request.ID, err)
		}
		loaded.Payloads[request.ID] = payload
	}
	return loaded, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func normalizedDigest(raw []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	if err := ensureEOF(decoder); err != nil {
		return "", err
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(normalized)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func digestJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
