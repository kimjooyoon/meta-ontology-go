package coupling

import (
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"unicode"
)

func validID(value string) bool {
	_, err := semantic.ParseIdentity(value)
	return err == nil
}
func validToken(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
