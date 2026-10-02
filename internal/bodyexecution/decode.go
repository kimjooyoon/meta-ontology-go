package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
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
	if err := scanValue(scan, 0); err != nil {
		return err
	}
	if _, err := scan.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	if err := exactKeys(data, reflect.TypeOf(output).Elem()); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	d.DisallowUnknownFields()
	return d.Decode(output)
}

func scanValue(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("JSON exceeds 64 nested levels")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	keys := map[string]bool{}
	for d.More() {
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return fmt.Errorf("duplicate or invalid JSON key")
			}
			keys[name] = true
		}
		if err := scanValue(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

// Exact spelling prevents encoding/json's case-insensitive field aliases from
// producing two interpretations of a hashed envelope. Opaque scope records and
// RawMessage retain their existing contracts; the shared receipt validates itself.
func exactKeys(data []byte, typ reflect.Type) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	if typ == reflect.TypeFor[json.RawMessage]() {
		return nil
	}
	if typ == reflect.TypeFor[[]byte]() {
		var encoded string
		return json.Unmarshal(data, &encoded) // JSON byte slices are base64 strings.
	}
	if typ.Kind() == reflect.Pointer {
		return exactKeys(data, typ.Elem())
	}
	switch typ.Kind() {
	case reflect.Struct:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(data, &object); err != nil {
			return err
		}
		fields := map[string]reflect.Type{}
		for f := range typ.Fields() {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name != "" && name != "-" {
				fields[name] = f.Type
			}
		}
		for name, value := range object {
			t, ok := fields[name]
			if !ok {
				return fmt.Errorf("unknown or aliased JSON field %q", name)
			}
			if err := exactKeys(value, t); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for _, value := range values {
			if err := exactKeys(value, typ.Elem()); err != nil {
				return err
			}
		}
	case reflect.Map:
		var values map[string]json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for _, value := range values {
			if err := exactKeys(value, typ.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
