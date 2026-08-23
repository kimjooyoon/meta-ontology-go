package toolchainlsp

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
)

func observeDiagnostics(messages []rpcMessage) (map[string]observation, int) {
	all := make([]lsp.PublishDiagnosticsParams, 0, 3)
	for _, message := range messages {
		if message.Method != "textDocument/publishDiagnostics" {
			continue
		}
		var params lsp.PublishDiagnosticsParams
		if json.Unmarshal(message.Params, &params) == nil {
			all = append(all, params)
		}
	}
	clean, malformed, closed := false, false, false
	if len(all) == 3 {
		clean = all[0].URI == sessionURI && len(all[0].Diagnostics) == 0
		malformed = all[1].URI == sessionURI && len(all[1].Diagnostics) == 1 && all[1].Diagnostics[0].Code == "lex.unterminated-string"
		closed = all[2].URI == sessionURI && len(all[2].Diagnostics) == 0
	}
	return map[string]observation{
		"diagnostics-clean":        {"ZERO_DIAGNOSTICS", clean},
		"diagnostics-malformed":    {"LEX_UNTERMINATED_STRING", malformed},
		"close-clears-diagnostics": {"ZERO_DIAGNOSTICS", closed},
	}, boolCount(clean, malformed, closed)
}
