package shadow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type productionProcess struct {
	root string
	bin  string
}

func buildProductionProcess(t *testing.T) productionProcess {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate equivalence test")
	}
	root := filepath.Dir(file)
	for range 4 {
		root = filepath.Dir(root)
	}
	bin := filepath.Join(t.TempDir(), "gooo")
	command := exec.Command("go", "build", "-o", bin, "./cmd/gooo")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build production command: %v\n%s", err, output)
	}
	return productionProcess{root: root, bin: bin}
}

func (p productionProcess) invoke(t *testing.T, fixture productionFixture) productionOutput {
	t.Helper()
	directory := t.TempDir()
	paths := map[string]string{}
	for name, data := range fixture.files {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		paths[name] = path
	}
	args := []string{"selective-ci", "shadow", "--base-snapshot", paths["base_snapshot.json"], "--head-snapshot", paths["head_snapshot.json"], "--plan-input", paths["plan_input.json"], "--evidence-input", paths["evidence_input.json"], "--lane-input", paths["lane_input.json"]}
	command := exec.Command(p.bin, args...)
	command.Dir = p.root
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("production command: %v\nstderr=%s\nstdout=%s", err, stderr.Bytes(), stdout.Bytes())
	}
	if stderr.Len() != 0 {
		t.Fatalf("production command stderr = %q", stderr.String())
	}
	var output productionOutput
	decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&output); err != nil {
		t.Fatalf("decode production receipt: %v\n%s", err, stdout.Bytes())
	}
	return output
}

func TestProductionEquivalenceAgainstIndependentCorpus(t *testing.T) {
	process := buildProductionProcess(t)
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if got := CorpusDigest(); got != "8448b309f64a05c06f75f03352d7516dcb296182af4b922532c28677353ca01e" {
		t.Fatalf("corpus digest changed: %s", got)
	}
	if got := ExpectedVectorDigest(corpus); got != "c48741ac3ba78be5cbd4ede9df04c962e32da0ba2dc761be79c2829749aad213" {
		t.Fatalf("expected vector digest changed: %s", got)
	}
	if len(corpus.Cases) != 33 {
		t.Fatalf("corpus case count = %d, want 33", len(corpus.Cases))
	}
	mismatches := []string{}
	for _, testCase := range corpus.Cases {
		if got := Evaluate(testCase); !reflect.DeepEqual(got, testCase.Expected) {
			mismatches = append(mismatches, fmt.Sprintf("%s: independent oracle changed: got=%#v want=%#v", testCase.Name, got, testCase.Expected))
		}
		fixture := productionPartition(t, testCase.Name)
		output := process.invoke(t, fixture)
		expectation := expectedProduction(testCase.Name)
		if output.ExecutionAuthorized {
			mismatches = append(mismatches, testCase.Name+": production authorized execution")
		}
		if expectation.status == "" {
			mismatches = append(mismatches, testCase.Name+": missing independent production expectation")
			continue
		}
		if output.CanonicalDigest == "" || output.CanonicalDigest != output.selfDigest() {
			mismatches = append(mismatches, fmt.Sprintf("%s: canonical self-digest got %q want %q", testCase.Name, output.CanonicalDigest, output.selfDigest()))
		}
		if expectation.vector != nil {
			if !reflect.DeepEqual(output, *expectation.vector) {
				mismatches = append(mismatches, fmt.Sprintf("%s: positive vector mismatch\ngot=%#v\nwant=%#v", testCase.Name, output, *expectation.vector))
			}
			continue
		}
		if output.Status != expectation.status || output.Stage != expectation.stage || output.Component != expectation.component || output.Reason != expectation.reason {
			mismatches = append(mismatches, fmt.Sprintf("%s: classification got %s/%s/%s/%s want %s/%s/%s/%s", testCase.Name, output.Status, output.Stage, output.Component, output.Reason, expectation.status, expectation.stage, expectation.component, expectation.reason))
		}
		if output.ExecutionAuthorized || !output.ShadowOnly {
			mismatches = append(mismatches, testCase.Name+": fallback execution flags are not closed")
		}
		if len(output.ChangedSemanticIDs) != 0 || len(output.SelectedCommands) != 0 || len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 0 || len(output.ResourceReceipts) != 0 {
			mismatches = append(mismatches, fmt.Sprintf("%s: fallback exposed selection: %#v", testCase.Name, output))
		}
	}
	if len(mismatches) != 0 {
		t.Fatalf("production equivalence mismatches (%d):\n%s", len(mismatches), strings.Join(mismatches, "\n"))
	}
}
