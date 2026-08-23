package transformationeffect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const splitGoEvaluationArtifactSchema = "gooo.meta.split-go-evaluation-artifact/v1"

// SplitGoEvaluationArtifact preserves the inputs and output required to replay
// receipt production without granting the source-splitting actor judge authority.
type SplitGoEvaluationArtifact struct {
	SchemaVersion        string                        `json:"schema_version"`
	ContractBytes        []byte                        `json:"contract_bytes"`
	EvidenceBytes        []byte                        `json:"evidence_bytes"`
	ReportBytes          []byte                        `json:"report_bytes"`
	RequiredIndicatorIDs []string                      `json:"required_indicator_ids"`
	ProofChoice          string                        `json:"proof_choice"`
	Receipts             []generation.IndicatorReceipt `json:"receipts"`
	Resolution           string                        `json:"resolution"`
	Reasons              []string                      `json:"reasons,omitempty"`
}

func EvaluateSplitGo(contractRaw, evidenceRaw []byte, requiredIDs []string, proofChoice string) (SplitGoEvaluationArtifact, error) {
	if err := validateSplitGoRequiredIDs(requiredIDs); err != nil {
		return SplitGoEvaluationArtifact{}, err
	}
	report, err := invokeSplitGoEvaluator(contractRaw, evidenceRaw)
	if err != nil {
		return SplitGoEvaluationArtifact{}, err
	}
	reportRaw, err := json.Marshal(report)
	if err != nil {
		return SplitGoEvaluationArtifact{}, fmt.Errorf("marshal SplitGo evaluator report: %w", err)
	}
	receipts, resolution, reasons, err := projectSplitGoReport(reportRaw, requiredIDs, proofChoice)
	if err != nil {
		return SplitGoEvaluationArtifact{}, err
	}
	return SplitGoEvaluationArtifact{
		SchemaVersion:        splitGoEvaluationArtifactSchema,
		ContractBytes:        bytes.Clone(contractRaw),
		EvidenceBytes:        bytes.Clone(evidenceRaw),
		ReportBytes:          reportRaw,
		RequiredIndicatorIDs: append([]string(nil), requiredIDs...),
		ProofChoice:          proofChoice,
		Receipts:             receipts,
		Resolution:           resolution,
		Reasons:              reasons,
	}, nil
}

func ValidateSplitGoEvaluation(artifact SplitGoEvaluationArtifact) error {
	if artifact.SchemaVersion != splitGoEvaluationArtifactSchema {
		return fmt.Errorf("unsupported SplitGo evaluation artifact schema %q", artifact.SchemaVersion)
	}
	replayed, err := EvaluateSplitGo(artifact.ContractBytes, artifact.EvidenceBytes, artifact.RequiredIndicatorIDs, artifact.ProofChoice)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(artifact, replayed) {
		return errors.New("SplitGo evaluation artifact does not match deterministic replay")
	}
	return nil
}
