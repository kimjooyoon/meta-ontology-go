package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func TestVariantArchiveReferencesFresh1685(t *testing.T) {
	for _, mode := range []string{"unchanged", "changed", "absent"} {
		t.Run(mode, func(t *testing.T) {
			base := fixture{Raw: []byte("receipt"), Inputs: co.SourceInputs{
				Strategy: []byte("plan"), StrategyVerification: []byte("strategy verification"),
				SourceMetrics: []byte("metrics"), Intervention: []byte("ledger"),
				InterventionVerification: []byte("intervention verification"), Program: []byte("program"),
				ProgramSource: []byte("source"), ProgramVerification: []byte("program verification"),
			}}
			v := variant{Receipt: base.Raw, Inputs: base.Inputs}
			wantState, wantFiles := "ORIGINAL_PRODUCER_ARCHIVE", 1
			if mode == "changed" {
				v.Inputs.Strategy, wantState, wantFiles = []byte("changed plan"), "CHANGED_MEMBER", 2
			} else if mode == "absent" {
				v.Inputs.Strategy, wantState = nil, "ABSENT_PAYLOAD"
			}
			directory := filepath.Join(t.TempDir(), "case")
			if err := saveVariant(directory, origin{ArtifactID: 88}, base, v); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(directory, "inputs.json"))
			if err != nil {
				t.Fatal(err)
			}
			var manifest struct{ Inputs []inputReference }
			if err := json.Unmarshal(raw, &manifest); err != nil || len(manifest.Inputs) != 9 {
				t.Fatal("variant input identity population changed", err)
			}
			for _, input := range manifest.Inputs {
				if input.Member == "strategy-plan.json" && input.State != wantState {
					t.Fatal("variant lost its original/changed/absent distinction")
				}
			}
			files, err := os.ReadDir(directory)
			if err != nil || len(files) != wantFiles {
				t.Fatal("unchanged producer inputs were copied", err)
			}
		})
	}
}
