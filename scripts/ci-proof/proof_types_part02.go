package main

type proofDigests struct {
	Source     string `json:"source_sha256"`
	Semantic   string `json:"semantic_sha256"`
	Provenance string `json:"provenance_sha256"`
	Projection string `json:"projection_sha256"`
	Build      string `json:"build_sha256"`
	Policy     string `json:"policy_sha256"`
	Schema     string `json:"schema_sha256"`
	Toolchain  string `json:"toolchain_sha256"`
	Target     string `json:"target_sha256"`
	Bundle     string `json:"bundle_sha256"`
}
