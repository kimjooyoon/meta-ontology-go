package main

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// GitHub upload outputs may omit the prefix retained by artifact metadata.
func canonicalDigest(value string) (string, error) {
	hexadecimal := strings.TrimPrefix(value, "sha256:")
	raw, err := hex.DecodeString(hexadecimal)
	if err != nil || len(raw) != 32 || hexadecimal != strings.ToLower(hexadecimal) {
		return "", fmt.Errorf("producer digest is not lowercase sha256")
	}
	return "sha256:" + hexadecimal, nil
}
