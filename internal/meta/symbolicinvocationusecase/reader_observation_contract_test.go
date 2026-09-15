package symbolicinvocationusecase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSymbolicReaderObservationContractMatchesEvaluator(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "symbolic-invocation-usecase", "reader-observation-contract.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Schema      string `json:"schema"`
		MetricID    string `json:"metric_id"`
		Denominator struct {
			Total        int            `json:"total"`
			Classes      map[string]int `json:"classes"`
			ProofChoices map[string]int `json:"proof_choices"`
		} `json:"denominator"`
		FailureResolution string `json:"failure_resolution"`
	}
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.Schema != SymbolicReaderObservationSchema || contract.MetricID != SymbolicReaderObservationMetric || contract.Denominator.Total != SymbolicReaderObservationTotal || contract.FailureResolution != "FAIL_CLOSED" {
		t.Fatalf("contract=%+v", contract)
	}
	classes := contract.Denominator.Classes
	if classes["OUTCOME"] != 3 || classes["DRIVER"] != 3 || classes["GUARDRAIL"] != 4 {
		t.Fatalf("classes=%v", classes)
	}
	proofs := contract.Denominator.ProofChoices
	if proofs["FOUNDATION"] != 4 || proofs["COHERENCE"] != 3 || proofs["REGRESSION"] != 3 {
		t.Fatalf("proof_choices=%v", proofs)
	}
}
