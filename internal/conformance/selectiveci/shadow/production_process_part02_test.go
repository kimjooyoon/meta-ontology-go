package shadow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestProductionEquivalenceAgainstIndependentCorpus(t *testing.T) {
	process := buildProductionProcess(t)
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if got := CorpusDigest(); got != "1749e4a01627483ca9b3f6ecb20e83244abb729b24501600dbe2ee553d295ca3" {
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
