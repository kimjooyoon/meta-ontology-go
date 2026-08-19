package lanefrontier

import (
	"crypto/sha256"
	"encoding/hex"
)

func CorpusDigest() string {
	data, _ := corpusFile.ReadFile("corpus.json")
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
