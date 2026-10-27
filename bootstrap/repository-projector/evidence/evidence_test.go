package evidence

import "testing"

func TestGeneratedLineDebtKeepsObservationButSelectsItsOwner(t *testing.T) {
	report := Build("source", []Entry{{Logical: "generated.go", Backing: "object", ObjectSHA: "digest", Language: "go", Lines: 100, GeneratorOwned: true}}, 1, 0, Topology{})
	if len(report.Subjects) != 1 {
		t.Fatal(report)
	}
	got := report.Subjects[0]
	if got.Value != 100 || got.Applicability != "NOT_APPLICABLE" || got.ApplicabilityReason != "GENERATED_SOURCE_OWNER_REQUIRED" || got.Operation != "regenerate-source-owner" {
		t.Fatal(got)
	}
	for _, indicator := range report.Indicators {
		if indicator.ID == "source.line-cap-debt" && indicator.Value != 1 {
			t.Fatal("the observed size debt was erased", indicator)
		}
	}
}

func TestBuildDoesNotClassifyLongGoooAsGoSplitDebt(t *testing.T) {
	report := Build("sha256:subject", []Entry{
		{Logical: "large.go", Backing: "objects/large", ObjectSHA: "sha256:go", Language: "go", Lines: 76},
		{Logical: "policy.gooo", Backing: "objects/policy", ObjectSHA: "sha256:gooo", Language: "gooo", Lines: 163},
	}, 0, 0, Topology{})
	if len(report.Subjects) != 1 || report.Subjects[0].Logical != "large.go" || report.Subjects[0].Indicator != "source.line-cap-debt" {
		t.Fatalf("unexpected source split subjects: %#v", report.Subjects)
	}
	if report.Subjects[0].Value != 76 || report.Subjects[0].Limit != 75 {
		t.Fatalf("Go line-cap evidence changed: %#v", report.Subjects[0])
	}
}
