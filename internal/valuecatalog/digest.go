package valuecatalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func digestValue(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func reportDigest(report Report) string {
	report.Digest = ""
	return digestValue(report)
}

func finalize(report Report) Report {
	report.Digest = reportDigest(report)
	return report
}

func validDigest(value string) bool {
	return len(value) == 71 && strings.HasPrefix(value, "sha256:")
}

func coordinate(satisfied, total int) Coordinate {
	basisPoints := 0
	if total > 0 {
		basisPoints = satisfied * 10_000 / total
	}
	return Coordinate{Satisfied: satisfied, Total: total, BasisPoints: basisPoints}
}

func countLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	lines := bytes.Count(source, []byte{'\n'})
	if source[len(source)-1] != '\n' {
		lines++
	}
	return lines
}
