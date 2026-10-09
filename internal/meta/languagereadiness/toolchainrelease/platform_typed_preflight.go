package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func smokeTypedPreflight(binary string, input BuildInput) (string, error) {
	contextSHA := ""
	for _, unary := range []bool{true, false} {
		activity, source, name := "Pick", "mixed-model.gooo.fixture", "record"
		if unary {
			activity, source, name = "Choose", "unary.gooo.fixture", "unary"
		}
		source = filepath.Join(typedPathsRoot, source)
		raw, runErr := commandOutput(input.Root, nil, binary, "body-context", "--activity", activity,
			"--model", typedPathsModel, source)
		raw, err := retainLanguageCommandOutput(input, "typed-path-"+name+"-preflight", raw, runErr)
		if err != nil {
			return "", err
		}
		original, err := os.ReadFile(filepath.Join(input.Root, source))
		if err != nil {
			return "", err
		}
		if unary {
			err = validateTypedPreflight(raw, original)
		} else {
			contextSHA, err = validateTypedRecordPreflight(raw, original)
		}
		if err != nil {
			return "", fmt.Errorf("TOOLCHAIN_RELEASE_TYPED_PREFLIGHT %s: %w", name, err)
		}
	}
	return contextSHA, nil
}

type typedPreflightCounters struct {
	Predictions *int  `json:"model_predictions"`
	Tests       *int  `json:"candidate_tests"`
	Writes      *int  `json:"repository_writes"`
	Selected    *bool `json:"selected_emission"`
}

func typedPreflightZeros(raw []byte, typed bool) bool {
	var c typedPreflightCounters
	if json.Unmarshal(raw, &c) != nil || !jointSmokeInt(c.Predictions, 0) || !jointSmokeInt(c.Tests, 0) {
		return false
	}
	return !typed || (jointSmokeInt(c.Writes, 0) && jointSmokeBool(c.Selected, false))
}

func validateTypedPreflight(raw, source []byte) error {
	var r bodycodegen.TypedPathContextExport
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	c := r.ModelCompatibility
	if r.Schema != "gooo/compiler-path-model-input-export/v1" || !typedPreflightZeros(raw, true) ||
		r.OriginalSourceSHA256 != scalarPreflightDigest(source) || !r.SourceBinding.Equivalent ||
		r.SourceBinding.Decision != "PASS" || r.SourceBinding.SourceSemanticDigest == "" ||
		r.SourceBinding.SourceSemanticDigest != r.SourceBinding.GeneratedSemanticDigest ||
		c == nil || c.Schema != "gooo/path-model-compatibility/v1" ||
		c.Status != "DECLINED_TO_DETERMINISTIC" || c.Reason != "THREE_DECISION_COUNT_UNSUPPORTED" ||
		r.Context == nil || r.Context.Status != c.Status || r.Context.Reason != c.Reason ||
		r.Context.ActivityID != "callerpaths://activity/choose" || len(r.Inputs) != 0 || r.CompleteModelInput != nil {
		return fmt.Errorf("typed model decline, source binding or explicit zero counters differ")
	}
	d := r.Context.DeclaredInputs
	if d == nil || d.Decisions != 1 || d.Text == "" || d.Bytes != len(d.Text) ||
		d.SHA256 != scalarPreflightDigest([]byte(d.Text)) || r.Context.MetadataSHA != c.Model.MetadataSHA256 ||
		r.Context.FeatureVersion != c.Model.FeatureVersion || r.Context.ArithmeticVersion != c.Model.ArithmeticVersion {
		return fmt.Errorf("typed decline lost the original declared input or model identity")
	}
	return validateScalarModelIdentity(c.Model)
}

func validateTypedRecordPreflight(raw, source []byte) (string, error) {
	var r bodycodegen.RecordAssemblyContextExport
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	if r.Schema != "gooo/record-assembly-input-export/v1" || !typedPreflightZeros(raw, false) ||
		r.OriginalSourceSHA256 != scalarPreflightDigest(source) || r.ActivityID != "callerpaths://activity/pick" ||
		len(r.Choices) != 3 || r.Context == nil || r.Context.Status != "ENCODED" || r.Context.Reason != "" ||
		r.Context.Schema != "gooo/record-value-graph-context/v3" ||
		r.Context.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion ||
		r.Context.Text == "" || r.Context.SHA256 != scalarPreflightDigest([]byte(r.Context.Text)) {
		return "", fmt.Errorf("mixed record model source, input or explicit zero counters differ")
	}
	if err := validateScalarPreflightModel(r.ModelCompatibility); err != nil {
		return "", err
	}
	if _, err := jointdecision.DecodeRecordGraphThree(r.Context.Text); err != nil {
		return "", fmt.Errorf("mixed record input graph differs: %w", err)
	}
	return r.Context.SHA256, nil
}
