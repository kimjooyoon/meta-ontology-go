package transformationeffectverification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func loadBundle(opts Options) (bundle, error) {
	if opts.PlanPath == "" || opts.ExecutionPath == "" || opts.LedgerPath == "" ||
		opts.ReceiptsPath == "" || opts.ProvenancePath == "" || opts.PatchPath == "" {
		return bundle{}, unknownFailure("read-input", "BUNDLE_INPUT_MISSING", "restore-transformation-evidence")
	}
	var result bundle
	items := []struct {
		path string
		into any
	}{
		{opts.PlanPath, &result.Plan}, {opts.ExecutionPath, &result.Execution},
		{opts.LedgerPath, &result.Ledger}, {opts.ReceiptsPath, &result.Receipts},
		{opts.ProvenancePath, &result.Provenance}, {opts.PatchPath, &result.Patch},
	}
	for _, item := range items {
		if err := decodePath(item.path, item.into); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return bundle{}, unknownFailure("read-input", "BUNDLE_INPUT_MISSING", "restore-transformation-evidence")
			}
			return bundle{}, malformedFailure(item.path, err)
		}
	}
	if opts.RuntimePath != "" {
		if err := decodePath(opts.RuntimePath, &result.Runtime); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return bundle{}, unknownFailure("read-runtime", "RUNTIME_OBSERVATION_MISSING", "restore-runtime-evidence")
			}
			return bundle{}, malformedFailure(opts.RuntimePath, err)
		}
	}
	return result, nil
}

func decodePath(path string, target any) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("decode %s: trailing JSON", path)
	}
	return nil
}

func unknownFailure(step, reason, next string) *validationFailure {
	return &validationFailure{Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION",
		Stage: "verify-bundle", Step: step, Reason: reason, Unknown: "DIRECT_MISSING",
		Next: next, Blocked: []string{}}
}

func malformedFailure(path string, cause error) *validationFailure {
	return &validationFailure{Decision: "REFUTED", Resolution: "EXACT",
		Stage: "verify-bundle", Step: "decode-input", Reason: "BUNDLE_INPUT_MALFORMED",
		Next: "report-counterexample", Blocked: []string{}, FieldPath: path,
		Expected: "canonical JSON", Observed: cause.Error()}
}
