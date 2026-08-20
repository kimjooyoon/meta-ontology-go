package generator

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode/utf8"
)

func markerManifestStringV1(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("generator: marker manifest v1 string is not valid UTF-8")
	}
	return strconv.Itoa(len([]byte(value))) + ":" + hex.EncodeToString([]byte(value)), nil
}

func markerLineStartV1(source []byte, offset int) bool {
	return offset == 0 || source[offset-1] == '\n'
}

func markerLineEndV1(source []byte, offset int) bool {
	return offset == len(source) || offset > 0 && source[offset-1] == '\n'
}

func markerLineStartIndexV1(source []byte, offset int) int {
	return bytes.Count(source[:offset], []byte{'\n'})
}

func markerLineEndIndexV1(source []byte, offset int) int {
	index := markerLineStartIndexV1(source, offset)
	if offset > 0 && source[offset-1] == '\n' {
		return index - 1
	}
	return index
}
