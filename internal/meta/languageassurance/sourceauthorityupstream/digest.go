package sourceauthorityupstream

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

func digestValue(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func selectLines(document []byte, selection Selection) ([]byte, error) {
	lines := bytes.Split(document, []byte{'\n'})
	if selection.StartLine < 1 || selection.EndLine < selection.StartLine || selection.EndLine > len(lines) {
		return nil, errors.New("selection is outside source")
	}
	selected := bytes.Join(lines[selection.StartLine-1:selection.EndLine], []byte{'\n'})
	if bytes.Contains(selected, []byte{'\r'}) {
		return nil, errors.New("carriage return is not canonical")
	}
	return selected, nil
}
