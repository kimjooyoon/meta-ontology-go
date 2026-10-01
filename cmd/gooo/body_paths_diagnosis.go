package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func decodePathDiagnosis(raw []byte) (*bodycodegen.PathDiagnosisOptions, error) {
	if len(raw) == 0 || len(raw) > 1024 || decision.RejectDuplicateJSONKeys(raw) != nil {
		return nil, fmt.Errorf("path diagnosis requires strict JSON of at most 1 KiB")
	}
	var request struct {
		Schema        string  `json:"schema"`
		Inputs        []int64 `json:"inputs"`
		MaxCandidates int     `json:"max_candidates"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil || request.Schema != "gooo/path-diagnosis-request/v1" ||
		len(request.Inputs) == 0 || len(request.Inputs) > 32 || request.MaxCandidates < 1 || request.MaxCandidates > 64 {
		return nil, fmt.Errorf("path diagnosis requires 1..32 integer inputs and 1..64 candidates")
	}
	return &bodycodegen.PathDiagnosisOptions{Inputs: request.Inputs, MaxCandidates: request.MaxCandidates}, nil
}

func generateWithPathDiagnosis(ctx context.Context, reader SourceReader, filename string, source []byte,
	activity string, document pathplan.Document, model, diagnosisPath, ciPath string,
	options bodycodegen.TypedPathOptions) (bodycodegen.Result, error) {
	raw, err := reader.ReadFile(diagnosisPath)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	options.Diagnosis, err = decodePathDiagnosis(raw)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	if ciPath != "" {
		raw, err := reader.ReadFile(ciPath)
		if err != nil {
			return bodycodegen.Result{}, err
		}
		options.CI, err = decodePathCIHint(raw)
		if err != nil {
			return bodycodegen.Result{}, err
		}
	}
	return bodycodegen.GenerateWithTypedPathOptions(ctx, filename, source, activity, document, model, options)
}
