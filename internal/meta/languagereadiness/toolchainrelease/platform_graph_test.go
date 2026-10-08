package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func graphSmokeFixture(t *testing.T) map[string]any {
	t.Helper()
	graph := jointdecision.RecordGraphInput{Nodes: []jointdecision.RecordGraphNode{{Kind: "input", Input: 1, InputType: "string"}}}
	for i, name := range []string{"title", "state", "note"} {
		graph.Choices[i] = jointdecision.RecordGraphChoice{Field: name, First: "input", Second: "input", Intent: "Keep the input",
			FieldID: "record://" + name, Roots: [2]uint16{1, 1}}
	}
	text, err := jointdecision.EncodeRecordGraphThree(graph)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]any{
		"original_source_sha256": "sha256:source", "model_predictions": 0, "candidate_tests": 0,
		"context": map[string]any{"status": "ENCODED", "feature_version": jointdecision.RecordGraphSharedFeatureVersion,
			"text": text, "sha256": fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(text)))},
		"value_flow": map[string]any{"status": "RESOLVED", "nodes": []any{map[string]any{"id": 1}}},
	}
}

func TestSourceGraphSmokeRequiresExactSourceAndNoExecution(t *testing.T) {
	raw, _ := json.Marshal(graphSmokeFixture(t))
	if err := validateSourceGraphSmoke(raw, "sha256:source"); err != nil {
		t.Fatal(err)
	}
	if err := validateSourceGraphSmoke(raw, "sha256:other"); err == nil {
		t.Fatal("wrong source accepted")
	}
	if err := validateSourceGraphSmoke([]byte(`{}`), "sha256:source"); err == nil {
		t.Fatal("missing observation accepted")
	}
	for _, mutate := range []func(map[string]any){
		func(r map[string]any) { delete(r, "model_predictions") },
		func(r map[string]any) { delete(r, "candidate_tests") },
		func(r map[string]any) { r["model_predictions"] = 1 },
		func(r map[string]any) { r["candidate_tests"] = 1 },
		func(r map[string]any) { r["context"].(map[string]any)["status"] = "DECLINED_TO_DETERMINISTIC" },
		func(r map[string]any) {
			r["context"].(map[string]any)["feature_version"] = jointdecision.RecordSharedFeatureVersion
		},
		func(r map[string]any) { r["context"].(map[string]any)["sha256"] = "sha256:other" },
		func(r map[string]any) { r["value_flow"].(map[string]any)["status"] = "UNRESOLVED" },
		func(r map[string]any) { r["value_flow"].(map[string]any)["nodes"] = []any{} },
		func(r map[string]any) {
			c := r["context"].(map[string]any)
			c["text"] = "malformed graph"
			c["sha256"] = fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("malformed graph")))
		},
	} {
		r := graphSmokeFixture(t)
		mutate(r)
		raw, _ := json.Marshal(r)
		if err := validateSourceGraphSmoke(raw, "sha256:source"); err == nil {
			t.Fatal("incomplete graph smoke accepted")
		}
	}
}
