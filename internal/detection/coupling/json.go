package coupling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func DecodeResult(data []byte) (Result, error) {
	var result Result
	if err := strictDecode(data, &result); err != nil {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: err.Error()}
	}
	if result.Schema != ResultSchemaV1 || (result.Status != StatusPass && result.Status != StatusFailClosed && result.Status != StatusUnknown) {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: "result schema or status"}
	}
	if err := validateResultShape(result); err != nil {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: err.Error()}
	}
	if !validDigest(result.InputDigest) || !validDigest(result.Digest) || stableDigest(resultCanonical(result)) != result.Digest {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: "result digest"}
	}
	return result, nil
}

func validateResultShape(result Result) error {
	validReasons := map[ReasonCode]struct{}{
		ReasonMalformedBinding: {}, ReasonRequiredInputMissing: {}, ReasonDuplicateSurface: {},
		ReasonSurfaceUnregistered: {}, ReasonDuplicateReceipt: {}, ReasonOrphanReceipt: {},
		ReasonStaleInput: {}, ReasonDigestMismatch: {}, ReasonSourceMapMismatch: {},
		ReasonContradictoryReceipt: {}, ReasonDeltaWithoutSource: {}, ReasonNoDeltaWithoutEquality: {},
		ReasonCandidateOnlyPath: {}, ReasonInferencePathMalformed: {}, ReasonMissingAuthorityPath: {},
		ReasonMissingVerification: {}, ReasonExternalReceiptMissing: {}, ReasonAuthorityInputSelfBound: {},
	}
	seenIDs := make(map[string]struct{}, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		if _, issue := normalizeID(id, "accepted surface ID"); issue != nil {
			return fmt.Errorf("invalid accepted surface ID")
		}
		if _, duplicate := seenIDs[id.String()]; duplicate {
			return fmt.Errorf("duplicate accepted surface ID")
		}
		seenIDs[id.String()] = struct{}{}
	}
	seenReasons := make(map[string]struct{}, len(result.Reasons))
	for _, reason := range result.Reasons {
		if _, ok := validReasons[reason.Code]; !ok || reason.Detail == "" {
			return fmt.Errorf("invalid result reason")
		}
		key := string(reason.Code) + "\x00" + reason.Detail
		if _, duplicate := seenReasons[key]; duplicate {
			return fmt.Errorf("duplicate result reason")
		}
		seenReasons[key] = struct{}{}
	}
	if result.Status == StatusPass {
		if len(result.Reasons) != 0 || result.FullSuiteRequired {
			return fmt.Errorf("PASS result has failure state")
		}
	} else if len(result.AcceptedSurfaceIDs) != 0 || len(result.Reasons) == 0 || !result.FullSuiteRequired {
		return fmt.Errorf("non-PASS result has invalid closure state")
	}
	for _, dimension := range []CountDimension{
		result.Observation.ChangedSurfaces, result.Observation.Receipts, result.Observation.InferenceRecords,
		result.Observation.InferencePaths, result.Observation.DeterministicWork, result.Observation.ResourceWork,
		result.Observation.CPU, result.Observation.Memory,
	} {
		if !dimension.Known && dimension.Value != 0 {
			return fmt.Errorf("unknown dimension carries a value")
		}
	}
	return nil
}

func EncodeResult(result Result) ([]byte, error) { return json.Marshal(result) }

func strictDecode(data []byte, destination any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, duplicate := seen[name]; duplicate {
				return fmt.Errorf("duplicate JSON key %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}
