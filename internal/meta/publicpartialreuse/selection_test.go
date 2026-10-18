package publicpartialreuse

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestGoooSelectionPrecedesPartitionExecution(t *testing.T) {
	for _, item := range []struct {
		caseID   string
		want     []string
		decision string
	}{
		{"NO_CHANGE_ALL_RECEIPTS_REUSED", nil, DecisionClosed},
		{"SINGLE_PARTITION_CHANGE_SELECTIVE_REUSE", []string{"orders"}, DecisionClosed},
		{"MISSING_DEPENDENCY_EDGE", nil, DecisionUnknown},
		{"UNBOUNDED_IMPACT", nil, DecisionUnknown},
		{"CHANGED_HIDDEN_DEPENDENCY", nil, DecisionRefuted},
	} {
		t.Run(item.caseID, func(t *testing.T) {
			input := receiptEvaluationFixture(t)
			setReceiptCase(t, &input, item.caseID)
			selection, err := Select(context.Background(), input)
			if err != nil || selection.Assessment.Decision != item.decision {
				t.Fatal("source selection failed", selection, err)
			}
			var ids []string
			for _, partition := range selection.Partitions {
				ids = append(ids, partition.ID)
			}
			if !reflect.DeepEqual(ids, item.want) || selection.Origin.ActivityID != AssessmentActivityID {
				t.Fatal("selection lost its source or partition identities", ids, selection.Origin)
			}
		})
	}
}

func TestGoooSourceOnlyChangeSelectsAdditionalWork(t *testing.T) {
	input := receiptEvaluationFixture(t)
	before, err := Select(context.Background(), input)
	if err != nil || len(before.Partitions) != 0 {
		t.Fatal("unchanged case unexpectedly schedules tests", before, err)
	}
	source := bytes.Replace(input.Policy.decisionSource, []byte("result.execute_partitions = input.impacted"), []byte("result.execute_partitions = []bool{true, true}"), 1)
	source = bytes.Replace(source, []byte("result.test_units_required = impacted_count"), []byte("result.test_units_required = 2"), 1)
	source = bytes.Replace(source, []byte("result.test_units_reusable = 2 - impacted_count"), []byte("result.test_units_reusable = 0"), 1)
	input.Policy.decisionSource, input.Policy.DecisionSourceDigest = source, cache.HashBytes(source).String()
	after, err := Select(context.Background(), input)
	if err != nil || !reflect.DeepEqual(after.Partitions, input.Policy.Partitions) ||
		before.Origin.SourceSHA256 == after.Origin.SourceSHA256 || before.Origin.ProjectionSHA256 == after.Origin.ProjectionSHA256 {
		t.Fatal("source-only selection did not schedule the two declared partitions", after, err)
	}
}

func TestGoooSelectionRejectsIncompleteProjectionAndMissingReceipts(t *testing.T) {
	for _, mask := range []string{"[]bool{}", "[]bool{true}", "[]bool{true, true}"} {
		t.Run(mask, func(t *testing.T) {
			input := receiptEvaluationFixture(t)
			source := bytes.Replace(input.Policy.decisionSource, []byte("result.execute_partitions = input.impacted"), []byte("result.execute_partitions = "+mask), 1)
			input.Policy.decisionSource, input.Policy.DecisionSourceDigest = source, cache.HashBytes(source).String()
			selection, err := Select(context.Background(), input)
			if err == nil || len(selection.Partitions) != 0 {
				t.Fatal("inconsistent source projection exposed executable work", selection, err)
			}
		})
	}
	input := receiptEvaluationFixture(t)
	delete(input.Receipts, "orders")
	selection, err := Select(context.Background(), input)
	if err != nil || selection.Assessment.Decision != DecisionUnknown || len(selection.Partitions) != 0 {
		t.Fatal("missing receipt exposed executable work", selection, err)
	}
}
