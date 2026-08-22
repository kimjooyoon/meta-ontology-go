package artifact

import (
	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"
)

const Schema = "gooo/language-readiness-artifact/v1"

type Receipt struct {
	Schema                  string                 `json:"schema"`
	HeadSHA                 string                 `json:"head_sha"`
	ProposalPromotionDigest string                 `json:"proposal_promotion_digest,omitempty"`
	Snapshot                readiness.Snapshot     `json:"snapshot"`
	TransitionInput         improvement.Snapshot   `json:"transition_input"`
	FixedPoint              improvement.Transition `json:"fixed_point"`
	ArtifactDigest          string                 `json:"artifact_digest"`
}
