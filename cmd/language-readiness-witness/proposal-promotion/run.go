package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

type buildResult struct {
	promotion  *proposalpromotion.Receipt
	resolution *proposalpredecessor.ResolutionReceipt
}

func run(cfg config) error {
	if cfg.root == "" || cfg.repository == "" || cfg.currentHead == "" ||
		cfg.predecessorSHA == "" || cfg.token == "" {
		return fmt.Errorf("root, repository, current-head, predecessor-sha, and GITHUB_TOKEN are required")
	}
	if (cfg.output == "") == (cfg.check == "") {
		return fmt.Errorf("exactly one of output or check is required")
	}
	target := cfg.output
	if cfg.check != "" {
		target = cfg.check
	}
	if err := requireExternal(cfg.root, target); err != nil {
		return err
	}
	if cfg.check != "" {
		return checkReceipt(cfg.check, cfg.repository, cfg.currentHead, cfg.predecessorSHA)
	}
	store, err := openProposalObservationStore(cfg.observationCapture, cfg.observationReplay)
	if err != nil {
		return err
	}
	if store == nil {
		return fmt.Errorf("proposal observation capture or replay is required")
	}
	client, err := newProposalHTTPClient(cfg.apiURL, store)
	if err != nil {
		return err
	}
	result, err := build(cfg, client, store)
	if err != nil {
		var replayErr *proposalObservationReplayError
		if errors.As(err, &replayErr) {
			return fmt.Errorf("FAIL_CLOSED: proposal predecessor observation replay: %w", err)
		}
		return err
	}
	data, err := marshalResult(result)
	if err != nil {
		return err
	}
	if err := writeExclusive(cfg.output, data); err != nil {
		return err
	}
	if result.resolution != nil {
		fmt.Printf("proposal-predecessor: conformance=%s decision=%s resolution=%s stage=%s step=%s reason=%s promotion_authority=false readiness_delta=null digest=%s\n",
			result.resolution.Conformance, result.resolution.Decision, result.resolution.Resolution,
			result.resolution.Stage, result.resolution.Step, result.resolution.Reason,
			result.resolution.ReportDigest)
		return nil
	}
	receipt := *result.promotion
	fmt.Printf("proposal-promotion: decision=%s coordinates=%d/%d bps=%d writes=%d digest=%s\n",
		receipt.Decision, receipt.Summary.Satisfied, receipt.Summary.Total,
		receipt.Summary.ReadinessBPS, receipt.RepositoryWrites, receipt.ReportDigest)
	return nil
}

func build(cfg config, client *http.Client, store *proposalObservationStore) (buildResult, error) {
	collection, err := proposalpredecessor.Collect(
		context.Background(), client, cfg.apiURL, cfg.token,
		cfg.repository, cfg.predecessorSHA,
	)
	if err != nil {
		reason := proposalpredecessor.FailureReason(err)
		if !proposalpredecessor.KnownFailureReason(reason) {
			return buildResult{}, err
		}
		observationEvidence, evidenceErr := store.finalize()
		if evidenceErr != nil {
			return buildResult{}, evidenceErr
		}
		resolution, resolutionErr := proposalpredecessor.BuildResolution(
			cfg.repository, cfg.currentHead, cfg.predecessorSHA, reason, nil, observationEvidence,
		)
		if resolutionErr != nil {
			return buildResult{}, resolutionErr
		}
		return buildResult{resolution: &resolution}, nil
	}
	selection, contract, err := proposalpredecessor.Select(
		cfg.repository, cfg.currentHead, cfg.predecessorSHA, collection,
	)
	if err != nil {
		if !proposalpredecessor.KnownFailureReason(selection.Reason) {
			return buildResult{}, err
		}
		observationEvidence, evidenceErr := store.finalize()
		if evidenceErr != nil {
			return buildResult{}, evidenceErr
		}
		resolution, resolutionErr := proposalpredecessor.BuildResolution(
			cfg.repository, cfg.currentHead, cfg.predecessorSHA, selection.Reason, &selection, observationEvidence,
		)
		if resolutionErr != nil {
			return buildResult{}, resolutionErr
		}
		return buildResult{resolution: &resolution}, nil
	}
	observationEvidence, err := store.finalize()
	if err != nil {
		return buildResult{}, err
	}
	receipt, err := proposalpromotion.Build(cfg.currentHead, cfg.predecessorSHA, selection, contract, observationEvidence)
	if err != nil {
		return buildResult{}, err
	}
	return buildResult{promotion: &receipt}, nil
}

func marshalResult(result buildResult) ([]byte, error) {
	var value any
	switch {
	case result.promotion != nil && result.resolution == nil:
		value = result.promotion
	case result.promotion == nil && result.resolution != nil:
		value = result.resolution
	default:
		return nil, fmt.Errorf("proposal predecessor result has invalid alternatives")
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func checkReceipt(path, expectedRepository, expectedHead, expectedPredecessorSHA string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var header struct {
		Schema string `json:"schema"`
	}
	if err := decoder.Decode(&header); err != nil {
		return fmt.Errorf("FAIL_CLOSED: malformed proposal receipt: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("FAIL_CLOSED: proposal receipt has trailing content")
	}
	switch header.Schema {
	case proposalpromotion.Schema:
		var receipt proposalpromotion.Receipt
		if err := decodeStrict(data, &receipt); err != nil {
			return err
		}
		return proposalpromotion.Validate(receipt, expectedHead)
	case proposalpredecessor.ResolutionSchema:
		var receipt proposalpredecessor.ResolutionReceipt
		if err := decodeStrict(data, &receipt); err != nil {
			return err
		}
		return proposalpredecessor.ValidateResolution(receipt, expectedRepository, expectedHead, expectedPredecessorSHA)
	default:
		return fmt.Errorf("FAIL_CLOSED: unknown proposal receipt schema %q", header.Schema)
	}
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("FAIL_CLOSED: malformed proposal receipt: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("FAIL_CLOSED: proposal receipt has trailing content")
	}
	return nil
}
