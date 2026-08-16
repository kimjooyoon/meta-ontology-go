package metarecognition

import (
	"reflect"
	"strings"
	"testing"
)

func TestClosedCorpus(t *testing.T) {
	cases := Corpus()
	if len(cases) < 17 {
		t.Fatalf("closed corpus cases = %d, want at least 17", len(cases))
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			gooo, baseline := Evaluate(c)
			if gooo.State != c.Expected.State || gooo.Reason != c.Expected.Reason || !equalIDs(gooo.LocalizedIDs, c.Expected.LocalizedIDs) {
				t.Fatalf("gooo = %#v, want %#v", gooo, c.Expected)
			}
			if baseline.State != c.Expected.State || baseline.Reason != c.Expected.Reason || !equalIDs(baseline.LocalizedIDs, c.Expected.LocalizedIDs) {
				t.Fatalf("baseline = %#v, want %#v", baseline, c.Expected)
			}
		})
	}
}

func TestCanonicalReplayAndManifest(t *testing.T) {
	cases := Corpus()
	forward, err := Run(cases)
	if err != nil {
		t.Fatal(err)
	}
	reverseCases := append([]Case(nil), cases...)
	for left, right := 0, len(reverseCases)-1; left < right; left, right = left+1, right-1 {
		reverseCases[left], reverseCases[right] = reverseCases[right], reverseCases[left]
	}
	reverse, err := Run(reverseCases)
	if err != nil {
		t.Fatal(err)
	}
	forwardJSON, err := forward.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	reverseJSON, err := reverse.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(forwardJSON) != string(reverseJSON) {
		t.Fatal("canonical manifest changed under case permutation")
	}
	if forward.Finding != NoUniqueBenefit {
		t.Fatalf("finding = %s, want %s", forward.Finding, NoUniqueBenefit)
	}
	if !forward.Summary.ExactOutcomeVector || !forward.Summary.ExactReasonLocalizationVector {
		t.Fatal("comparison vectors are not exact")
	}
	if forward.Summary.GoooFalsePasses != 0 || forward.Summary.GoooFalseNegatives != 0 || forward.Summary.BaselineFalsePasses != 0 || forward.Summary.BaselineFalseNegatives != 0 {
		t.Fatalf("false-pass/negative counts = %#v", forward.Summary)
	}
	if forward.Summary.BaselineWorkUnits > forward.Summary.GoooWorkUnits || !forward.Summary.GoooRatio.Known || !forward.Summary.BaselineRatio.Known {
		t.Fatalf("fair-baseline work/ratio rule failed: %#v", forward.Summary)
	}
	digest, err := ManifestDigest(forward)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("manifest_json=%s", forwardJSON)
	t.Logf("manifest_digest=%s", digest)
}

func TestGoooIgnoresExpectedMutation(t *testing.T) {
	caseValue := Corpus()[0]
	before, _ := Evaluate(caseValue)
	caseValue.Expected.Reason = ReasonExternalMissing
	caseValue.Expected.LocalizedIDs = []string{"forged://id"}
	after, _ := Evaluate(caseValue)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("gooo changed after expected mutation: before=%#v after=%#v", before, after)
	}
	manifest, err := Run([]Case{caseValue})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Summary.ExactReasonLocalizationVector {
		t.Fatal("manifest accepted forged expected reason/localization")
	}
}

func TestBaselineCountersDoNotAffectGooo(t *testing.T) {
	caseValue := Corpus()[0]
	beforeGooo, beforeBaseline := Evaluate(caseValue)
	caseValue.Baseline.FullCommands = 99
	caseValue.Baseline.SelectedCommands = 88
	caseValue.Baseline.ProvRecords = 77
	caseValue.Baseline.ProvPaths = 66
	afterGooo, afterBaseline := Evaluate(caseValue)
	if !reflect.DeepEqual(beforeGooo, afterGooo) {
		t.Fatalf("gooo copied baseline counters: before=%#v after=%#v", beforeGooo, afterGooo)
	}
	if reflect.DeepEqual(beforeBaseline, afterBaseline) {
		t.Fatal("baseline counters did not affect baseline accounting")
	}
}

func TestReplayCanonicalizesRootPathAndOrder(t *testing.T) {
	original := Corpus()
	original[0].Baseline.Roots = []string{"root://a", "root://z"}
	original[0].Baseline.Path.IDs = []string{"path://a", "path://z"}
	perturbed := append([]Case(nil), original...)
	perturbed[0].Baseline.Roots = []string{"root://z", "root://a"}
	perturbed[0].Baseline.Path.IDs = []string{"path://z", "path://a"}
	commands := append([]CommandAssertion(nil), perturbed[6].Baseline.Commands...)
	for left, right := 0, len(commands)-1; left < right; left, right = left+1, right-1 {
		commands[left], commands[right] = commands[right], commands[left]
	}
	perturbed[6].Baseline.Commands = commands
	for left, right := 0, len(perturbed)-1; left < right; left, right = left+1, right-1 {
		perturbed[left], perturbed[right] = perturbed[right], perturbed[left]
	}
	first, err := ReplayJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ReplayJSON(perturbed)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("replay changed under root/path/order permutation")
	}
	decoded, err := DecodeReplayJSON(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Cases) != len(original) {
		t.Fatalf("decoded cases = %d, want %d", len(decoded.Cases), len(original))
	}
	if decoded.Cases[0].ID != "case-01" {
		t.Fatalf("decoded cases are not canonicalized: first=%q", decoded.Cases[0].ID)
	}
}

func TestReplayRejectsSchemaAndDuplicates(t *testing.T) {
	caseJSON := `{"id":"case-x","subject":"SEMANTIC_BINDING","roots":[],"commands":[],"paths":[]}`
	valid := `{"schema":"` + SchemaVersion + `","cases":[` + caseJSON + `]}`
	unknown := strings.Replace(valid, `,"cases"`, `,"extra":true,"cases"`, 1)
	duplicateCase := `{"schema":"` + SchemaVersion + `","cases":[` + caseJSON + `,` + caseJSON + `]}`
	duplicateCommand := `{"schema":"` + SchemaVersion + `","cases":[{"id":"case-x","subject":"SEMANTIC_BINDING","roots":[],"commands":["cmd","cmd"],"paths":[]}]}`
	duplicatePath := `{"schema":"` + SchemaVersion + `","cases":[{"id":"case-x","subject":"SEMANTIC_BINDING","roots":[],"commands":[],"paths":["path","path"]}]}`
	duplicateField := `{"schema":"` + SchemaVersion + `","schema":"` + SchemaVersion + `","cases":[]}`
	for name, data := range map[string]string{"unknown": unknown, "case": duplicateCase, "command": duplicateCommand, "path": duplicatePath, "field": duplicateField} {
		if _, err := DecodeReplayJSON([]byte(data)); err == nil {
			t.Errorf("DecodeReplayJSON accepted %s duplicate/invalid input", name)
		}
	}
	if _, err := DecodeReplayJSON([]byte(valid)); err != nil {
		t.Fatal(err)
	}
}
