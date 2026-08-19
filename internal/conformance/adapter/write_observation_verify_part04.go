package adapter

import (
	"encoding/hex"
	"strings"
)

func validKind(kind string) bool {
	switch kind {
	case "file", "directory", "symlink", "other":
		return true
	default:
		return false
	}
}
func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}
func comparePrimary(before, after FileObservation, label string) error {
	if before.Exists != after.Exists || before.ByteDigest != after.ByteDigest {
		return oracleError(OracleNW004, label+" bytes or existence changed")
	}
	if before.Lstat != after.Lstat {
		return oracleError(OracleNW005, label+" lstat identity changed")
	}
	return nil
}
