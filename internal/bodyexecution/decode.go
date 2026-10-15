package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// DecodeGeneration retains the exact embedded parent bytes. The envelope is
// untrusted input: this checks shape, not source or model-execution authority.
func DecodeGeneration(data []byte) (bodycodegen.Result, []byte, error) {
	var result bodycodegen.Result
	if err := decode(data, &result, 2<<20); err != nil {
		return result, nil, err
	}
	var envelope struct {
		Report map[string]json.RawMessage `json:"report"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return result, nil, err
	}
	parent := envelope.Report["completeness_receipt"]
	common, err := completeness.Decode(parent)
	if err != nil {
		return result, nil, err
	}
	result.Report.CompletenessReceipt = common
	return result, append([]byte(nil), parent...), nil
}

func DecodePlan(data []byte) (pathplan.Document, error) {
	var plan pathplan.Document
	if err := decode(data, &plan, 256<<10); err != nil {
		return plan, err
	}
	return pathplan.DecodeDocument(data)
}

// DecodeSourcePlan expands recipes from the execution's original source while
// retaining the existing full-document decoder and its exact-field contract.
func DecodeSourcePlan(ctx context.Context, filename string, source []byte, activity string, data []byte) (pathplan.Document, error) {
	if ctx != nil && ctx.Err() == nil {
		assembly, err := bodycodegen.SourceAssembly(ctx, filename, source, activity)
		if err == nil && bodycodegen.IsSourceIRSearch(assembly) {
			if len(data) != 0 {
				return pathplan.Document{}, fmt.Errorf("source IR search owns its plan; omit the external path document")
			}
			return pathplan.Document{}, nil // Execute replays the source-owned IR search contract.
		}
		if err != nil || assembly != nil || len(data) == 0 {
			return bodycodegen.DecodeSourcePathDocument(ctx, filename, source, activity, data)
		}
	}
	if len(data) == 0 {
		return bodycodegen.DecodeSourcePathDocument(ctx, filename, source, activity, nil)
	}
	if len(data) > 256<<10 || !utf8.Valid(data) {
		return pathplan.Document{}, fmt.Errorf("invalid plan JSON size or UTF-8")
	}
	var header struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return pathplan.Document{}, err
	}
	if header.Schema == bodycodegen.SourcePathRecipeSchema {
		return bodycodegen.DecodeSourcePathDocument(ctx, filename, source, activity, data)
	}
	return DecodePlan(data)
}

func DecodeCases(data []byte) ([]pathplan.TestCase, error) {
	var suite struct {
		Schema string              `json:"schema"`
		Cases  []pathplan.TestCase `json:"cases"`
	}
	if err := decode(data, &suite, 32<<10); err != nil {
		return nil, err
	}
	if suite.Schema != "gooo/body-runtime-cases/v1" || len(suite.Cases) < 1 || len(suite.Cases) > 128 {
		return nil, fmt.Errorf("runtime suite requires its declared schema and 1..128 cases")
	}
	return suite.Cases, nil
}

func decode(data []byte, output any, limit int) error {
	if len(data) == 0 || len(data) > limit || !utf8.Valid(data) {
		return fmt.Errorf("invalid JSON size or UTF-8")
	}
	scan := json.NewDecoder(bytes.NewReader(data))
	scan.UseNumber()
	schema := wireSchema{fields: map[reflect.Type]map[string]reflect.Type{}}
	if err := schema.scanValue(scan, 0, reflect.TypeOf(output).Elem()); err != nil {
		return err
	}
	if _, err := scan.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	d.DisallowUnknownFields()
	return d.Decode(output)
}
