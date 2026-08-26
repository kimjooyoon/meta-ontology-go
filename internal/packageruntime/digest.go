package packageruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestValue(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func imageDigest(image Image) string {
	image.Digest = ""
	return digestValue(image)
}

func resultDigest(result Result) string {
	result.ResultDigest = ""
	return digestValue(result)
}
