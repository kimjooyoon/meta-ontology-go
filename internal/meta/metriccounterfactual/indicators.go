package metriccounterfactual

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

type indicatorCheck struct {
	id, family, trilemma, expected, actual string
	satisfied                              bool
	evidence                               any
}

func EvaluateIndicators(manifest Manifest, plan Plan, before, after State, receipts []Receipt, delta Delta) ([]Indicator, error) {
	expectedPlan, err := CounterfactualPlan()
	if err != nil {
		return nil, err
	}
	root := ProjectRootPolicy()
	rootOK := artifact.Equal(before.RootPolicy, root) && artifact.Equal(after.RootPolicy, root)
	planOK := artifact.Equal(plan, expectedPlan) && len(manifest.Files) == 3
	languageOK := delta.GoFiles == 1 && delta.GoLines == 3 && delta.GoooFiles == 0 && delta.GoooLines == 1
	topologyOK := delta.DirectFolders == 1 && delta.DirectFiles == 0 && delta.RecursiveFolders == 2 &&
		delta.RecursiveFiles == 1 && delta.ChangedFiles == 2 && delta.ChangedDirectories == 4
	receiptOK := len(receipts) == 2 && receipts[0].Kind == "APPEND" && receipts[0].Path == "logic/rules.gooo" &&
		receipts[0].BeforeDigest != receipts[0].AfterDigest && receipts[0].AfterLines-receipts[0].BeforeLines == 1 &&
		receipts[1].Kind == "CREATE" && receipts[1].BeforeDigest == "ABSENT" && receipts[1].AfterLines == 3
	boundaryOK := before.Digest != after.Digest && ValidState(before) && ValidState(after) &&
		ValidManifest(manifest) && ValidPlan(plan)
	checks := []indicatorCheck{
		{"MCF-FOUNDATION-ROOT-001", "FOUNDATION", "AXIOM", "counts=OBSERVED topology=NOT_APPLICABLE readme=NOT_APPLICABLE", fmt.Sprintf("counts=%s topology=%s readme=%s", before.RootPolicy.CountsApplicability, before.RootPolicy.TopologyApplicability, before.RootPolicy.ReadmeRequirement), rootOK, []RootPolicy{before.RootPolicy, after.RootPolicy}},
		{"MCF-FOUNDATION-PLAN-002", "FOUNDATION", "AXIOM", expectedPlan.Digest, plan.Digest, planOK, []string{manifest.Digest, plan.Digest}},
		{"MCF-COHERENCE-LANGUAGE-001", "COHERENCE", "COHERENCE", "go_files=1 go_lines=3 gooo_files=0 gooo_lines=1", fmt.Sprintf("go_files=%d go_lines=%d gooo_files=%d gooo_lines=%d", delta.GoFiles, delta.GoLines, delta.GoooFiles, delta.GoooLines), languageOK, []any{before.Digest, after.Digest, delta}},
		{"MCF-COHERENCE-TOPOLOGY-002", "COHERENCE", "COHERENCE", "direct_folders=1 recursive_folders=2 recursive_files=1 changed_files=2 changed_directories=4", fmt.Sprintf("direct_folders=%d recursive_folders=%d recursive_files=%d changed_files=%d changed_directories=%d", delta.DirectFolders, delta.RecursiveFolders, delta.RecursiveFiles, delta.ChangedFiles, delta.ChangedDirectories), topologyOK, delta},
		{"MCF-REGRESSION-RECEIPTS-001", "REGRESSION", "REGRESS", "APPEND:+1 line; CREATE:ABSENT->3 lines", fmt.Sprintf("count=%d", len(receipts)), receiptOK, receipts},
		{"MCF-REGRESSION-BOUNDARY-002", "REGRESSION", "REGRESS", "sealed before != sealed after", before.Digest + " != " + after.Digest, boundaryOK, []string{before.Digest, after.Digest}},
	}
	indicators := make([]Indicator, 0, len(checks))
	for _, check := range checks {
		digest, err := artifact.Digest(check.evidence)
		if err != nil {
			return nil, err
		}
		status := "UNSATISFIED"
		if check.satisfied {
			status = "SATISFIED"
		}
		indicators = append(indicators, Indicator{
			ID: check.id, Family: check.family, Trilemma: check.trilemma,
			Expected: check.expected, Actual: check.actual, Status: status, EvidenceDigest: digest,
		})
	}
	return indicators, nil
}

func AllSatisfied(indicators []Indicator) bool {
	for _, indicator := range indicators {
		if indicator.Status != "SATISFIED" {
			return false
		}
	}
	return len(indicators) > 0
}
