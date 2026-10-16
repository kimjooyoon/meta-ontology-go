package languagereadiness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func TestRegistryOriginReadinessBoundaryFresh1647(t *testing.T) {
	receipt := conceptOperationEvidenceFixture()
	payload, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []bool{false, true} {
		name := "missing-declaration"
		if changed {
			name = "changed-declaration"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if changed {
				path := filepath.Join(root, metricprogram.RegistrySourceOrigin().SourcePath)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("changed declaration"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			input := []byte("{}")
			_, err := EvaluateWithConceptOperation(artifactFixture("PASS", currentConceptIDs...), payload,
				conceptoperation.SourceInputs{
					Strategy: input, StrategyVerification: input, SourceMetrics: input, Intervention: input,
					InterventionVerification: input, Program: input, ProgramSource: input,
					ProgramVerification: input, ScratchDirectory: t.TempDir(),
				}, root, receipt.Repository, receipt.SubjectSHA)
			if err == nil || !strings.Contains(err.Error(), "metric registry member") {
				t.Fatalf("readiness did not stop at the actual registry source frontier: %v", err)
			}
		})
	}
}
