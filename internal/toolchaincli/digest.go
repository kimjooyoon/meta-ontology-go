package toolchaincli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

func digestJSON(value any) string {
	raw, _ := json.Marshal(value)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
