package nativeobservation

// ToolchainObservation retains the original version check separately from
// current work. Reused describes this request, not a new process occurrence.
type ToolchainObservation struct {
	Schema            string             `json:"schema"`
	KeySHA256         string             `json:"key_sha256"`
	GoToolSHA256      string             `json:"go_tool_sha256"`
	GoVersion         string             `json:"go_version"`
	Reused            bool               `json:"reused"`
	BytesVerified     bool               `json:"bytes_verified"`
	BuildInfoVerified bool               `json:"build_info_verified"`
	SourceCheck       ProcessObservation `json:"source_check"`
	SourceOutput      []byte             `json:"source_output"`
	SourceCheckSHA256 string             `json:"source_check_sha256"`
}
