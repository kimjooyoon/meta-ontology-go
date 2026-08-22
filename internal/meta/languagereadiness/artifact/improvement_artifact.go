package artifact

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/improvement"

const (
	ImprovementArtifactSchema = "gooo/language-readiness-first-improvement/v1"
	improvementProducer       = "artifact.BuildImprovement"
	improvementConsumer       = "self-improvement-cycle"
	improvementOperation      = "prove-quantified-improvement"
)

type ImprovementArtifact struct {
	Schema               string                 `json:"schema"`
	Producer             string                 `json:"producer"`
	Consumer             string                 `json:"consumer"`
	MetaOperation        string                 `json:"meta_operation"`
	Baseline             BaselineReference      `json:"baseline"`
	HeadSHA              string                 `json:"head_sha"`
	BeforeArtifactDigest string                 `json:"before_artifact_digest"`
	AfterArtifactDigest  string                 `json:"after_artifact_digest"`
	BeforeSnapshotDigest string                 `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string                 `json:"after_snapshot_digest"`
	Transition           improvement.Transition `json:"transition"`
	RepositoryWrites     int                    `json:"repository_writes"`
	ArtifactDigest       string                 `json:"artifact_digest"`
}
