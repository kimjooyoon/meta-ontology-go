package valueexecution

import (
	"context"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const recordForwardProgram = "record.forward:v1"
const RecordTransportScope = "RECORD_TRANSPORT_ONLY"

type RecordField struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type recordProgram struct {
	authority resultAuthority
	fields    []RecordField
}

// RecordPlan is issued by CompileRecordPlan. Its private programs, field
// schemas, and edges are not reconstructible from serialized report data.
type RecordPlan struct {
	SourceDigest        string
	SemanticFingerprint string
	sourceDigest        string
	fingerprint         string
	programs            map[string]recordProgram
	incoming            map[string]string
	order               []string
}

// CompileRecordPlan supports source-declared required, single string fields.
// It transports data, including claims, without judging or authorizing claims.
func CompileRecordPlan(filename string, source []byte) (RecordPlan, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(filename, string(source), syntax.EntityFieldsV1Support())
	if file == nil || diagnostics.HasErrors() {
		return RecordPlan{}, failAt(ReasonSourceParseFailed, "PARSE", "parse-record-source", "record source syntax is invalid")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, bidir.EntityFieldsV1Support())
	if err != nil {
		return RecordPlan{}, failAt(ReasonSemanticBindingFailed, "LOWER", "lower-record-source", err.Error())
	}
	plan := RecordPlan{SourceDigest: digestBytes(source), SemanticFingerprint: ir.StableHash(),
		sourceDigest: digestBytes(source), fingerprint: ir.StableHash(), programs: map[string]recordProgram{}}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity {
			continue
		}
		program, err := compileRecordProgram(ir, node.Name, plan.sourceDigest)
		if err != nil {
			return RecordPlan{}, err
		}
		if _, duplicate := plan.programs[node.Name]; duplicate {
			return RecordPlan{}, failAt(ReasonPlanInvalid, "PLAN", "duplicate-record-activity", node.Name)
		}
		plan.programs[node.Name] = program
	}
	if len(plan.programs) == 0 {
		return RecordPlan{}, failAt(ReasonPlanInvalid, "PLAN", "require-record-activity", "source has no activities")
	}
	plan.incoming, err = compileRecordBindings(ir, plan.programs)
	if err != nil {
		return RecordPlan{}, err
	}
	plan.order, err = recordExecutionOrder(plan.programs, plan.incoming)
	if err != nil {
		return RecordPlan{}, err
	}
	return plan, nil
}

func compileRecordProgram(ir semantic.IR, name, sourceDigest string) (recordProgram, error) {
	node, found := ir.Graph.NodeByName(ir.Namespace, name)
	if !found || node.ValueProgram != recordForwardProgram {
		return recordProgram{}, failAt(ReasonProgramUnknown, "RESOLVE", "resolve-record-operation", name)
	}
	var inputName, outputName, inputID, outputID string
	inputs, outputs := 0, 0
	for _, entity := range ir.Graph.Nodes() {
		if entity.Kind != semantic.Entity {
			continue
		}
		if ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: entity.ID}) {
			inputs++
			inputName, inputID = entity.Name, entity.ID.String()
		}
		if ir.Graph.HasFact(semantic.FactKey{Subject: entity.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}) {
			outputs++
			outputName, outputID = entity.Name, entity.ID.String()
		}
	}
	if inputs != 1 || outputs != 1 || inputID != outputID || inputName != outputName {
		return recordProgram{}, failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "bind-record-signature", name)
	}
	fields, err := compileRecordFields(ir, inputName)
	if err != nil {
		return recordProgram{}, err
	}
	specDigest := digestValue([]any{recordForwardProgram, inputID, fields, RecordTransportScope, OperationAuthority{}})
	authority := resultAuthority{
		activityID: node.ID.String(), activityName: name, outputEntityID: outputID, outputEntityName: outputName,
		sourceDigest: sourceDigest, semanticFingerprint: ir.StableHash(),
		valueProgramDigest: digestBytes([]byte(node.ValueProgram)), modelProgramDigest: digestBytes([]byte(node.ValueProgram)),
		operationDigest: digestValue([]string{node.ID.String(), specDigest}), operationSpecDigest: specDigest,
	}
	return recordProgram{authority: authority, fields: fields}, nil
}

func compileRecordFields(ir semantic.IR, name string) ([]RecordField, error) {
	entity, found := ir.Graph.NodeByName(ir.Namespace, name)
	if !found || len(entity.Fields) == 0 {
		return nil, failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "require-record-fields", name)
	}
	fields := make([]RecordField, 0, len(entity.Fields))
	seen := map[string]bool{}
	for _, field := range entity.Fields {
		if field.Presence != semantic.Required || field.Cardinality != semantic.One ||
			field.TypeRef.ID != semantic.BuiltinStringTypeID || field.ID.String() == "" || seen[field.Name] {
			return nil, failAt(ReasonSignatureTypeMismatch, "TYPECHECK", "require-single-string-record-fields", name+"."+field.Name)
		}
		seen[field.Name] = true
		fields = append(fields, RecordField{Name: field.Name, ID: field.ID.String()})
	}
	slices.SortFunc(fields, func(left, right RecordField) int { return strings.Compare(left.Name, right.Name) })
	return fields, nil
}
