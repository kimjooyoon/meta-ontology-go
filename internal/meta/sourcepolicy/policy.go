package sourcepolicy

import "fmt"

const (
	Schema                    = "gooo/source-policy/v1"
	DefaultMaxFileLines       = 75
	DefaultMaxFunctionLines   = 75
	DefaultMaxDirectoryInputs = 10
)

// Policy is the single foundational source-structure policy used by metrics,
// refactoring tools, and CI verification.
type Policy struct {
	Schema                        string `json:"schema"`
	MaxFileLines                  int    `json:"max_file_lines"`
	MaxFunctionLines              int    `json:"max_function_lines"`
	MaxDirectDirectoryIn          int    `json:"max_direct_directory_entries"`
	RequireHomogeneousDirectories bool   `json:"require_homogeneous_directories"`
	ExemptProjectRootTopology     bool   `json:"exempt_project_root_topology"`
}

// Default returns the repository policy. Consumers derive indicators from it;
// they do not copy its values.
func Default() Policy {
	return Policy{
		Schema:                        Schema,
		MaxFileLines:                  DefaultMaxFileLines,
		MaxFunctionLines:              DefaultMaxFunctionLines,
		MaxDirectDirectoryIn:          DefaultMaxDirectoryInputs,
		RequireHomogeneousDirectories: true,
		ExemptProjectRootTopology:     true,
	}
}

// Validate rejects incomplete policy foundations.
func (p Policy) Validate() error {
	if p.Schema == "" {
		p.Schema = Schema
	}
	if p.MaxFileLines <= 0 || p.MaxFunctionLines <= 0 {
		return fmt.Errorf("source policy line limits must be positive")
	}
	if p.MaxDirectDirectoryIn < 0 {
		return fmt.Errorf("source policy directory limit must not be negative")
	}
	return nil
}
