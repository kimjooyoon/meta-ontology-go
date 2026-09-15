package languageexampleexperiment

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

const ContractSchema = "gooo/language-example-experiment-contract/v2"

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
	ArtifactDigestChecks     int `json:"artifact_digest_checks"`
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

func CanonicalContract() Contract {
	return Contract{
		Schema: ContractSchema, ID: "billing-operation-manifest-v2",
		ArtifactSchema: artifactemit.OperationManifestSchema,
		EmitterKind:    artifactemit.OperationManifestKind,
		ExamplePath:    "examples/billing-package",
		GoldenPath:     "examples/billing-package/operation-manifest.golden.json",
		Fixed: Fixed{SourceFiles: 2, PrimaryArtifacts: 1, DeterministicReplays: 1,
			RegisteredEmitters: 3, ResourceSamples: 5, UnknownEmitterRejections: 1,
			ArtifactDigestChecks: 3, Indicators: 15, NotClaimed: 5},
		Limits: Limits{WallMS: 2000, RSSKiB: 131072, BinaryBytes: 33554432},
		NotClaimed: []string{"business correctness", "value-level computation", "production readiness",
			"performance beyond this runner and fixed sample set", "general-purpose code generation"},
	}
}

func ContractValid(value Contract) bool { return reflect.DeepEqual(value, CanonicalContract()) }
