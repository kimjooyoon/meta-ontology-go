package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}

func read(path string) []byte {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		file, err := os.Open(path + ".gz")
		require(err == nil, "compressed input unavailable")
		defer file.Close()
		reader, err := gzip.NewReader(file)
		require(err == nil, "invalid gzip")
		defer reader.Close()
		raw, err = io.ReadAll(reader)
		require(err == nil, "gzip read failed")
		return raw
	}
	require(err == nil, "input unavailable")
	return raw
}

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func main() {
	require(len(os.Args) == 2, "observation directory required")
	root := os.Args[1]
	var output []map[string]any
	var baseline bodycodegen.Result
	for _, arm := range []string{"baseline", "feedback"} {
		var result bodycodegen.Result
		require(json.Unmarshal(read(filepath.Join(root, arm+".json")), &result) == nil, "invalid typed observation")
		require(result.Report.CompilerSourceSHA == "5fba132655ec9145a5fa7bccb5411ace8341bf02", "producer changed")
		p := result.Report.BodyPaths
		require(p != nil && p.Search.Status == "TRAINING_COMPLETE" && p.FunctionalCompleteness == 100, "incomplete result")
		require(p.Conditions != nil && p.Conditions.Passed == 3 && p.Conditions.Declared == 3, "condition results changed")
		require(p.Search.Selection.WeightsSHA256 == "dcd8e44591626d421d4961bfef82ec197e947cb1d5d2cd92868d908bc7de4aed", "weights changed")
		require(p.Search.Selection.ExternalCallsKnown && p.Search.Selection.ExternalCalls == 0, "unknown external calls")
		require(len(p.NativeCases) == 7, "native suite changed")
		large := 0
		for _, row := range p.NativeCases {
			require(row.Passed && row.Expected == row.Actual, "native output failed")
			if row.Input == -9007199254740995 || row.Input == 9007199254740995 {
				require(row.Actual == 9007199254740995, "large integer rounded")
				large++
			}
		}
		require(large == 2, "large native cases absent")
		var predictNS int64
		minBytes, maxBytes, newCalls := 512, 0, 0
		for _, receipt := range p.Search.Selection.Receipts {
			predictNS += receipt.PredictNS
		}
		for _, receipt := range p.Feedback {
			require(receipt.Applied && !receipt.ContextDeclined && receipt.FirstConditionFailure != nil, "feedback missing")
			require(receipt.FirstConditionFailure.Result.Case.Input == -9007199254740995, "counterexample rounded")
			newCalls += receipt.ModelCalls
			for _, judgment := range receipt.Judgments {
				require(hash([]byte(judgment.Input)) == judgment.InputSHA, "input digest differs")
				require(strings.Contains(judgment.Input, "condition_input=-9007199254740995") &&
					strings.Contains(judgment.Input, "condition_expected=true condition_actual=false condition_status=MISMATCH"), "counterexample missing from model input")
				require(len(judgment.Input) <= 512, "model input overflow")
				minBytes, maxBytes = min(minBytes, len(judgment.Input)), max(maxBytes, len(judgment.Input))
				predictNS += judgment.Prediction.PredictNS
			}
		}
		if arm == "baseline" {
			baseline = result
			require(len(p.Feedback) == 0 && p.Search.Evaluated == 5 && p.Search.Selection.ModelCalls == 3, "baseline counts differ")
			minBytes = 0
		} else {
			require(result.Source == baseline.Source && result.GoooSource == baseline.GoooSource, "selected source differs")
			require(p.OriginalSourceSHA256 == baseline.Report.BodyPaths.OriginalSourceSHA256 && p.TestSuiteSHA256 == baseline.Report.BodyPaths.TestSuiteSHA256, "source or cases changed")
			require(len(p.Feedback) == 2 && newCalls == 6 && p.Search.Evaluated == 3 && p.Search.Selection.ModelCalls == 9, "feedback counts differ")
		}
		output = append(output, map[string]any{"arm": arm, "candidate_evaluations": p.Search.Evaluated,
			"condition_rejections": p.Search.ConditionRejected, "model_predictions": p.Search.Selection.ModelCalls,
			"feedback_predictions": newCalls, "prediction_total_ns": predictNS,
			"feedback_input_min_bytes": minBytes, "feedback_input_max_bytes": maxBytes,
			"native_cases": p.NativeCases, "conditions_passed": p.Conditions.Passed, "conditions_declared": p.Conditions.Declared,
			"external_calls": p.Search.Selection.ExternalCalls, "selected_choices": p.Search.Selection.Choices})
	}
	raw, err := json.MarshalIndent(map[string]any{"schema": "gooo/condition-feedback-observation/v1", "results": output,
		"scope": "one source, frozen model and finite cases; rank once versus two feedback rounds; no training, unseen-input or isolated-condition-causality claim"}, "", "  ")
	require(err == nil, "encode recount failed")
	fmt.Println(string(raw))
}
