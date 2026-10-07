package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type constructionInputEvidence struct {
	Schema        string                                        `json:"schema"`
	ReceiptSHA256 string                                        `json:"receipt_sha256"`
	SourceSHA256  string                                        `json:"source_sha256"`
	Rows          []bodyexecution.RecordConstructionObservation `json:"rows"`
	ModelCalls    int                                           `json:"model_calls"`
	Scope         string                                        `json:"scope"`
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
	if result.Schema != "gooo/workspace-body-execution/v1" || len(result.BodyFills) != 0 {
		return nil, nil, fmt.Errorf("construction input currently supports record choices without preceding body fills")
	}
	source := []byte(result.Program.Source)
	if workspaceDigest(source) != result.SourceSHA256 {
		return nil, nil, fmt.Errorf("construction input source differs from the saved execution source")
	}
	rows, err := bodyexecution.ObserveRecordConstruction(ctx, source, result.Composition)
	if err != nil {
		return nil, nil, err
	}
	input := struct {
		Schema string                                    `json:"schema"`
		Inputs []map[string]bodyexecution.AssemblyCounts `json:"inputs"`
	}{Schema: bodyexecution.CompositionInputsSchema}
	for _, row := range rows {
		input.Inputs = append(input.Inputs, map[string]bodyexecution.AssemblyCounts{
			target.PackagePath + ":" + target.Activity: row.Counts,
		})
	}
	encoded, err := json.Marshal(input)
	evidence := &constructionInputEvidence{Schema: "gooo/record-construction-input/v1", ReceiptSHA256: workspaceDigest(raw),
		SourceSHA256: result.SourceSHA256, Rows: rows,
		Scope: "recomputed whole construction cases and candidate attempts; original declared budget capped by available candidates; no model call or historical native-runtime claim"}
	return encoded, evidence, err
}
