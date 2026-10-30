package generation

import (
	"reflect"
	"strings"
	"testing"
)

func TestGoSplitContractRetainsThreeSourceLinkedStages(t *testing.T) {
	evidence, err := GoSplitInputContractEvidence()
	if err != nil || len(evidence.Obligations) != 3 || len(evidence.SourceOrigins.Declarations) != 26 {
		t.Fatalf("split source contract is incomplete: %+v, %v", evidence, err)
	}
	want := [][2]string{
		{"file-input", "rendered-capacity-obligation"},
		{"rendered-capacity-obligation", "directory-capacity-obligation"},
		{"directory-capacity-obligation", "projected-conformance-obligation"},
	}
	for index, step := range evidence.Obligations {
		if !step.UsedInputFact || !step.GeneratedOutputFact ||
			step.InputEntityID != "gooo://meta-operation-input-contract/entity/"+want[index][0] ||
			step.OutputEntityID != "gooo://meta-operation-input-contract/entity/"+want[index][1] {
			t.Fatalf("source dependency was lost: %+v", step)
		}
	}
	extraction, err := ExtractFunctionInputContractEvidence()
	if err != nil || len(extraction.Obligations) != 6 {
		t.Fatal("split prerequisites changed extraction's six obligations", err)
	}
}

func TestGoSplitContractActivityLabelsKeepTheirIDs(t *testing.T) {
	raw := string(operationContractFixture(t))
	before, err := parseOperationInputContract([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"ProveGoSplitRenderedCapacity", "ProveGoSplitDirectoryCapacity", "ProveGoSplitProjectedConformance",
	} {
		after, err := parseOperationInputContract([]byte(strings.ReplaceAll(raw, name, "Renamed"+name)))
		if err != nil || before.SourceDigest == after.SourceDigest || before.SemanticDigest != after.SemanticDigest ||
			!reflect.DeepEqual(before.ObligationFacts, after.ObligationFacts) {
			t.Fatalf("same-ID activity label changed the prerequisite meaning: %s: %v", name, err)
		}
	}
}

func TestGoSplitContractRejectsMissingChangedAndReorderedLinks(t *testing.T) {
	raw := string(operationContractFixture(t))
	const activityID = "gooo://meta-operation-input-contract/activity/go-split-directory-capacity"
	for name, source := range map[string]string{
		"missing stage": strings.ReplaceAll(raw,
			"activity ProveGoSplitDirectoryCapacity(RenderedCapacityObligation) -> DirectoryCapacityObligation\n"+
				"    id \""+activityID+"\"\n", ""),
		"changed input": strings.ReplaceAll(raw, "ProveGoSplitDirectoryCapacity(RenderedCapacityObligation)",
			"ProveGoSplitDirectoryCapacity(FileInput)"),
		"changed ID": strings.ReplaceAll(raw, activityID, activityID+"-unknown"),
		"reordered dependency": strings.ReplaceAll(raw, "ProveGoSplitProjectedConformance(DirectoryCapacityObligation)",
			"ProveGoSplitProjectedConformance(RenderedCapacityObligation)"),
	} {
		t.Run(name, func(t *testing.T) {
			if source == raw {
				t.Fatal("source control did not change the authored program")
			}
			got, err := parseOperationInputContract([]byte(source))
			if err == nil || !reflect.DeepEqual(got, operationInputContract{}) {
				t.Fatalf("invalid prerequisite chain produced admission: %v", err)
			}
		})
	}
}
