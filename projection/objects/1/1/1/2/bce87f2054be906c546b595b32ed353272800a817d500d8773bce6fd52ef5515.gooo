package metarecognition

import (
	"crypto/sha256"
	"encoding/hex"
)

func ManifestDigest(m Manifest) (string, error) {
	data, err := m.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
