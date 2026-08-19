package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func caseDigest(c Case) string {
	value := struct {
		Name      string `json:"name"`
		Partition string `json:"partition"`
		Files     Files  `json:"files"`
	}{c.Name, c.Partition, c.Files}
	return hashJSON(value)
}
func hashJSON(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func validNonEmpty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
