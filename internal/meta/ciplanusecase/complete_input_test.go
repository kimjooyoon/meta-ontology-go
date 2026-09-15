package ciplanusecase

import (
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

func completeInput(t *testing.T) Input {
	t.Helper()
	program, err := metainvocation.Compile("examples/ci-plan/main.gooo", []byte(scorecardSource), metainvocation.StandardRegistry())
	if err != nil {
		t.Fatal(err)
	}
	contract := FixedContract()
	files := map[string][]string{
		"pass-combined":   {"internal/metainvocation/model.go", "docs/language/ci-plan.md", ".github/workflows/ci.yml"},
		"pass-docs":       {"docs/language/ci-plan.md"},
		"pass-go":         {"internal/metainvocation/model.go"},
		"pass-yaml":       {".github/workflows/ci.yml"},
		"fail-absolute":   {"/tmp/main.go"},
		"fail-duplicate":  {"internal/metainvocation/model.go", "internal/metainvocation/model.go"},
		"fail-empty":      {},
		"fail-traversal":  {"../secret.go"},
		"unknown-asset":   {"web/logo.svg"},
		"unknown-license": {"LICENSE"},
		"unknown-python":  {"tools/check.py"},
		"unknown-toml":    {"config/tool.toml"},
	}
	reports := map[string]metainvocation.Report{}
	replays := map[string]metainvocation.Report{}
	goldens := map[string]GoldenPlan{}
	profile := Profile{Schema: "gooo/ci-plan-resource-profile/v1", Samples: []ProfileSample{}}
	for _, spec := range contract.Cases {
		raw := []byte(fmt.Sprintf(`{"schema":"gooo/ci-plan-input/v1","case_id":%q,"files":%s}`, spec.ID, mustJSON(files[spec.ID])))
		report, invokeErr := metainvocation.Invoke(program, "PlanCI", raw)
		if invokeErr != nil {
			t.Fatal(invokeErr)
		}
		reports[spec.ID] = report
		replays[spec.ID] = report
		if spec.ExpectedDecision == metainvocation.DecisionPass {
			goldens[spec.ID] = ProjectGolden(report)
		}
		profile.Samples = append(profile.Samples, ProfileSample{CaseID: spec.ID, WallMS: 10, PeakRSSKiB: 1024, ReceiptBytes: 2048})
	}
	return Input{Contract: contract, Reports: reports, Replays: replays, Goldens: goldens, Profile: profile, Source: SourceProfile{GoooFiles: 1, GoFiles: 0, GoooLines: 12, GoLines: 0}, GeneratedReplay: true}
}
