package promotioncontinuity

import (
	"encoding/json"
	"fmt"
	"os"
)

type recoveryEnvelope struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`
	Resolution string `json:"resolution"`
	Mode       string `json:"mode"`
	Source struct {
		ExpectedHeadSHA string `json:"expected_head_sha"`
		Guard struct {
			Decision, Resolution string
		} `json:"guard"`
		Transformation struct {
			Decision                  string `json:"decision"`
			WriteBoundary             string `json:"write_boundary"`
			Effects                   int    `json:"effects"`
			SourceWorkspaceUnchanged  bool   `json:"source_workspace_unchanged"`
			PromotionAuthorized       bool   `json:"promotion_authorized"`
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
	var e recoveryEnvelope
	if err := json.Unmarshal(data, &e); err != nil {
		return RecoveryEvidence{}, fmt.Errorf("decode recovery: %w", err)
	}
	return RecoveryEvidence{
		Schema: e.Schema, FileSHA256: fileSHA256(data), ReportDigest: e.ReportDigest,
		HeadSHA: e.Source.ExpectedHeadSHA, Decision: e.Decision, Reason: e.Reason,
		Resolution: e.Resolution, Mode: e.Mode, GuardDecision: e.Source.Guard.Decision,
		GuardResolution: e.Source.Guard.Resolution, Satisfied: e.Summary.Satisfied,
		Total: e.Summary.Total, Unresolved: e.Summary.Unresolved,
		ReadinessBPS: e.Summary.ReadinessBPS, RecoveredFixedPoints: e.Summary.RecoveredFixedPoints,
		AuthorizedPromotions: e.Summary.AuthorizedPromotions,
		TransformationDecision: e.Source.Transformation.Decision,
		TransformationEffects: e.Source.Transformation.Effects,
		WriteBoundary: e.Source.Transformation.WriteBoundary,
		SourceWorkspaceUnchanged: e.Source.Transformation.SourceWorkspaceUnchanged,
		TransformationAuthorization: e.Source.Transformation.PromotionAuthorized,
		SourceRepositoryWrites: e.Source.RepositoryWrites,
		SummaryRepositoryWrites: e.Summary.RepositoryWrites,
		RepositoryWrites: e.RepositoryWrites,
		MutationAuthorized: e.RepositoryMutationAuthorized,
	}, nil
}
