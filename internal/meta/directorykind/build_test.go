package directorykind

import "testing"

func TestBuildCreatesDeterministicReadOnlyKindPlan(t *testing.T) {
	source := kindFixture()
	first, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest != second.Digest || !first.Summary.ReplayVerified {
		t.Fatal("directory kind replay did not reach a fixed point")
	}
	if first.Decision != "PLAN_REVIEW" || len(first.Candidates) != 1 {
		t.Fatalf("unexpected directory kind decision: %+v", first)
	}
	candidate := first.Candidates[0]
	if candidate.GroupCount != 2 || len(candidate.Moves) != 2 || first.Summary.RepositoryWrites != 0 {
		t.Fatalf("invalid read-only kind plan: %+v", candidate)
	}
	if first.Summary.ProjectRootExemptions != 3 {
		t.Fatalf("root exemptions = %d", first.Summary.ProjectRootExemptions)
	}
	for _, indicator := range first.Indicators {
		if !indicator.Satisfied {
			t.Fatalf("unsatisfied planner indicator: %+v", indicator)
		}
	}
}

func TestBuildRecognizesKindFixedPoint(t *testing.T) {
	source := kindFixture()
	source.Meta.Indicators[3].Satisfied = true
	report, err := Build(source)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FIXED_POINT" || len(report.Candidates) != 0 {
		t.Fatalf("expected kind fixed point, got %+v", report)
	}
}

func kindFixture() SourceMetrics {
	root := SourceIndicator{Subject: ".", Applicability: "NOT_APPLICABLE", Satisfied: true}
	return SourceMetrics{Repository: "kimjooyoon/meta-ontology-go", CommitSHA: "0123456789abcdef",
		Files: []SourceFile{{Path: "pkg/a.go", Language: "go", Lines: 10}},
		Directories: []SourceDirectory{{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 1,
			RecursiveFolders: 2, RecursiveFiles: 1}, {Path: "pkg", SubjectKind: "DIRECTORY",
			DirectFolders: 1, DirectFiles: 1, RecursiveFolders: 1, RecursiveFiles: 1},
			{Path: "pkg/child", SubjectKind: "DIRECTORY"}},
		Meta: SourceMeta{Schema: indicatorSchema, Policy: SourcePolicy{Schema: policySchema,
			RequireHomogeneousDirectories: true, ExemptProjectRootTopology: true, ExemptProjectRootREADME: true},
			Indicators: []SourceIndicator{root, root, root, {MetricID: "gooo.metric.layout.entry-kinds.v1",
				Subject: "pkg", SubjectKind: "DIRECTORY", Value: 2, Limit: 1, Applicability: "APPLICABLE",
				Blocking: true, ProofChoice: "foundation", MetaOperation: "separate-directory-kinds"}}}}
}
