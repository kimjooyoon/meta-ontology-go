package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func commandIDToID(value string) semantic.ID { return semantic.MustIdentity(value) }
func shadowDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func prefixedShadowDigest(value string) string { return "sha256:" + shadowDigest(value) }
