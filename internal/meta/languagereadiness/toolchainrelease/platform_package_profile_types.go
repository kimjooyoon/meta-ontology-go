package toolchainrelease

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

type packageProfile struct {
	Name, Root, Package, Activity, ID string
	Total                             int
}

type packageProfileCase struct {
	Inputs   map[string]json.RawMessage
	Expected map[string]json.RawMessage
}

type packageProfileReceipt struct {
	Schema, Decision, Error string
	Manifest                string `json:"manifest_digest"`
	Cases                   string `json:"cases_digest"`
	From                    string `json:"replayed_from_sha256"`
	Result                  struct {
		SourceSHA   string `json:"execution_source_sha256"`
		Program     workspaceexecution.Program
		Composition struct {
			SHA string `json:"generated_sha256"`
		}
		Replay *struct {
			Calls *int `json:"model_calls"`
		}
		Runtime packageProfileRuntime
		Fills   []packageProfileFill `json:"body_fills"`
	}
}
