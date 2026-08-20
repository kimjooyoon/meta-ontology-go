package main

const diagnosticSchemaVersion = "gooo/diagnostics/v1"

type cliDiagnostic struct {
	Severity string  `json:"severity"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	Span     cliSpan `json:"span"`
}
type cliSpan struct {
	File  string      `json:"file"`
	Start cliPosition `json:"start"`
	End   cliPosition `json:"end"`
}
type cliPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
type jsonReport struct {
	SchemaVersion            string          `json:"schema_version"`
	Command                  string          `json:"command"`
	Status                   string          `json:"status"`
	File                     string          `json:"file,omitempty"`
	Output                   string          `json:"output,omitempty"`
	Manifest                 string          `json:"manifest,omitempty"`
	PreviousGo               string          `json:"previous_go,omitempty"`
	ProtectedBytesEqual      *bool           `json:"protected_bytes_equal,omitempty"`
	SemanticHash             string          `json:"semantic_hash,omitempty"`
	OriginalSemanticHash     string          `json:"original_semantic_hash,omitempty"`
	RoundTrippedSemanticHash string          `json:"round_tripped_semantic_hash,omitempty"`
	Equivalent               *bool           `json:"equivalent,omitempty"`
	GetPut                   *bool           `json:"get_put,omitempty"`
	PutGet                   *bool           `json:"put_get,omitempty"`
	Diagnostics              []cliDiagnostic `json:"diagnostics"`

	Provenance *provenancePublishResponse `json:"provenance,omitempty"`
}

func parseJSONFlag(args []string) (clean []string, jsonMode bool) {
	clean = make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--json" {
			jsonMode = true
			continue
		}
		clean = append(clean, arg)
	}
	return clean, jsonMode
}
func newJSONReport(command, status, filename string, diagnostics []cliDiagnostic) jsonReport {
	if diagnostics == nil {
		diagnostics = []cliDiagnostic{}
	}
	return jsonReport{
		SchemaVersion: diagnosticSchemaVersion,
		Command:       command,
		Status:        status,
		File:          filename,
		Diagnostics:   diagnostics,
	}
}
