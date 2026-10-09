package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func smokeScalarModelPreflight(binary string, input BuildInput, example languageSmokeCase) (string, error) {
	raw, runErr := commandOutput(input.Root, nil, binary, "body-context", "--activity", example.entry,
		"--model", filepath.Join(scalarIdentityRoot, "model", "model.json"), example.source)
	raw, err := retainLanguageCommandOutput(input, example.name+"-preflight", raw, runErr)
	if err != nil {
		return "", fmt.Errorf("TOOLCHAIN_RELEASE_SCALAR_PREFLIGHT %s: %w", example.name, err)
	}
	source, err := os.ReadFile(filepath.Join(input.Root, example.source))
	if err != nil {
		return "", err
	}
	return validateScalarModelPreflight(raw, source)
}

func validateScalarModelPreflight(raw, source []byte) (string, error) {
	var r struct {
		bodycodegen.RecordAssemblyContextExport
		Predictions *int `json:"model_predictions"`
		Tests       *int `json:"candidate_tests"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	if r.Schema != "gooo/record-assembly-input-export/v1" || r.ActivityID != scalarDescribeID ||
		r.OriginalSourceSHA256 != scalarPreflightDigest(source) || r.Predictions == nil || *r.Predictions != 0 ||
		r.Tests == nil || *r.Tests != 0 || r.ExpandedPlan != nil || r.ValueFlow != nil || len(r.Choices) != 3 {
		return "", fmt.Errorf("scalar model preflight source, scope or counters differ")
	}
	if err := validateScalarPreflightModel(r.ModelCompatibility); err != nil {
		return "", err
	}
	if err := validateScalarPreflightInput(r.Context, r.Choices); err != nil {
		return "", err
	}
	return r.Context.SHA256, nil
}

func validateScalarPreflightInput(c *bodycodegen.RecordOrdinalContext, choices []bodycodegen.RecordValueChoice) error {
	if c == nil || c.Schema != "gooo/record-value-graph-context/v3" || c.Status != "ENCODED" || c.Reason != "" ||
		c.FeatureVersion != jointdecision.RecordGraphSharedFeatureVersion || c.SHA256 != scalarPreflightDigest([]byte(c.Text)) {
		return fmt.Errorf("scalar model preflight input contract differs")
	}
	if _, err := jointdecision.DecodeRecordGraphThree(c.Text); err != nil {
		return fmt.Errorf("scalar model preflight graph: %w", err)
	}
	for i, field := range []string{"text", "active", "count"} {
		choice := choices[i]
		if choice.ID != field || choice.Field != field || choice.Occurrence != i || choice.Picked != "" ||
			choice.RecordID != "urn:gooo:scalar-identity:label" ||
			choice.FieldID != "urn:gooo:scalar-identity:label:"+field || choice.Intent == "" {
			return fmt.Errorf("scalar model preflight stable field identity differs")
		}
	}
	return nil
}

func validateScalarPreflightBinding(raw []byte, contextSHA string) error {
	var result struct {
		Composition bodyexecution.Composition `json:"composition"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}
	if len(result.Composition.Steps) != 1 {
		return fmt.Errorf("scalar model preflight assembly count differs")
	}
	r := result.Composition.Steps[0].Generation.Report.RecordAssembly
	if r == nil || r.Context == nil || contextSHA == "" || r.Context.SHA256 != contextSHA {
		return fmt.Errorf("scalar model preflight input differs from assembly or saved replay")
	}
	return nil
}

func scalarPreflightDigest(raw []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)) }
