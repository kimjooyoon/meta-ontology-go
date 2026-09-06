package generation

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// CallbackPreviewFieldEvidence is the exact field schema consumed by the
// native preview adapter. Values are intentionally strings because the
// EntityFields V1 semantic registry currently exposes only the builtin string
// type; native preview evidence still carries counters as integers.
type CallbackPreviewFieldEvidence struct {
	Entity      string `json:"entity"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	Presence    string `json:"presence"`
	Cardinality string `json:"cardinality"`
	TypeID      string `json:"type_id"`
}

type CallbackPreviewActivityEvidence struct {
	Name                string `json:"name"`
	InputEntity         string `json:"input_entity"`
	OutputEntity        string `json:"output_entity"`
	ValueProgram        string `json:"value_program"`
	UsedInputFact       bool   `json:"used_input_fact"`
	GeneratedOutputFact bool   `json:"generated_output_fact"`
}

type CallbackPreviewBindingEvidence struct {
	ProducerActivity string `json:"producer_activity"`
	ProducerPort     string `json:"producer_port"`
	ConsumerActivity string `json:"consumer_activity"`
	ConsumerPort     string `json:"consumer_port"`
	Entity           string `json:"entity"`
}

// CallbackPreviewContractEvidence is a preview-only contract. It has no
// OperationResult output and cannot authorize repository writes.
type CallbackPreviewContractEvidence struct {
	SourceDigest    string                            `json:"source_digest"`
	SemanticDigest  string                            `json:"semantic_digest"`
	InputEntity     string                            `json:"input_entity"`
	CandidateEntity string                            `json:"candidate_entity"`
	CapturesEntity  string                            `json:"captures_entity"`
	EffectsEntity   string                            `json:"effects_entity"`
	EvidenceEntity  string                            `json:"evidence_entity"`
	Fields          []CallbackPreviewFieldEvidence    `json:"fields"`
	Activities      []CallbackPreviewActivityEvidence `json:"activities"`
	Bindings        []CallbackPreviewBindingEvidence  `json:"bindings"`
}

const CallbackPreviewListCodecPrefix = "json-array:v1:"

type CallbackPreviewFieldValue struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CallbackPreviewRecord struct {
	Entity string                      `json:"entity"`
	Fields []CallbackPreviewFieldValue `json:"fields"`
}

func EncodeCallbackPreviewList(values []string) (string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return CallbackPreviewListCodecPrefix + string(raw), nil
}

func ValidateCallbackPreviewList(encoded string) error {
	if len(encoded) < len(CallbackPreviewListCodecPrefix) || encoded[:len(CallbackPreviewListCodecPrefix)] != CallbackPreviewListCodecPrefix {
		return fmt.Errorf("callback preview list does not use %s", CallbackPreviewListCodecPrefix)
	}
	var values []string
	if err := json.Unmarshal([]byte(encoded[len(CallbackPreviewListCodecPrefix):]), &values); err != nil {
		return fmt.Errorf("decode callback preview list: %w", err)
	}
	if values == nil {
		return fmt.Errorf("callback preview list must be a JSON array")
	}
	canonical, err := EncodeCallbackPreviewList(values)
	if err != nil {
		return err
	}
	if canonical != encoded {
		return fmt.Errorf("callback preview list is not canonical")
	}
	return nil
}

func EncodeCallbackPreviewNestedList(values [][]string) (string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return CallbackPreviewListCodecPrefix + string(raw), nil
}

func ValidateCallbackPreviewNestedList(encoded string) error {
	if len(encoded) < len(CallbackPreviewListCodecPrefix) || encoded[:len(CallbackPreviewListCodecPrefix)] != CallbackPreviewListCodecPrefix {
		return fmt.Errorf("callback preview nested list does not use %s", CallbackPreviewListCodecPrefix)
	}
	var values [][]string
	if err := json.Unmarshal([]byte(encoded[len(CallbackPreviewListCodecPrefix):]), &values); err != nil {
		return fmt.Errorf("decode callback preview nested list: %w", err)
	}
	if values == nil {
		return fmt.Errorf("callback preview nested list must be a JSON array")
	}
	for _, value := range values {
		if value == nil {
			return fmt.Errorf("callback preview nested list entries must be JSON arrays")
		}
	}
	canonical, err := EncodeCallbackPreviewNestedList(values)
	if err != nil {
		return err
	}
	if canonical != encoded {
		return fmt.Errorf("callback preview nested list is not canonical")
	}
	return nil
}

func (contract CallbackPreviewContractEvidence) BuildCallbackPreviewRecord(entity string, values map[string]string) (CallbackPreviewRecord, error) {
	fields := make([]CallbackPreviewFieldValue, 0)
	for _, field := range contract.Fields {
		if field.Entity != entity {
			continue
		}
		value, ok := values[field.Name]
		if !ok {
			return CallbackPreviewRecord{}, fmt.Errorf("callback preview record %s is missing field %s", entity, field.Name)
		}
		fields = append(fields, CallbackPreviewFieldValue{ID: field.ID, Name: field.Name, Value: value})
	}
	if len(fields) == 0 {
		return CallbackPreviewRecord{}, fmt.Errorf("callback preview record entity %q is not in the contract", entity)
	}
	if len(values) != len(fields) {
		return CallbackPreviewRecord{}, fmt.Errorf("callback preview record %s has fields outside the contract", entity)
	}
	sort.Slice(fields, func(left, right int) bool { return fields[left].ID < fields[right].ID })
	return CallbackPreviewRecord{Entity: entity, Fields: fields}, nil
}

func (contract CallbackPreviewContractEvidence) ValidateCallbackPreviewFlow(records []CallbackPreviewRecord) error {
	expected := []string{contract.InputEntity, contract.CandidateEntity, contract.CapturesEntity, contract.EffectsEntity, contract.EvidenceEntity}
	if len(records) != len(expected) {
		return fmt.Errorf("callback preview record flow has %d records, want %d", len(records), len(expected))
	}
	for index, record := range records {
		if record.Entity != expected[index] {
			return fmt.Errorf("callback preview record flow step %d is %q, want %q", index, record.Entity, expected[index])
		}
		if err := contract.validateCallbackPreviewRecordFields(record); err != nil {
			return err
		}
		if _, err := contract.BuildCallbackPreviewRecord(record.Entity, callbackPreviewRecordValues(record)); err != nil {
			return err
		}
	}
	if len(contract.Activities) != 4 || len(contract.Bindings) != 3 {
		return fmt.Errorf("callback preview contract activity/binding flow is incomplete")
	}
	for index, expectedBinding := range callbackPreviewBindingSpecs {
		matched := false
		for _, binding := range contract.Bindings {
			if binding.Entity == expectedBinding.Entity && binding.ProducerActivity == expectedBinding.ProducerActivity && binding.ConsumerActivity == expectedBinding.ConsumerActivity {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("callback preview contract binding %d does not match activity flow", index)
		}
	}
	return nil
}

func (contract CallbackPreviewContractEvidence) ValidateCallbackPreviewRecord(record CallbackPreviewRecord) error {
	return contract.validateCallbackPreviewRecordFields(record)
}

func (contract CallbackPreviewContractEvidence) validateCallbackPreviewRecordFields(record CallbackPreviewRecord) error {
	expected := make(map[string]string)
	for _, field := range contract.Fields {
		if field.Entity == record.Entity {
			expected[field.Name] = field.ID
		}
	}
	if len(expected) != len(record.Fields) {
		return fmt.Errorf("callback preview record %s has an unexpected field count", record.Entity)
	}
	seen := make(map[string]bool, len(record.Fields))
	for _, field := range record.Fields {
		id, ok := expected[field.Name]
		if !ok || seen[field.Name] || id != field.ID {
			return fmt.Errorf("callback preview record %s field %s has an unbound ID", record.Entity, field.Name)
		}
		seen[field.Name] = true
		if field.Name == "EffectBlockedBy" {
			if err := ValidateCallbackPreviewNestedList(field.Value); err != nil {
				return fmt.Errorf("callback preview record %s field %s: %w", record.Entity, field.Name, err)
			}
		} else if callbackPreviewListField(field.Name) {
			if err := ValidateCallbackPreviewList(field.Value); err != nil {
				return fmt.Errorf("callback preview record %s field %s: %w", record.Entity, field.Name, err)
			}
		}
	}
	return nil
}

func callbackPreviewListField(name string) bool {
	switch name {
	case "CaptureNames", "ObjectIdentities", "ObjectTypes", "BindingModes", "CallIdentities", "Symbols", "Signatures", "ReceiverTypes", "EffectKinds", "States", "BlockedBy", "EffectStages", "EffectSteps", "EffectReasons", "EffectUnknownClasses", "EffectNextOperations":
		return true
	default:
		return false
	}
}

func callbackPreviewRecordValues(record CallbackPreviewRecord) map[string]string {
	values := make(map[string]string, len(record.Fields))
	for _, field := range record.Fields {
		values[field.Name] = field.Value
	}
	return values
}

//go:embed callback-preview-contract.gooo
var callbackPreviewContractSource []byte

type callbackPreviewEntitySpec struct {
	Name   string
	ID     string
	Fields []callbackPreviewFieldSpec
}

type callbackPreviewFieldSpec struct {
	Name        string
	ID          string
	Presence    semantic.Presence
	Cardinality semantic.Cardinality
}

type callbackPreviewActivitySpec struct {
	Name         string
	InputEntity  string
	OutputEntity string
	ValueProgram string
}

type callbackPreviewBindingSpec struct {
	ProducerActivity string
	ConsumerActivity string
	Entity           string
}

var callbackPreviewEntitySpecs = []callbackPreviewEntitySpec{
	{Name: "CallbackPreviewInput", ID: "gooo://meta-callback-preview/entity/input", Fields: []callbackPreviewFieldSpec{
		{Name: "LogicalPath", ID: "gooo://meta-callback-preview/field/input-logical-path", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Subject", ID: "gooo://meta-callback-preview/field/input-subject", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "SourceDigest", ID: "gooo://meta-callback-preview/field/input-source-digest", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "State", ID: "gooo://meta-callback-preview/field/input-state", Presence: semantic.Required, Cardinality: semantic.One},
	}},
	{Name: "BoundedCallbackCandidate", ID: "gooo://meta-callback-preview/entity/candidate", Fields: []callbackPreviewFieldSpec{
		{Name: "CandidateIdentity", ID: "gooo://meta-callback-preview/field/candidate-identity", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "SourceDigest", ID: "gooo://meta-callback-preview/field/candidate-source-digest", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CandidateDigest", ID: "gooo://meta-callback-preview/field/candidate-digest", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "HelperName", ID: "gooo://meta-callback-preview/field/candidate-helper-name", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "HelperBytes", ID: "gooo://meta-callback-preview/field/candidate-helper-bytes", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "HelperLines", ID: "gooo://meta-callback-preview/field/candidate-helper-lines", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ParentFunctionLines", ID: "gooo://meta-callback-preview/field/candidate-parent-lines", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CaptureCount", ID: "gooo://meta-callback-preview/field/candidate-capture-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "PendingEffectCount", ID: "gooo://meta-callback-preview/field/candidate-pending-effect-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "State", ID: "gooo://meta-callback-preview/field/candidate-state", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Promotion", ID: "gooo://meta-callback-preview/field/candidate-promotion", Presence: semantic.Required, Cardinality: semantic.One},
	}},
	{Name: "CallbackCaptures", ID: "gooo://meta-callback-preview/entity/captures", Fields: []callbackPreviewFieldSpec{
		{Name: "CandidateIdentity", ID: "gooo://meta-callback-preview/field/captures-candidate-identity", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CaptureNames", ID: "gooo://meta-callback-preview/field/captures-names", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ObjectIdentities", ID: "gooo://meta-callback-preview/field/captures-object-identities", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ObjectTypes", ID: "gooo://meta-callback-preview/field/captures-object-types", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "BindingModes", ID: "gooo://meta-callback-preview/field/captures-binding-modes", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Count", ID: "gooo://meta-callback-preview/field/captures-count", Presence: semantic.Required, Cardinality: semantic.One},
	}},
	{Name: "PendingCallbackEffects", ID: "gooo://meta-callback-preview/entity/pending-effects", Fields: []callbackPreviewFieldSpec{
		{Name: "CandidateIdentity", ID: "gooo://meta-callback-preview/field/effects-candidate-identity", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CallIdentities", ID: "gooo://meta-callback-preview/field/effects-call-identities", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Symbols", ID: "gooo://meta-callback-preview/field/effects-symbols", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Signatures", ID: "gooo://meta-callback-preview/field/effects-signatures", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ReceiverTypes", ID: "gooo://meta-callback-preview/field/effects-receiver-types", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectKinds", ID: "gooo://meta-callback-preview/field/effects-kinds", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "States", ID: "gooo://meta-callback-preview/field/effects-states", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectStages", ID: "gooo://meta-callback-preview/field/effects-stages", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectSteps", ID: "gooo://meta-callback-preview/field/effects-steps", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectReasons", ID: "gooo://meta-callback-preview/field/effects-reasons", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectUnknownClasses", ID: "gooo://meta-callback-preview/field/effects-unknown-classes", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectNextOperations", ID: "gooo://meta-callback-preview/field/effects-next-operations", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "EffectBlockedBy", ID: "gooo://meta-callback-preview/field/effects-blocked-by", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Count", ID: "gooo://meta-callback-preview/field/effects-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ResolvedCount", ID: "gooo://meta-callback-preview/field/effects-resolved-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "State", ID: "gooo://meta-callback-preview/field/effects-state", Presence: semantic.Required, Cardinality: semantic.One},
	}},
	{Name: "CallbackPreviewEvidence", ID: "gooo://meta-callback-preview/entity/evidence", Fields: []callbackPreviewFieldSpec{
		{Name: "CandidateIdentity", ID: "gooo://meta-callback-preview/field/evidence-candidate-identity", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "SourceDigest", ID: "gooo://meta-callback-preview/field/evidence-source-digest", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CandidateDigest", ID: "gooo://meta-callback-preview/field/evidence-candidate-digest", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "State", ID: "gooo://meta-callback-preview/field/evidence-state", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "CaptureCount", ID: "gooo://meta-callback-preview/field/evidence-capture-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "PendingEffectCount", ID: "gooo://meta-callback-preview/field/evidence-pending-effect-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ResolvedEffectCount", ID: "gooo://meta-callback-preview/field/evidence-resolved-effect-count", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "HelperLines", ID: "gooo://meta-callback-preview/field/evidence-helper-lines", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ParentFunctionLines", ID: "gooo://meta-callback-preview/field/evidence-parent-lines", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "OperationResultAdmission", ID: "gooo://meta-callback-preview/field/evidence-operation-result-admission", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "ApplyPermission", ID: "gooo://meta-callback-preview/field/evidence-apply-permission", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Stage", ID: "gooo://meta-callback-preview/field/evidence-stage", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Step", ID: "gooo://meta-callback-preview/field/evidence-step", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "Reason", ID: "gooo://meta-callback-preview/field/evidence-reason", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "UnknownClass", ID: "gooo://meta-callback-preview/field/evidence-unknown-class", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "NextOperation", ID: "gooo://meta-callback-preview/field/evidence-next-operation", Presence: semantic.Required, Cardinality: semantic.One},
		{Name: "BlockedBy", ID: "gooo://meta-callback-preview/field/evidence-blocked-by", Presence: semantic.Required, Cardinality: semantic.One},
	}},
}

var callbackPreviewActivitySpecs = []callbackPreviewActivitySpec{
	{Name: "GenerateBoundedCallbackCandidate", InputEntity: "CallbackPreviewInput", OutputEntity: "BoundedCallbackCandidate", ValueProgram: "callback-preview.candidate:v1;promotion=NONE"},
	{Name: "BindCallbackCaptures", InputEntity: "BoundedCallbackCandidate", OutputEntity: "CallbackCaptures", ValueProgram: "callback-preview.captures:v1"},
	{Name: "RecordPendingCallbackEffects", InputEntity: "CallbackCaptures", OutputEntity: "PendingCallbackEffects", ValueProgram: "callback-preview.effects:v1;state=UNKNOWN"},
	{Name: "CloseCallbackPreview", InputEntity: "PendingCallbackEffects", OutputEntity: "CallbackPreviewEvidence", ValueProgram: "callback-preview.evidence:v1;operation-result=FORBIDDEN;apply=FORBIDDEN"},
}

var callbackPreviewBindingSpecs = []callbackPreviewBindingSpec{
	{ProducerActivity: "GenerateBoundedCallbackCandidate", ConsumerActivity: "BindCallbackCaptures", Entity: "BoundedCallbackCandidate"},
	{ProducerActivity: "BindCallbackCaptures", ConsumerActivity: "RecordPendingCallbackEffects", Entity: "CallbackCaptures"},
	{ProducerActivity: "RecordPendingCallbackEffects", ConsumerActivity: "CloseCallbackPreview", Entity: "PendingCallbackEffects"},
}

// LoadCallbackPreviewContract parses, lowers, and validates the explicit
// EntityFields V1 contract. It is intentionally separate from the
// OperationResult contract loader.
func LoadCallbackPreviewContract() (CallbackPreviewContractEvidence, error) {
	return parseCallbackPreviewContract(callbackPreviewContractSource)
}

func parseCallbackPreviewContract(raw []byte) (CallbackPreviewContractEvidence, error) {
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport("callback-preview-contract.gooo", string(raw), syntax.EntityFieldsV1Support())
	if file == nil || diagnostics.HasErrors() {
		return CallbackPreviewContractEvidence{}, fmt.Errorf("callback preview contract has syntax errors")
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(context.Background(), file, bidir.EntityFieldsV1Support())
	if err != nil {
		return CallbackPreviewContractEvidence{}, fmt.Errorf("lower callback preview contract: %w", err)
	}
	if err := validateCallbackPreviewIR(file, ir); err != nil {
		return CallbackPreviewContractEvidence{}, err
	}

	sourceDigest := sha256.Sum256(raw)
	fields := make([]CallbackPreviewFieldEvidence, 0)
	for _, entity := range callbackPreviewEntitySpecs {
		node, _ := ir.Graph.NodeByName(ir.Namespace, entity.Name)
		for _, field := range node.Fields {
			fields = append(fields, CallbackPreviewFieldEvidence{Entity: entity.Name, Name: field.Name, ID: field.ID.String(), Presence: string(field.Presence), Cardinality: string(field.Cardinality), TypeID: field.TypeRef.ID.String()})
		}
	}
	slices.SortFunc(fields, func(left, right CallbackPreviewFieldEvidence) int {
		if left.Entity != right.Entity {
			return stringsCompare(left.Entity, right.Entity)
		}
		return stringsCompare(left.ID, right.ID)
	})
	activities := make([]CallbackPreviewActivityEvidence, 0, len(callbackPreviewActivitySpecs))
	bindings := make([]CallbackPreviewBindingEvidence, 0, len(callbackPreviewBindingSpecs))
	for _, activity := range callbackPreviewActivitySpecs {
		node, _ := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		input, _ := ir.Graph.NodeByName(ir.Namespace, activity.InputEntity)
		output, _ := ir.Graph.NodeByName(ir.Namespace, activity.OutputEntity)
		activities = append(activities, CallbackPreviewActivityEvidence{Name: activity.Name, InputEntity: activity.InputEntity, OutputEntity: activity.OutputEntity, ValueProgram: node.ValueProgram, UsedInputFact: ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}), GeneratedOutputFact: ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID})})
	}
	for _, binding := range ir.RuntimeBindings {
		producer, _ := ir.Graph.Node(binding.ProducerActivity)
		consumer, _ := ir.Graph.Node(binding.ConsumerActivity)
		entity, _ := ir.Graph.Node(binding.Entity)
		bindings = append(bindings, CallbackPreviewBindingEvidence{ProducerActivity: producer.Name, ProducerPort: binding.ProducerPort, ConsumerActivity: consumer.Name, ConsumerPort: binding.ConsumerPort, Entity: entity.Name})
	}
	sort.Slice(bindings, func(left, right int) bool { return bindings[left].ProducerActivity < bindings[right].ProducerActivity })
	return CallbackPreviewContractEvidence{
		SourceDigest: hex.EncodeToString(sourceDigest[:]), SemanticDigest: ir.StableHash(),
		InputEntity: "CallbackPreviewInput", CandidateEntity: "BoundedCallbackCandidate", CapturesEntity: "CallbackCaptures", EffectsEntity: "PendingCallbackEffects", EvidenceEntity: "CallbackPreviewEvidence",
		Fields: fields, Activities: activities, Bindings: bindings,
	}, nil
}

func validateCallbackPreviewIR(file *syntax.File, ir semantic.IR) error {
	if file.Package == nil || file.Namespace == nil || file.Package.Name != "callbackpreview" || file.Namespace.Name != "callbackpreview" {
		return fmt.Errorf("callback preview contract headers are not exact")
	}
	if len(file.Decls) != len(callbackPreviewEntitySpecs)+len(callbackPreviewActivitySpecs) || len(ir.Graph.Nodes()) != len(file.Decls) || len(file.Bindings) != len(callbackPreviewBindingSpecs) {
		return fmt.Errorf("callback preview contract declaration cardinality is not exact")
	}
	for _, entity := range callbackPreviewEntitySpecs {
		declaration, ok := findCallbackPreviewEntity(file, entity.Name)
		if !ok || declaration.ID != entity.ID || len(declaration.Fields) != len(entity.Fields) {
			return fmt.Errorf("callback preview entity %q is not exact", entity.Name)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, entity.Name)
		if !ok || node.Kind != semantic.Entity || node.ID.String() != entity.ID || len(node.Fields) != len(entity.Fields) {
			return fmt.Errorf("callback preview semantic entity %q is not exact", entity.Name)
		}
		for index, expected := range entity.Fields {
			actual := node.Fields[index]
			if actual.Name != expected.Name || actual.ID.String() != expected.ID || actual.Presence != expected.Presence || actual.Cardinality != expected.Cardinality || actual.TypeRef.ID != semantic.BuiltinStringTypeID {
				return fmt.Errorf("callback preview field %s.%s is not exact", entity.Name, expected.Name)
			}
		}
	}
	for _, activity := range callbackPreviewActivitySpecs {
		declaration, ok := findCallbackPreviewActivity(file, activity.Name)
		if !ok || len(declaration.Inputs) != 1 || declaration.Inputs[0].Name != activity.InputEntity || declaration.Output != activity.OutputEntity || !declaration.ValueProgramPresent || declaration.ValueProgram != activity.ValueProgram {
			return fmt.Errorf("callback preview activity %q is not exact", activity.Name)
		}
		node, ok := ir.Graph.NodeByName(ir.Namespace, activity.Name)
		if !ok || node.Kind != semantic.Activity || node.ValueProgram != activity.ValueProgram {
			return fmt.Errorf("callback preview semantic activity %q is not exact", activity.Name)
		}
		input, inputOK := ir.Graph.NodeByName(ir.Namespace, activity.InputEntity)
		output, outputOK := ir.Graph.NodeByName(ir.Namespace, activity.OutputEntity)
		if !inputOK || !outputOK || !ir.Graph.HasFact(semantic.FactKey{Subject: node.ID, Predicate: semantic.Used, Object: input.ID}) || !ir.Graph.HasFact(semantic.FactKey{Subject: output.ID, Predicate: semantic.WasGeneratedBy, Object: node.ID}) {
			return fmt.Errorf("callback preview semantic facts are incomplete for %q", activity.Name)
		}
	}
	for _, binding := range callbackPreviewBindingSpecs {
		matched := false
		for _, actual := range ir.RuntimeBindings {
			producer, producerOK := ir.Graph.Node(actual.ProducerActivity)
			consumer, consumerOK := ir.Graph.Node(actual.ConsumerActivity)
			entity, entityOK := ir.Graph.Node(actual.Entity)
			if producerOK && consumerOK && entityOK && producer.Name == binding.ProducerActivity && consumer.Name == binding.ConsumerActivity && actual.ProducerPort == semantic.RuntimeOutputPort && actual.ConsumerPort == semantic.RuntimeInputPort && entity.Name == binding.Entity {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("callback preview runtime binding %s -> %s is not exact", binding.ProducerActivity, binding.ConsumerActivity)
		}
	}
	return nil
}

func findCallbackPreviewEntity(file *syntax.File, name string) (*syntax.EntityDecl, bool) {
	for _, declaration := range file.Decls {
		if entity, ok := declaration.(*syntax.EntityDecl); ok && entity.Name == name {
			return entity, true
		}
	}
	return nil, false
}

func findCallbackPreviewActivity(file *syntax.File, name string) (*syntax.ActivityDecl, bool) {
	for _, declaration := range file.Decls {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok && activity.Name == name {
			return activity, true
		}
	}
	return nil, false
}

func stringsCompare(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
