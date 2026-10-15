package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/contractdecision"
)

func TestSourceRelativeContractEquivalentFormsAndChangedGoals(t *testing.T) {
	model := writeSourceRelativeContractModel(t)
	assignment := interactionAssignmentSource(t)
	direct := directCanonicalSource(t, assignment)
	changed := []byte(strings.Replace(string(direct), `input "9007199254740995" -> "false"`, `input "9007199254740995" -> "true"`, 1))
	if string(changed) == string(direct) {
		t.Fatal("fixture did not change a declared condition")
	}
	var exports []TypedPathContextExport
	for _, source := range [][]byte{assignment, direct, changed} {
		doc := declaredContractDocument(t, source)
		e, err := ExportTypedPathModelContext(context.Background(), "relative.gooo", source, "Choose", doc, model, "")
		if err != nil {
			t.Fatal(err)
		}
		checkSourceRelativeContractArrays(t, e, doc)
		exports = append(exports, e)
	}
	if exports[0].OriginalSourceSHA256 == exports[1].OriginalSourceSHA256 ||
		exports[0].Context.OriginalPlanSHA == exports[1].Context.OriginalPlanSHA {
		t.Fatal("equivalent source representations lost their original identities")
	}
	for i, row := range exports[0].Inputs {
		other, goalChanged := exports[1].Inputs[i], exports[2].Inputs[i]
		if *row.SourceRelativeFeatures != *other.SourceRelativeFeatures || row.InputSHA != other.InputSHA {
			t.Fatal("equivalent calculation changed source-relative input", i)
		}
		if other.InputSHA == goalChanged.InputSHA || *other.CanonicalBranch != *goalChanged.CanonicalBranch ||
			other.SourceRelativeConditions.Buckets == goalChanged.SourceRelativeConditions.Buckets {
			t.Fatal("condition target did not reach its relation input", i)
		}
	}
}

func TestSourceRelativeContractRejectsMislabeledInputFormat(t *testing.T) {
	name := writeSourceRelativeContractModel(t)
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(raw), contractdecision.SourceRelativeOrderedSourceFeatureVersion,
		contractdecision.CanonicalOrderedSourceFeatureVersion, 1)
	if _, err := NewTypedPathGenerator(writeCandidateArtifact(t, []byte(changed))); err == nil {
		t.Fatal("source-relative model accepted canonical input label")
	}
}
