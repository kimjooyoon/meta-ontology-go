package metriccounterfactual

import "fmt"

func EvaluateIndicators(
	manifest Manifest,
	plan Plan,
	before State,
	after State,
	receipts []Receipt,
	delta Delta,
) ([]Indicator, error) {
	expectedPlan, err := CounterfactualPlan()
	if err != nil {
		return nil, err
	}
	root := ProjectRootPolicy()
	rootOK := CanonicalEqual(before.RootPolicy, root) && CanonicalEqual(after.RootPolicy, root)
	planOK := CanonicalEqual(plan, expectedPlan) && len(manifest.Files) == 3
	languageOK := delta.GoFiles == 1 && delta.GoLines == 3 &&
		delta.GoooFiles == 0 && delta.GoooLines == 1
	topologyOK := delta.DirectFolders == 1 && delta.DirectFiles == 0 &&
		delta.RecursiveFolders == 2 && delta.RecursiveFiles == 1 &&
		delta.ChangedFiles == 2 && delta.ChangedDirectories == 4
	receiptOK := len(receipts) == 2 &&
		receipts[0].Kind == "APPEND" && receipts[0].Path == "logic/rules.gooo" &&
		receipts[0].BeforeDigest != receipts[0].AfterDigest &&
		receipts[0].AfterLines-receipts[0].BeforeLines == 1 &&
		receipts[1].Kind == "CREATE" && receipts[1].BeforeDigest == "ABSENT" &&
		receipts[1].AfterLines == 3
	boundaryOK := before.Digest != after.Digest &&
		ValidState(before) && ValidState(after) && ValidManifest(manifest) && ValidPlan(plan)
	specifications := []struct {
		id, family, trilemma, expected, actual string
		satisfied                           bool
		evidence                            any
	}{
		{
			"MCF-FOUNDATION-ROOT-001", "FOUNDATION", "AXIOM",
			"counts=OBSERVED topology=NOT_APPLICABLE readme=NOT_APPLICABLE",
			fmt.Sprintf("counts=%s topology=%s readme=%s",
				before.RootPolicy.CountsApplicability,
				before.RootPolicy.TopologyApplicability,
				before.RootPolicy.ReadmeRequirement),
			rootOK, []RootPolicy{before.RootPolicy, after.RootPolicy},
		},
		{
			"MCF-FOUNDATION-PLAN-002", "FOUNDATION", "AXIOM",
			expectedPlan.Digest, plan.Digest, planOK,
			struct {
				Manifest string `json:"manifest"`
				Plan     string `json:"plan"`
			}{manifest.Digest, plan.Digest},
		},
		{
			"MCF-COHERENCE-LANGUAGE-001", "COHERENCE", "COHERENCE",
			"go_files=1 go_lines=3 gooo_files=0 gooo_lines=1",
			fmt.Sprintf("go_files=%d go_lines=%d gooo_files=%d gooo_lines=%d",
				delta.GoFiles, delta.GoLines, delta.GoooFiles, delta.GoooLines),
			languageOK, struct {
				Before string `json:"before"`
				After  string `json:"after"`
				Delta  Delta  `json:"delta"`
			}{before.Digest, after.Digest, delta},
		},
		{
			"MCF-COHERENCE-TOPOLOGY-002", "COHERENCE", "COHERENCE",
			"direct_folders=1 recursive_folders=2 recursive_files=1 changed_files=2 changed_directories=4",
			fmt.Sprintf("direct_folders=%d recursive_folders=%d recursive_files=%d changed_files=%d changed_directories=%d",
				delta.DirectFolders, delta.RecursiveFolders, delta.RecursiveFiles,
				delta.ChangedFiles, delta.ChangedDirectories),
			topologyOK, delta,
		},
		{
			"MCF-REGRESSION-RECEIPTS-001", "REGRESSION", "REGRESS",
			"APPEND:+1 line; CREATE:ABSENT->3 lines",
			fmt.Sprintf("count=%d", len(receipts)), receiptOK, receipts,
		},
		{
			"MCF-REGRESSION-BOUNDARY-002", "REGRESSION", "REGRESS",
			"sealed before != sealed after",
			before.Digest + " != " + after.Digest, boundaryOK,
			[]string{before.Digest, after.Digest},
		},
	}
	indicators := make([]Indicator, 0, len(specifications))
	for _, specification := range specifications {
		digest, err := Digest(specification.evidence)
		if err != nil {
			return nil, err
		}
		status := "UNSATISFIED"
		if specification.satisfied {
			status = "SATISFIED"
		}
		indicators = append(indicators, Indicator{
			ID: specification.id, Family: specification.family,
			Trilemma: specification.trilemma, Expected: specification.expected,
			Actual: specification.actual, Status: status, EvidenceDigest: digest,
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
