package toolchainconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	return digestPattern.MatchString(value)
}

func validHead(value string) bool {
	return headPattern.MatchString(value)
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestValue(report)
	return report
}

func digestArtifacts(artifacts map[string][]byte) string {
	type item struct {
		ID  string `json:"id"`
		Raw []byte `json:"raw"`
	}
	items := make([]item, 0, len(fixedSurfaces)+1)
	for _, definition := range fixedSurfaces {
		if raw, ok := artifacts[definition.ID]; ok {
			items = append(items, item{definition.ID, raw})
		}
	}
	if raw, ok := artifacts["unexpected-surface"]; ok {
		items = append(items, item{"unexpected-surface", raw})
	}
	return digestValue(items)
}
