package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

type Summary struct {
	Coordinates     Coordinates           `json:"coordinates"`
	Value           ValueSummary          `json:"value"`
	Compiler        CompilerSummary       `json:"compiler"`
	Resources       ResourceSummary       `json:"resources"`
	Counterexamples CounterexampleSummary `json:"counterexamples"`
	Effects         artifactemit.Effects  `json:"effects"`
	NotClaimed      int                   `json:"not_claimed"`
	Unknowns        int                   `json:"unknowns"`
}

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type ValueSummary struct {
	PrimaryArtifacts     int `json:"primary_artifacts"`
	GoldenMatches        int `json:"golden_matches"`
	DeterministicReplays int `json:"deterministic_replays"`
}

type CompilerSummary struct {
	SourceFiles        int `json:"source_files"`
	GoooFiles          int `json:"gooo_files"`
	GoFiles            int `json:"go_files"`
	GoooDefinitionBPS  int `json:"gooo_definition_basis_points"`
	RegisteredEmitters int `json:"registered_emitters"`
}

type ResourceSummary struct {
	Samples          int   `json:"samples"`
	MaxWallMS        int64 `json:"max_wall_ms"`
	MaxRSSKiB        int64 `json:"max_rss_kib"`
	BinaryBytes      int64 `json:"binary_bytes"`
	WallViolations   int   `json:"wall_violations"`
	RSSViolations    int   `json:"rss_violations"`
	BinaryViolations int   `json:"binary_violations"`
}

type CounterexampleSummary struct {
	UnknownEmitterRejections int `json:"unknown_emitter_rejections"`
}
