package main

import (
	"bytes"
	"encoding/json"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	"reflect"
	"testing"
)

func TestSelectiveCIShadowPositiveSelfDigestAndNoExecution(t *testing.T) {
	fixture := newShadowFixture(t)
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("shadow stderr = %q", stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.Status != "SHADOW_SELECTIVE" || output.Stage != "SELECTIVE" || output.Reason != "VERIFIED" {
		t.Fatalf("shadow output classification = %#v", output)
	}
	if output.ExecutionAuthorized || !output.ShadowOnly {
		t.Fatalf("shadow execution flags = authorized:%t shadow_only:%t", output.ExecutionAuthorized, output.ShadowOnly)
	}
	if len(output.SelectedCommands) != 1 || output.SelectedCommands[0].ID != fixture.commandID || !reflect.DeepEqual(output.SelectedCommands[0].Argv, []string{"gooo-shadow-sentinel", "never-run"}) {
		t.Fatalf("selected command projection = %#v", output.SelectedCommands)
	}
	if len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 1 || len(output.ResourceReceipts) != 1 {
		t.Fatalf("selected receipt projection = guards:%#v work:%#v receipts:%#v", output.SelectedGuards, output.SelectedWorkIDs, output.ResourceReceipts)
	}
	if output.ResourceReceipts[0].CPUWorkUnits != fixture.commandCPU || output.ResourceReceipts[0].MemoryBytes != fixture.commandMemory {
		t.Fatalf("resource receipt = %#v", output.ResourceReceipts[0])
	}
	if output.Lane.BaseSHA != fixture.laneInput.BaseSHA || output.Lane.LaneHeadSHA != fixture.laneInput.LaneHeadSHA || output.Lane.LaneID != fixture.laneInput.LaneID || output.Lane.Reason != string(lanesci.ReasonEligible) {
		t.Fatalf("lane projection = %#v", output.Lane)
	}
	if output.CanonicalDigest == "" || output.CanonicalDigest != output.stableDigest() {
		t.Fatalf("output self digest = %q", output.CanonicalDigest)
	}
	t.Logf("canonical receipt digest=%s", output.CanonicalDigest)
	if bytes.Contains(stdout.Bytes(), []byte(`"execution_authorized":true`)) {
		t.Fatal("shadow receipt authorized execution")
	}
}
func TestSelectiveCIShadowInputPermutationIsByteStable(t *testing.T) {
	left := newShadowFixture(t)
	right := newShadowFixture(t)
	right.reverseInputs()
	var leftOut, leftErr, rightOut, rightErr bytes.Buffer
	if code := runSelectiveCI(left.args(), left.reader(), &leftOut, &leftErr); code != exitOK {
		t.Fatalf("left code = %d, stderr = %q", code, leftErr.String())
	}
	if code := runSelectiveCI(right.argsReversed(), right.reader(), &rightOut, &rightErr); code != exitOK {
		t.Fatalf("right code = %d, stderr = %q", code, rightErr.String())
	}
	if !bytes.Equal(leftOut.Bytes(), rightOut.Bytes()) {
		t.Fatalf("input permutation changed receipt:\nleft=%s\nright=%s", leftOut.Bytes(), rightOut.Bytes())
	}
}
