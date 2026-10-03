package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func decodePathObservation(raw []byte) (*bodycodegen.PathObservationOptions, error) {
	if len(raw) == 0 || len(raw) > 4096 || decision.RejectDuplicateJSONKeys(raw) != nil {
		return nil, fmt.Errorf("path observation requires strict JSON of at most 4 KiB")
	}
	var wire struct {
		Schema                 string   `json:"schema"`
		Inputs                 []*int64 `json:"inputs"`
		MaxCandidates          int      `json:"max_candidates"`
		MaxRounds              int      `json:"max_rounds"`
		OracleActivity         string   `json:"oracle_activity"`
		ReuseProbeOutputs      bool     `json:"reuse_probe_outputs"`
		ResolveUniqueCandidate bool     `json:"resolve_unique_candidate"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&wire); err != nil {
		return nil, err
	}
	if d.Decode(&struct{}{}) != io.EOF || wire.Schema != "gooo/path-observation-request/v1" ||
		len(wire.Inputs) == 0 || len(wire.Inputs) > 32 || wire.MaxCandidates < 1 || wire.MaxCandidates > 64 ||
		wire.MaxRounds < 1 || wire.MaxRounds > 8 {
		return nil, fmt.Errorf("invalid bounded path observation request")
	}
	opts := &bodycodegen.PathObservationOptions{MaxCandidates: wire.MaxCandidates, MaxRounds: wire.MaxRounds,
		OracleActivity: wire.OracleActivity, ReuseProbeOutputs: wire.ReuseProbeOutputs,
		ResolveUniqueCandidate: wire.ResolveUniqueCandidate}
	for _, input := range wire.Inputs {
		if input == nil {
			return nil, fmt.Errorf("path probe input requires an explicit integer")
		}
		opts.Inputs = append(opts.Inputs, *input)
	}
	return opts, nil
}

func generateWithPathObservation(ctx context.Context, reader SourceReader, filename string, source []byte,
	activity string, document pathplan.Document, model, observationPath, diagnosisPath, ciPath string,
	options bodycodegen.TypedPathOptions) (bodycodegen.Result, error) {
	raw, err := reader.ReadFile(observationPath)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	options.Observation, err = decodePathObservation(raw)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	if diagnosisPath != "" {
		return generateWithPathDiagnosis(ctx, reader, filename, source, activity, document, model,
			diagnosisPath, ciPath, options)
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
