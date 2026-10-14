package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func TestLengthIntegerAuthoredFlow20537(t *testing.T) {
	source := []byte(lengthContextPrelude20537 + "activity Choose(Text,Integer)->Summary computes `return Summary{size:len(input0)+input1,present:true}` assembling {\n" +
		"choice \"size\" field_value at \"0\" alternative \"20537\" intent \"Retain input byte length plus the offset.\"\n" +
		"value_case \"[\\\"가A\\\",7]\" -> \"{\\\"size\\\":11,\\\"present\\\":true}\"\nattempts \"2\"\n}\n")
	result, err := ExportRecordAssemblyContextWithFlow(context.Background(), "길이흐름20537.gooo", source, "Choose", false, jointdecision.RecordSharedFeatureVersion)
	if err != nil || result.ValueFlow == nil || result.ValueFlow.Status != "RESOLVED" {
		t.Fatal("new length-choice symbolic flow", result.ValueFlow, err)
	}
	if result.CandidateTests != 0 || result.ModelPredictions != 0 || len(result.Choices) != 1 || result.Choices[0].First != "len(input0)+input1" {
		t.Fatal("source choice or observation scope changed", result)
	}
	lengths := 0
	for _, node := range result.ValueFlow.Nodes {
		if node.Operator == "int64" {
			t.Fatal("implicit representation conversion claimed as authored intent", node)
		}
		if node.Operator == "len" {
			lengths++
		}
	}
	if lengths != 1 {
		t.Fatal("original runtime length node lost", lengths)
	}
	if root := os.Getenv("GOOO_FRESH_SOURCE_OUTPUT"); root != "" {
		raw, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "length-flow20537.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("new source choice retains authored len and arithmetic;implicit widening absent from intent graph;candidate/model/native0")
}
