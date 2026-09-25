package artifactresolutionexperiment

type Coordinates struct {
	Satisfied   int `json:"satisfied"`
	Total       int `json:"total"`
	BasisPoints int `json:"basis_points"`
}

type ArtifactSummary struct {
	Manifest      int `json:"manifest"`
	Interface     int `json:"interface"`
	GoldenMatches int `json:"golden_matches"`
	Replays       int `json:"replays"`
}

type ResolutionSummary struct {
	ManifestDefinitions  int `json:"manifest_definitions"`
	InterfaceDefinitions int `json:"interface_definitions"`
	RegisteredEmitters   int `json:"registered_emitters"`
	CoherentOperations   int `json:"coherent_operations"`
}
