package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

type constructionInputEvidence struct {
	Schema        string                                  `json:"schema"`
	ReceiptSHA256 string                                  `json:"receipt_sha256"`
	SourceSHA256  string                                  `json:"source_sha256"`
	Rows          []bodyexecution.ConstructionObservation `json:"rows"`
	ModelCalls    int                                     `json:"model_calls"`
	Scope         string                                  `json:"scope"`
}

func readPackageInputBytes(reader SourceReader, path string, construction bool) ([]byte, error) {
	if !construction {
		return readSource(reader, path)
	}
	raw, err := reader.ReadFile(path)
	if err == nil && len(raw) > 32<<20 {
		return nil, inputLimitError(32 << 20)
	}
	return raw, err
}

func packageConstructionInputs(ctx context.Context, raw []byte, target packageruntime.EntrySpec) ([]byte, *constructionInputEvidence, error) {
	var saved packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		return nil, nil, fmt.Errorf("decode construction receipt: %w", err)
	}
	if saved.Schema != "gooo/workspace-body-execution-receipt/v1" || saved.Result == nil || saved.Error != "" {
		return nil, nil, fmt.Errorf("construction input requires a saved workspace execution")
	}
	result := saved.Result
	rows, err := workspaceexecution.ObserveConstruction(ctx, *result)
	if err != nil {
		return nil, nil, err
	}
	input := struct {
		Schema string                                    `json:"schema"`
		Inputs []map[string]bodyexecution.AssemblyCounts `json:"inputs"`
	}{Schema: bodyexecution.CompositionInputsSchema}
	for i := range rows {
		row := &rows[i]
		if !row.ScoringCompleted {
			continue
		}
		index := len(input.Inputs)
		row.InputIndex = &index
		input.Inputs = append(input.Inputs, map[string]bodyexecution.AssemblyCounts{
			target.PackagePath + ":" + target.Activity: row.Counts,
		})
	}
	if len(input.Inputs) == 0 || len(input.Inputs) > 128 {
		return nil, nil, fmt.Errorf("construction input requires 1..128 scored candidate rows; got %d", len(input.Inputs))
	}
	encoded, err := json.Marshal(input)
	evidence := &constructionInputEvidence{Schema: "gooo/construction-input/v2", ReceiptSHA256: workspaceDigest(raw),
		SourceSHA256: result.SourceSHA256, Rows: rows,
		Scope: "recomputed whole training cases; attempt_prefix for search, scored_set for fills; unscored rows retain reasons without policy inputs; no model call or historical runtime/timing attestation"}
	return encoded, evidence, err
}
