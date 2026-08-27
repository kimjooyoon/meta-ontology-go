package semanticdeltareceiptconsumer

import (
	"encoding/hex"
	"strings"
)

func validSubject(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
