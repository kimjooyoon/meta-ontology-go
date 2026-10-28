package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func encoded(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func normalizeObservation(side map[string]any, modelDigests map[string]string, dataOrigin, executionOrigin string) map[string]any {
	request := side["request"].(map[string]any)
	comparable := make(map[string]any, len(request)-1)
	for k, v := range request {
		if k != "authorization_model_id" {
			comparable[k] = v
		}
	}
	allowed := false
	if side["response_present"].(bool) {
		allowed = side["response"].(map[string]any)["allowed"].(bool)
	}
	errorCode, _ := side["grpc_code"].(string)
	return map[string]any{
		"recorded": true, "invoked": side["invoked"], "completed": side["completed"],
		"response_present": side["response_present"], "error_present": side["error_present"],
		"allowed": allowed, "error_code": errorCode,
		"request_input_digest":         digest(encoded(comparable)),
		"request_serialization_digest": side["request_sha256"],
		"model_id":                     side["model_id"], "request_model_id": request["authorization_model_id"],
		"model_source_digest": modelDigests[side["model_id"].(string)],
		"data_digest":         dataOrigin, "execution_scope_digest": executionOrigin,
	}
}

func readSnapshot(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil || !strings.HasSuffix(path, ".gz") {
		return raw, err
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(io.LimitReader(reader, 1<<20))
}

func main() {
	if len(os.Args) != 3 {
		panic("usage: normalize raw-observations.json normalized.json")
	}
	raw, err := readSnapshot(os.Args[1])
	if err != nil {
		panic(err)
	}
	if digest(raw) != "sha256:91d0b0c3358570e1745eed48c9376da2df22b60e7c4d9cbdbb736743d0309907" {
		panic("observation fixture digest mismatch")
	}
	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		panic(err)
	}
	models := snapshot["models"].([]any)
	modelDigests := map[string]string{}
	for _, item := range models {
		model := item.(map[string]any)
		if digest([]byte(model["dsl"].(string))) != model["dsl_sha256"].(string) {
			panic("model source digest mismatch")
		}
		modelDigests[model["id"].(string)] = model["dsl_sha256"].(string)
	}
	beforeModel := models[0].(map[string]any)["id"].(string)
	afterModel := models[1].(map[string]any)["id"].(string)
	write := snapshot["stored_tuple_write_request"].(map[string]any)
	dataOrigin := digest(encoded(map[string]any{"store_id": write["store_id"], "writes": write["writes"], "deletes": write["deletes"]}))
	executionOrigin := digest(encoded(map[string]any{"recorded_scope": snapshot["execution"], "at_utc": snapshot["at_utc"], "snapshot": digest(raw)}))
	pairs := []any{}
	for _, item := range snapshot["observations"].([]any) {
		observation := item.(map[string]any)
		pair := map[string]any{"case_id": observation["case"], "origin_kind": "ACTUAL_API_OBSERVATION",
			"scope_qualified": snapshot["no_tuple_writes_during_checks"], "before_model_id": beforeModel, "after_model_id": afterModel}
		pair["before"] = normalizeObservation(observation["old"].(map[string]any), modelDigests, dataOrigin, executionOrigin)
		pair["after"] = normalizeObservation(observation["new"].(map[string]any), modelDigests, dataOrigin, executionOrigin)
		pairs = append(pairs, pair)
	}
	result := map[string]any{"raw_observation_sha256": digest(raw), "normalizer_go_version": runtime.Version(), "pairs": pairs}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], append(output, '\n'), 0600); err != nil {
		panic(err)
	}
	fmt.Printf("normalized %d observed pairs; no OpenFGA Check calls\n", len(pairs))
}
