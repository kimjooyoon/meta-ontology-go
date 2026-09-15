package directorypartition

import "testing"

func TestBuildCreatesDeterministicReadOnlyPlan(t *testing.T) {
	source := fixtureSource()
	first, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest != second.Digest || !first.Summary.ReplayVerified {
		t.Fatal("replay did not reach a deterministic fixed point")
	}
	if first.Decision != "PLAN_REVIEW" || len(first.Candidates) != 1 {
		t.Fatalf("unexpected partition decision: %+v", first)
	}
	if first.Summary.ProjectRootExemptions != 1 || first.Summary.RepositoryWrites != 0 {
		t.Fatalf("root or write guardrail regressed: %+v", first.Summary)
	}
	for _, indicator := range first.Indicators {
		if !indicator.Satisfied {
			t.Fatalf("unsatisfied planner indicator: %+v", indicator)
		}
	}
}

func TestBuildRecognizesFixedPoint(t *testing.T) {
	source := fixtureSource()
	source.Meta.Indicators[1].Satisfied = true
	report, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FIXED_POINT" || len(report.Candidates) != 0 {
		t.Fatalf("expected fixed point, got %+v", report)
	}
}

func fixtureSource() SourceMetrics {
	return SourceMetrics{
		Repository: "kimjooyoon/meta-ontology-go", CommitSHA: "0123456789abcdef",
		Files: []SourceFile{
			{Path: "pkg/a.go", Language: "go", Lines: 10},
			{Path: "pkg/b.go", Language: "go", Lines: 10},
			{Path: "pkg/model.gooo", Language: "gooo", Lines: 10},
		},
		Directories: []SourceDirectory{
			{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 1, RecursiveFolders: 1, RecursiveFiles: 3},
			{Path: "pkg", SubjectKind: "DIRECTORY", DirectFiles: 3, RecursiveFiles: 3},
		},
		Meta: SourceMeta{
			Schema: indicatorSchema,
			Policy: SourcePolicy{Schema: policySchema, MaxDirectDirectoryEntries: 2, ExemptProjectRootTopology: true, ExemptProjectRootREADME: true},
			Indicators: []SourceIndicator{
				{MetricID: "root", Subject: ".", SubjectKind: "PROJECT_ROOT", Applicability: "NOT_APPLICABLE", ApplicabilityReason: "ROOT_TOPOLOGY_EXEMPT", Satisfied: true},
				{MetricID: "entries", Subject: "pkg", SubjectKind: "DIRECTORY", Value: 3, Limit: 2, Applicability: "APPLICABLE", Blocking: true, Satisfied: false, ProofChoice: "foundation", MetaOperation: "partition-directory"},
			},
		},
	}
}
