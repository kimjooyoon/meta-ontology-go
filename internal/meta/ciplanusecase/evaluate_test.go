package ciplanusecase

import (
	"fmt"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"
)

const scorecardSource = `package ciplan
namespace meta.ciplan

entity ChangeSet id "gooo://meta/ci-plan/entity/change-set"
entity CheckPlan id "gooo://meta/ci-plan/entity/check-plan"
entity VerificationReceipt id "gooo://meta/ci-plan/entity/verification-receipt"

activity SelectGoCheck(ChangeSet) -> CheckPlan computes "ci.rule:go:v1"
activity SelectDocsCheck(ChangeSet) -> CheckPlan computes "ci.rule:docs:v1"
activity SelectYAMLCheck(ChangeSet) -> CheckPlan computes "ci.rule:yaml:v1"
activity PlanCI(ChangeSet) -> CheckPlan computes "ci.plan:v1"
activity VerifyCIPlan(CheckPlan) -> VerificationReceipt computes "ci.verify:v1"
`

func TestFixedScorecardCountsOnlyConformedCases(t *testing.T) {
	input := completeInput(t)
	report := Evaluate(input)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.CasesSatisfied != 12 || report.Summary.PassDecisions != 4 || report.Summary.FailClosedDecisions != 4 || report.Summary.UnknownDecisions != 4 {
		t.Fatalf("decision denominator drifted: %+v", report.Summary)
	}
	if report.Summary.PersistentClaims != 36 || report.Summary.DirectUnknownClaims != 4 || report.Summary.DependencyBlocked != 8 || report.Summary.RefutedClaims != 4 {
		t.Fatalf("claim coordinates drifted: %+v", report.Summary)
	}
	if report.Summary.RuleEvidenceRefs != 6 || report.Summary.GoldenPlans != 4 || report.Summary.DeterministicReplays != 12 {
		t.Fatalf("evidence denominator drifted: %+v", report.Summary)
	}

	broken := input.Reports["pass-go"]
	broken.Claims[2].Status = metainvocation.ClaimOpen
	input.Reports["pass-go"] = broken
	degraded := Evaluate(input)
	if degraded.Summary.CasesSatisfied != 11 || degraded.Decision != "FAIL_CLOSED" {
		t.Fatalf("non-conforming decision was counted as satisfied: %+v", degraded.Summary)
	}
}

func completeInput(t *testing.T) Input {
	t.Helper()
	program, err := metainvocation.Compile("examples/ci-plan/main.gooo", []byte(scorecardSource), metainvocation.StandardRegistry())
	if err != nil {
		t.Fatal(err)
	}
	contract := FixedContract()
	files := map[string][]string{
		"pass-combined": {"internal/metainvocation/model.go", "docs/language/ci-plan.md", ".github/workflows/ci.yml"},
		"pass-docs": {"docs/language/ci-plan.md"},
		"pass-go": {"internal/metainvocation/model.go"},
		"pass-yaml": {".github/workflows/ci.yml"},
		"fail-absolute": {"/tmp/main.go"},
		"fail-duplicate": {"internal/metainvocation/model.go", "internal/metainvocation/model.go"},
		"fail-empty": {},
		"fail-traversal": {"../secret.go"},
		"unknown-asset": {"web/logo.svg"},
		"unknown-license": {"LICENSE"},
		"unknown-python": {"tools/check.py"},
		"unknown-toml": {"config/tool.toml"},
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
	return Input{
		Contract: contract, Reports: reports, Replays: replays, Goldens: goldens, Profile: profile,
		Source: SourceProfile{GoooFiles: 1, GoFiles: 0, GoooLines: 12, GoLines: 0}, GeneratedReplay: true,
	}
}

func mustJSON(files []string) string {
	raw := "["
	for index, file := range files {
		if index != 0 {
			raw += ","
		}
		raw += fmt.Sprintf("%q", file)
	}
	return raw + "]"
}
