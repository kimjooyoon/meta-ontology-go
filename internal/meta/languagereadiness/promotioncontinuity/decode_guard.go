package promotioncontinuity

import (
	"encoding/json"
	"fmt"
	"os"
)

type guardEnvelope struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Reason     string `json:"reason"`
	Resolution string `json:"resolution"`
	Source     struct {
		CurrentHeadSHA string `json:"current_head_sha"`
	} `json:"source"`
	Summary struct {
		Satisfied                   int  `json:"satisfied"`
		Total                       int  `json:"total"`
		Unresolved                  int  `json:"unresolved"`
		RepositoryWrites            int  `json:"repository_writes"`
		ReadinessPromotionAuthorized bool `json:"readiness_promotion_authorized"`
		RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
	} `json:"summary"`
	ReportDigest string `json:"report_digest"`
}

func readGuard(file string) (GuardEvidence, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return GuardEvidence{}, fmt.Errorf("read guard: %w", err)
	}
	var envelope guardEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return GuardEvidence{}, fmt.Errorf("decode guard: %w", err)
	}
	return GuardEvidence{
		Schema: envelope.Schema, FileSHA256: fileSHA256(data),
		ReportDigest: envelope.ReportDigest, HeadSHA: envelope.Source.CurrentHeadSHA,
		Decision: envelope.Decision, Reason: envelope.Reason, Resolution: envelope.Resolution,
		Satisfied: envelope.Summary.Satisfied, Total: envelope.Summary.Total,
		Unresolved: envelope.Summary.Unresolved,
		RepositoryWrites: envelope.Summary.RepositoryWrites,
		PromotionAuthorized: envelope.Summary.ReadinessPromotionAuthorized,
		MutationAuthorized: envelope.Summary.RepositoryMutationAuthorized,
	}, nil
}
