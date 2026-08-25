package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

const ContractSchema = "gooo/language-example-experiment-contract/v1"

type Contract struct {
	Schema         string   `json:"schema"`
	ID             string   `json:"id"`
	ArtifactSchema string   `json:"artifact_schema"`
	EmitterKind    string   `json:"emitter_kind"`
	ExamplePath    string   `json:"example_path"`
	GoldenPath     string   `json:"golden_path"`
	Fixed          Fixed    `json:"fixed"`
	Limits         Limits   `json:"limits"`
	NotClaimed     []string `json:"not_claimed"`
}

type Fixed struct {
	SourceFiles              int `json:"source_files"`
	PrimaryArtifacts         int `json:"primary_artifacts"`
	DeterministicReplays     int `json:"deterministic_replays"`
	RegisteredEmitters       int `json:"registered_emitters"`
	ResourceSamples          int `json:"resource_samples"`
	UnknownEmitterRejections int `json:"unknown_emitter_rejections"`
	Indicators               int `json:"indicators"`
	NotClaimed               int `json:"not_claimed"`
}

type Limits struct {
	WallMS      int64 `json:"wall_ms"`
	RSSKiB      int64 `json:"rss_kib"`
	BinaryBytes int64 `json:"binary_bytes"`
}

type Golden struct {
	Package   artifactemit.Package   `json:"package"`
	Operation artifactemit.Operation `json:"operation"`
}
