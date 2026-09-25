package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func entityFieldsMetadata(support entityFieldsSupport) *EntityFieldsMetadata {
	return &EntityFieldsMetadata{
		State: string(support.State),
		Profile: EntityFieldsProfileMetadata{
			ID: support.Profile.ID, Version: support.Profile.Version, Digest: support.Profile.Digest,
		},
	}
}
func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}
func digestIR(ir SemanticIR) string {
	normalized, err := normalizeIR(ir)
	if err != nil {
		return ""
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return ""
	}
	return digestBytes(payload)
}
func digestSourceMap(sourceMap SourceMap) string {
	payload, err := json.Marshal(sourceMap)
	if err != nil {
		return ""
	}
	return digestBytes(payload)
}
