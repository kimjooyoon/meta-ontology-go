package nativeobservation

// ArtifactObservation retains the original build of an owned executable.
// Reused describes this request; SourceBuild remains the original observation.
type ArtifactObservation struct {
	Schema             string             `json:"schema"`
	KeySHA256          string             `json:"key_sha256"`
	Reused             bool               `json:"reused"`
	ExecutableVerified bool               `json:"executable_verified"`
	SourceBuild        ProcessObservation `json:"source_build"`
	SourceBuildSHA256  string             `json:"source_build_sha256"`
	MissReason         string             `json:"miss_reason,omitempty"`
	WaitNS             int64              `json:"wait_ns"`
}
