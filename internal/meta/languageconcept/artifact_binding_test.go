package languageconcept

import (
	"testing"
	"testing/fstest"
)

func TestArtifactDigestChangesWithBoundCode(t *testing.T) {
	concepts := []Concept{artifactFixtureConcept("meta.go")}
	first := buildArtifact(fstest.MapFS{
		"meta.go": &fstest.MapFile{Data: []byte("package meta\nconst value = 1\n")},
	}, concepts)
	second := buildArtifact(fstest.MapFS{
		"meta.go": &fstest.MapFile{Data: []byte("package meta\nconst value = 2\n")},
	}, concepts)
	if first.Report.ReportDigest != second.Report.ReportDigest {
		t.Fatal("catalog report changed when only bound code changed")
	}
	if first.Bindings.Digest == second.Bindings.Digest || first.ArtifactDigest == second.ArtifactDigest {
		t.Fatal("bound code did not change artifact identity")
	}
}

func TestMissingBoundCodeFailsClosed(t *testing.T) {
	artifact := buildArtifact(fstest.MapFS{}, []Concept{artifactFixtureConcept("missing.go")})
	if artifact.Decision != "FAIL_CLOSED" || artifact.Bindings.Missing == 0 {
		t.Fatalf("artifact = %#v", artifact)
	}
}

func artifactFixtureConcept(path string) Concept {
	return Concept{
		ID: "fixture", Problem: "scalar claim", PositiveEffect: "bound claim",
		MetaOperation: "bind-fixture", Rarity: "UNCOMMON_COMBINATION", Stage: "OPERATING",
		CodeBindings: []string{path}, MetricBindings: []string{"fixture.metric"},
		UseCases: []UseCase{{ID: "fixture", Trigger: "input", ExpectedOutcome: "PASS"}},
	}
}
