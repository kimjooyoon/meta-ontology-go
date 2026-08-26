package languagesyntax

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
)

func digestJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return digestBytes(encoded)
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func registryDigest() string { return digestJSON(expectedRegistry()) }

func caseDigest(item CaseResult, source Source) string {
	item.EvidenceDigest = ""
	return digestJSON(struct {
		Case   CaseResult `json:"case"`
		Source Source     `json:"source"`
	}{item, source})
}

func seal(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != 71 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validHead(value string) bool {
	if len(value) != 40 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func conceptBound(repository fs.FS, artifact languageconcept.Artifact) bool {
	if languageconcept.ValidateArtifact(repository, artifact) != nil {
		return false
	}
	for _, concept := range artifact.Report.Concepts {
		if concept.ID == conceptID {
			return true
		}
	}
	return false
}
