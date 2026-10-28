package bodycodegen

import (
	"bytes"
	"fmt"
	"go/types"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
	"github.com/kimjooyoon/meta-ontology-go/internal/valuejson"
)

type recordCaseDecoder struct {
	records []RecordType
	nodes   *int
}

func newRecordCaseDecoder(t types.Type, records []RecordType) recordCaseDecoder {
	d := recordCaseDecoder{records: records}
	if recordCaseTypeHasReferences(t) {
		d.nodes = new(int)
	}
	return d
}

func recordCaseTypeHasReferences(t types.Type) bool {
	if sequence, ok := t.(*types.Slice); ok {
		return recordBodyElementName(sequence.Elem()) != ""
	}
	structure, ok := t.Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for field := range structure.Fields() {
		kind, _, _ := recordBodyScalarKind(field.Type())
		if kind == recordBodyRecordKind || kind == recordBodyRecordSequenceKind {
			return true
		}
	}
	return false
}

func (d recordCaseDecoder) charge(n int) error {
	if d.nodes != nil {
		*d.nodes += n
		if *d.nodes > syntax.RecordValueSlotLimit {
			return fmt.Errorf("record value exceeds %d slots", syntax.RecordValueSlotLimit)
		}
	}
	return nil
}

func (d recordCaseDecoder) value(raw []byte, t types.Type) (any, error) {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		return nil, fmt.Errorf("null is outside the pure value profile")
	}
	if named, ok := t.(*types.Named); ok {
		return d.object(raw, named)
	}
	if sequence, ok := t.(*types.Slice); ok {
		return d.sequence(raw, sequence, 0)
	}
	switch t {
	case types.Typ[types.Int64]:
		return strconv.ParseInt(string(raw), 10, 64)
	case types.Typ[types.Bool]:
		if string(raw) == "true" {
			return true, nil
		}
		if string(raw) == "false" {
			return false, nil
		}
	case types.Typ[types.String]:
		return valuejson.DecodeText(raw, 1024)
	}
	return nil, fmt.Errorf("value does not match %s or exceeds its text bound", t)
}

func (d recordCaseDecoder) field(raw []byte, t types.Type, bound textbound.UTF8) (any, error) {
	if err := bound.Validate(); err != nil {
		return nil, err
	}
	if !bound.IsSet() {
		return d.value(raw, t)
	}
	if sequence, ok := t.(*types.Slice); ok {
		return d.sequence(raw, sequence, bound)
	}
	maximum, err := bound.Resolve(0)
	if err != nil || t != types.Typ[types.String] {
		return nil, fmt.Errorf("max_utf8_bytes requires a Text field")
	}
	return valuejson.DecodeText(raw, maximum)
}

func recordBodyElementName(t types.Type) string {
	if named, ok := t.(*types.Named); ok {
		if _, record := named.Underlying().(*types.Struct); record {
			return named.Obj().Name()
		}
	}
	return ""
}
