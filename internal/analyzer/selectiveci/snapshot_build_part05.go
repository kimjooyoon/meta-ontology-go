package selectiveci

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"path"
	"strings"
)

func normalizeID(raw string) (string, error) {
	if raw != strings.TrimSpace(raw) {
		return "", fail(CodeInvalidBinding, "semantic ID is padded")
	}
	id, err := semantic.ParseIdentity(raw)
	if err != nil {
		return "", fail(CodeInvalidBinding, "semantic ID %q is invalid: %v", raw, err)
	}
	if id.String() != raw {
		return "", fail(CodeInvalidBinding, "semantic ID %q is not canonical", raw)
	}
	return raw, nil
}
func normalizeRepoPath(raw string) (string, error) {
	if raw == "" || raw != strings.TrimSpace(raw) || strings.Contains(raw, "\\") {
		return "", fail(CodeMalformedPath, "repository path %q is malformed", raw)
	}
	if strings.HasPrefix(raw, "/") || path.IsAbs(raw) || (len(raw) >= 2 && raw[1] == ':') {
		return "", fail(CodeMalformedPath, "repository path %q must be relative", raw)
	}
	clean := path.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fail(CodeMalformedPath, "repository path %q escapes the repository", raw)
	}
	return clean, nil
}
func normalizeDigest(raw, label string) (string, error) {
	if !validDigest(raw) {
		return "", fail(CodeMalformedDigest, "%s %q is not a lowercase sha256 digest", label, raw)
	}
	return raw, nil
}
func validDigest(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+sha256.Size*2 || !strings.HasPrefix(value, prefix) {
		return false
	}
	for _, char := range value[len(prefix):] {
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}
func validRawDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}
func validRole(role semanticbinding.Role) bool {
	switch role {
	case semanticbinding.RoleHandwrittenImpl, semanticbinding.RoleGeneratedImpl, semanticbinding.RoleAdapter:
		return true
	default:
		return false
	}
}
func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
