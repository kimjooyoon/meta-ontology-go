package cache

import (
	"encoding/binary"
	"hash"
)

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func writeKeyPart(hasher hash.Hash, value string) {
	var length [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(length[:], uint64(len(value)))
	_, _ = hasher.Write(length[:n])
	_, _ = hasher.Write([]byte(value))
}
