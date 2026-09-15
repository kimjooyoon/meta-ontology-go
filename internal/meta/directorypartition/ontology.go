package directorypartition

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
		"entity PartitionPlan",
		"activity BindPartitionFoundation",
		"activity ResolvePartitionCandidates",
		"activity PreserveProjectRootExemption",
		"activity ReplayDirectoryPartitionPlan",
	}
	source := string(ontologySource)
	for _, token := range required {
		if !strings.Contains(source, token) {
			return "", fmt.Errorf("directory partition ontology misses %q", token)
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

func initialProofs(ontologyDigest, candidateDigest string) []Proof {
	return []Proof{
		{Choice: "foundation", MetaOperation: "bind-partition-ontology", Activity: "BindPartitionFoundation", Satisfied: true, EvidenceDigest: ontologyDigest},
		{Choice: "coherence", MetaOperation: "resolve-directory-partition-plan", Activity: "ResolvePartitionCandidates", Satisfied: true, EvidenceDigest: candidateDigest},
	}
}
