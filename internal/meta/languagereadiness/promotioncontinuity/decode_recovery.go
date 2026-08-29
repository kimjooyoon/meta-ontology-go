package promotioncontinuity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"
)

type recoveryEnvelope struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`
	Resolution string `json:"resolution"`
	Mode       string `json:"mode"`
	Source     struct {
		ExpectedHeadSHA string `json:"expected_head_sha"`
		Guard           struct {
			Decision                     string `json:"decision"`
			Reason                       string `json:"reason"`
			Resolution                   string `json:"resolution"`
			Satisfied                    int    `json:"satisfied"`
			Total                        int    `json:"total"`
			Unresolved                   int    `json:"unresolved"`
			RepositoryWrites             int    `json:"repository_writes"`
			RepositoryMutationAuthorized bool   `json:"repository_mutation_authorized"`
		} `json:"guard"`
		Transformation struct {
			HeadSHA                       string `json:"head_sha"`
			Decision                      string `json:"decision"`
			Reason                        string `json:"reason"`
			WorkspaceMode                 string `json:"workspace_mode"`
			WriteBoundary                 string `json:"write_boundary"`
			Effects                       int    `json:"effects"`
			AppliedEffects                int    `json:"applied_effects"`
			RefutedEffects                int    `json:"refuted_effects"`
			OperationOutcome              string `json:"operation_outcome"`
			ReceiptDecision               string `json:"receipt_decision"`
			ReceiptCount                  int    `json:"receipt_count"`
			FailureCount                  int    `json:"failure_count"`
			UnknownCount                  int    `json:"unknown_count"`
			DirectUnknownCount            int    `json:"direct_unknown_count"`
			DependencyBlockedUnknownCount int    `json:"dependency_blocked_unknown_count"`
			UnknownCausalDigest           string `json:"unknown_causal_digest"`
			SourceWorkspaceUnchanged      bool   `json:"source_workspace_unchanged"`
			PromotionAuthorized           bool   `json:"promotion_authorized"`
		} `json:"transformation"`
		RepositoryWrites int `json:"repository_writes"`
	} `json:"source"`
	Summary struct {
		Satisfied            int `json:"satisfied"`
		Total                int `json:"total"`
		Unresolved           int `json:"unresolved"`
		ReadinessBPS         int `json:"readiness_bps"`
		RecoveredFixedPoints int `json:"recovered_fixed_points"`
		AuthorizedPromotions int `json:"authorized_promotions"`
		RepositoryWrites     int `json:"repository_writes"`
	} `json:"summary"`
	RepositoryWrites             int    `json:"repository_writes"`
	RepositoryMutationAuthorized bool   `json:"repository_mutation_authorized"`
	ReportDigest                 string `json:"report_digest"`
}

func readRecovery(file string) (RecoveryEvidence, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return RecoveryEvidence{}, fmt.Errorf("read recovery: %w", err)
	}
	var typed rollbackfixedpoint.Report
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&typed); err != nil {
		return RecoveryEvidence{}, fmt.Errorf("decode recovery: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return RecoveryEvidence{}, fmt.Errorf("recovery has trailing data")
	}
	if err := rollbackfixedpoint.Validate(typed); err != nil {
		return RecoveryEvidence{}, fmt.Errorf("validate recovery: %w", err)
	}
	var e recoveryEnvelope
	if err := json.Unmarshal(data, &e); err != nil {
		return RecoveryEvidence{}, fmt.Errorf("decode recovery: %w", err)
	}
	value := RecoveryEvidence{
		Schema: e.Schema, FileSHA256: fileSHA256(data), ReportDigest: e.ReportDigest,
		HeadSHA: e.Source.ExpectedHeadSHA, Decision: e.Decision, Reason: e.Reason,
		Resolution: e.Resolution, Mode: e.Mode, GuardDecision: e.Source.Guard.Decision,
		GuardReason: e.Source.Guard.Reason, GuardResolution: e.Source.Guard.Resolution,
		GuardSatisfied: e.Source.Guard.Satisfied, GuardTotal: e.Source.Guard.Total,
		GuardUnresolved:         e.Source.Guard.Unresolved,
		GuardRepositoryWrites:   e.Source.Guard.RepositoryWrites,
		GuardMutationAuthorized: e.Source.Guard.RepositoryMutationAuthorized,
		Satisfied:               e.Summary.Satisfied,
		Total:                   e.Summary.Total, Unresolved: e.Summary.Unresolved,
		ReadinessBPS: e.Summary.ReadinessBPS, RecoveredFixedPoints: e.Summary.RecoveredFixedPoints,
		AuthorizedPromotions:                        e.Summary.AuthorizedPromotions,
		TransformationHeadSHA:                       e.Source.Transformation.HeadSHA,
		TransformationDecision:                      e.Source.Transformation.Decision,
		TransformationReason:                        e.Source.Transformation.Reason,
		TransformationWorkspaceMode:                 e.Source.Transformation.WorkspaceMode,
		TransformationEffects:                       e.Source.Transformation.Effects,
		TransformationAppliedEffects:                e.Source.Transformation.AppliedEffects,
		TransformationRefutedEffects:                e.Source.Transformation.RefutedEffects,
		TransformationOperationOutcome:              e.Source.Transformation.OperationOutcome,
		TransformationReceiptDecision:               e.Source.Transformation.ReceiptDecision,
		TransformationReceiptCount:                  e.Source.Transformation.ReceiptCount,
		TransformationFailureCount:                  e.Source.Transformation.FailureCount,
		TransformationUnknownCount:                  e.Source.Transformation.UnknownCount,
		TransformationDirectUnknownCount:            e.Source.Transformation.DirectUnknownCount,
		TransformationDependencyBlockedUnknownCount: e.Source.Transformation.DependencyBlockedUnknownCount,
		TransformationUnknownCausalDigest:           e.Source.Transformation.UnknownCausalDigest,
		WriteBoundary:                               e.Source.Transformation.WriteBoundary,
		SourceWorkspaceUnchanged:                    e.Source.Transformation.SourceWorkspaceUnchanged,
		TransformationAuthorization:                 e.Source.Transformation.PromotionAuthorized,
		SourceRepositoryWrites:                      e.Source.RepositoryWrites,
		SummaryRepositoryWrites:                     e.Summary.RepositoryWrites,
		RepositoryWrites:                            e.RepositoryWrites,
		MutationAuthorized:                          e.RepositoryMutationAuthorized,
	}
	value.TransformationCausalBindingDigest = causalBindingDigest(value)
	return value, nil
}
