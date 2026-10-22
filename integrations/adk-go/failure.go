package adkcheck

const (
	StageConfiguration = "configuration"
	StageInput         = "input"
	StageExecution     = "execution"
	StageOutput        = "output"

	CodeInvalidBinary       = "invalid_binary"
	CodeInvalidSourceRoot   = "invalid_source_root"
	CodeUnavailableRoot     = "unavailable_root"
	CodeInvalidPath         = "invalid_path"
	CodeInvalidContractPath = "invalid_contract_path"
	CodeInvalidQuery        = "invalid_query"
	CodeDeadline            = "deadline"
	CodeCancelled           = "cancelled"
	CodeStartFailed         = "start_failed"
	CodeTerminated          = "terminated"
	CodeOutputLimit         = "output_limit"
	CodeInvalidJSON         = "invalid_json"
)

// Failure separates an ADK boundary failure from a compiler JSON diagnostic.
// The public error text is stable and omits source paths and command output;
// hosts can inspect Cause through errors.Is or errors.As when needed.
type Failure struct {
	Stage string
	Code  string
	Cause error
}

func (f *Failure) Error() string {
	return "gooo-adk/" + f.Stage + "/" + f.Code
}

func (f *Failure) Unwrap() error { return f.Cause }

func failure(stage, code string, cause error) error {
	return &Failure{Stage: stage, Code: code, Cause: cause}
}
