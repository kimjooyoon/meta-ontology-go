package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

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

func TestProjectionOwnershipReachesPublicSplitPlanner(t *testing.T) {
	root := t.TempDir()
	var source strings.Builder
	source.WriteString("package fixture\n")
	for index := range 99 {
		fmt.Fprintf(&source, "const Value%d = %d\n", index, index)
	}
	if err := os.WriteFile(filepath.Join(root, "handwritten.go"), []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	report := Build("head", []Entry{
		{Logical: "generated.go", Backing: "object", ObjectSHA: "digest", Language: "go", Lines: 100, GeneratorOwned: true},
		{Logical: "handwritten.go", Backing: "object", ObjectSHA: "other", Language: "go", Lines: 100},
	}, 2, 0, Topology{})
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	evidence, plan := filepath.Join(root, "evidence.json"), filepath.Join(root, "plan.json")
	if err := os.WriteFile(evidence, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "run", "./bootstrap/logical-split-planner", "--root", root, "--evidence", evidence, "--expected-sha", "head", "--output", plan)
	command.Dir, command.Env = repository, append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	result, err := os.ReadFile(plan)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Subjects []struct {
			Logical string `json:"logical"`
		} `json:"subjects"`
	}
	if err := json.Unmarshal(result, &decoded); err != nil || len(decoded.Subjects) != 1 || decoded.Subjects[0].Logical != "handwritten.go" {
		t.Fatal(string(result), err)
	}
}
