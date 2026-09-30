package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func nativeFeedbackPlan(t *testing.T) map[string]any {
	t.Helper()
	var plan map[string]any
	decoder := json.NewDecoder(strings.NewReader(validBodySearchPlan))
	decoder.UseNumber()
	if err := decoder.Decode(&plan); err != nil {
		t.Fatal(err)
	}
	var typed bodycodegen.IRBodySearchPlan
	if err := json.Unmarshal([]byte(validBodySearchPlan), &typed); err != nil {
		t.Fatal(err)
	}
	training, err := json.Marshal(typed.TestCases)
	if err != nil {
		t.Fatal(err)
	}
	plan["max_attempts"] = 1
	plan["provider_model"] = "english"
	plan["prompt_profile"] = "compact"
	plan["external_training_feedback"] = map[string]any{
		"source_digest":         fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(bodySearchFixture))),
		"training_suite_sha256": fmt.Sprintf("sha256:%x", sha256.Sum256(training)),
		"candidate_id":          "identity",
		"observations": []any{
			map[string]any{"input": -2, "expected": 0, "actual": -2, "passed": false},
			map[string]any{"input": -1, "expected": 0, "actual": -1, "passed": false},
		},
	}
	return plan
}

func runNativeFeedbackPlan(t *testing.T, plan map[string]any) (int, map[string]any) {
	t.Helper()
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"fixture.gooo": []byte(bodySearchFixture), "search-plan.json": raw}
	var stdout, stderr bytes.Buffer
	code := runBodyCodegen(
		[]string{"--json", "--fill-search", "search-plan.json", "--activity", "ClampNegativeToZero", "fixture.gooo"},
		reader, &stdout, &stderr,
	)
	var result map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("code=%d stdout=%s stderr=%s: %v", code, stdout.String(), stderr.String(), err)
	}
	return code, result
}

func TestBodyCodegenCLIUsesPinnedModelAndCompactTrainingFeedback(t *testing.T) {
	requests := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read model request: %v", err)
		}
		requests <- raw
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"model":"mock-laya","routing":{"model":"english"},"answers":{"body_ir_search":{"choice":"zero"}}}`)
	}))
	defer server.Close()
	t.Setenv("GOOO_LAYA_URL", server.URL)
	t.Setenv("GOOO_LAYA_API_KEY", "")
	plan := nativeFeedbackPlan(t)
	code, result := runNativeFeedbackPlan(t, plan)
	if code != exitOK {
		t.Fatalf("native feedback CLI failed: %+v", result)
	}
	search := result["report"].(map[string]any)["body_search"].(map[string]any)
	if search["selected_candidate_id"] != "zero" || search["training_passed"] != float64(5) {
		t.Fatalf("model choice did not flow into tested emission: %+v", search)
	}
	decision := search["attempts"].([]any)[0].(map[string]any)["decision"].(map[string]any)
	if decision["requested_provider_model"] != "english" || decision["mode"] != "laya" {
		t.Fatalf("model binding missing from receipt: %+v", decision)
	}
	raw := <-requests
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil || wire["model"] != "english" {
		t.Fatalf("outbound model binding: %s, error=%v", raw, err)
	}
	feedback := plan["external_training_feedback"].(map[string]any)
	if strings.Contains(string(raw), feedback["source_digest"].(string)) || strings.Contains(string(raw), "holdout") {
		t.Fatalf("local source hash or holdout metadata leaked into prompt: %s", raw)
	}
	if !strings.Contains(string(raw), "external_training_feedback") || !strings.Contains(string(raw), "failed_cases") {
		t.Fatalf("compact feedback missing from prompt: %s", raw)
	}
}

func TestBodyCodegenCLIRejectsInvalidFeedbackBeforeProvider(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	t.Setenv("GOOO_LAYA_URL", server.URL)
	t.Setenv("GOOO_LAYA_API_KEY", "")
	for _, kind := range []string{
		"stale source", "missing actual", "holdout observation", "unsupported model", "unsupported prompt profile",
	} {
		t.Run(kind, func(t *testing.T) {
			plan := nativeFeedbackPlan(t)
			feedback := plan["external_training_feedback"].(map[string]any)
			switch kind {
			case "stale source":
				feedback["source_digest"] = "sha256:" + strings.Repeat("0", 64)
			case "missing actual":
				delete(feedback["observations"].([]any)[0].(map[string]any), "actual")
			case "holdout observation":
				feedback["observations"] = []any{
					map[string]any{"input": int64(-9223372036854775808), "expected": 0, "actual": 0, "passed": true},
				}
			case "unsupported model":
				plan["provider_model"] = "unknown"
			case "unsupported prompt profile":
				plan["prompt_profile"] = "unknown"
			}
			code, result := runNativeFeedbackPlan(t, plan)
			if code == exitOK || result["decision"] != "FAIL_CLOSED" {
				t.Fatalf("invalid feedback accepted: %+v", result)
			}
		})
	}
	if posts.Load() != 0 {
		t.Fatalf("invalid feedback made %d provider calls", posts.Load())
	}
}

func TestBodyCodegenCLIFeedbackKeepsOfflineEmissionDeterministic(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	withFeedback := nativeFeedbackPlan(t)
	withoutFeedback := nativeFeedbackPlan(t)
	delete(withoutFeedback, "external_training_feedback")
	delete(withoutFeedback, "provider_model")
	delete(withoutFeedback, "prompt_profile")
	var sources []string
	for _, plan := range []map[string]any{withFeedback, withFeedback, withoutFeedback} {
		code, result := runNativeFeedbackPlan(t, plan)
		if code != exitOK {
			t.Fatalf("offline feedback CLI failed: %+v", result)
		}
		sources = append(sources, result["source"].(string))
	}
	if sources[0] != sources[1] || sources[0] != sources[2] {
		t.Fatal("advisory feedback or model pin changed declared-order offline emission")
	}
}
