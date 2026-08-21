package coupling

import (
	"encoding/json"
)

type inputWire struct {
	Schema          string                   `json:"schema"`
	Config          Config                   `json:"config"`
	Registry        Registry                 `json:"registry"`
	Manifest        ChangeManifest           `json:"manifest"`
	Receipts        []CouplingReceipt        `json:"receipts"`
	InferencePath   pathWire                 `json:"inference_path"`
	ExternalReceipt *ExternalResourceReceipt `json:"external_receipt,omitempty"`
	WorkspaceRoot   string                   `json:"workspace_root,omitempty"`
}
type JSONError struct {
	Code   ReasonCode
	Detail string
}

func (e *JSONError) Error() string { return string(e.Code) + ": " + e.Detail }
func (input Input) MarshalJSON() ([]byte, error) {
	wire := inputWire{
		Schema: input.Schema, Config: input.Config, Registry: input.Registry, Manifest: input.Manifest,
		Receipts: input.Receipts, InferencePath: pathToWire(input.InferencePath),
		ExternalReceipt: input.ExternalReceipt, WorkspaceRoot: input.WorkspaceRoot,
	}
	return json.Marshal(wire)
}
func (input *Input) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeInput(data)
	if err != nil {
		return err
	}
	*input = decoded
	return nil
}
func DecodeInput(data []byte) (Input, error) {
	var wire inputWire
	if err := strictDecode(data, &wire); err != nil {
		return Input{}, &JSONError{Code: ReasonMalformedBinding, Detail: err.Error()}
	}
	return Input{
		Schema: wire.Schema, Config: wire.Config, Registry: wire.Registry, Manifest: wire.Manifest,
		Receipts: wire.Receipts, InferencePath: pathFromWire(wire.InferencePath),
		ExternalReceipt: wire.ExternalReceipt, WorkspaceRoot: wire.WorkspaceRoot,
	}, nil
}
func EvaluateJSON(data []byte, authority AuthorityContext) Result {
	input, err := DecodeInput(data)
	if err != nil {
		result := resultFor(StatusFailClosed, ReasonMalformedBinding, "JSON input", ObservationVector{})
		result.InputDigest = stableDigest(string(data) + authorityCanonical(authority))
		result.Digest = stableDigest(resultCanonical(result))
		return result
	}
	return Evaluate(input, authority)
}
func EncodeInput(input Input) ([]byte, error) { return json.Marshal(input) }
