package directorykind

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed ontology.gooo
var ontologySource []byte

func validateOntology() (string, error) {
	required := []string{
		"entity SourceMetrics",
		"entity KindSeparationPlan",
		"activity BindDirectoryKindFoundation",
		"activity ResolveMixedDirectoryKinds",
		"activity PlanDirectoryKindSeparation",
		"activity PreserveProjectRootExemption",
		"activity ReplayDirectoryKindSeparation",
	}
	source := string(ontologySource)
	for _, token := range required {
		if !strings.Contains(source, token) {
			return "", fmt.Errorf("directory kind ontology misses %q", token)
		}
	}
	return digest(source)
}

func digest(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func sealReport(report Report) (Report, error) {
	report.Digest = ""
	value, err := digest(report)
	report.Digest = value
	return report, err
}
