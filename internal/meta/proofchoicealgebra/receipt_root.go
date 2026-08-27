package proofchoicealgebra

type Receipt struct {
	Schema               string         `json:"schema"`
	Decision             string         `json:"decision"`
	Reason               string         `json:"reason"`
	SubjectResolution    string         `json:"subject_resolution"`
	SourcePath           string         `json:"source_path"`
	SourceDigest         string         `json:"source_digest"`
	SemanticDigest       string         `json:"semantic_digest"`
	SourceReconstruction Reconstruction `json:"source_reconstruction"`
	Items                []Item         `json:"items"`
	Evidence             []Evidence     `json:"evidence"`
	Compositions         []Composition  `json:"compositions"`
	Transitions          []Transition   `json:"transitions"`
	Summary              Summary        `json:"summary"`
	Effects              Effects        `json:"effects"`
	Digest               string         `json:"digest"`
}
