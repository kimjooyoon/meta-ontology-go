package bodycodegen

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"reflect"
)

// BodyConditionSource locates a condition in the selected Gooo carrier. The
// carrier's hash is distinct from original workspace input-file identity.
type ConditionInputSource struct {
	Declaration  *ConditionDeclarationSource `json:"declaration,omitempty"`
	Status       string                      `json:"status"`
	Reason       string                      `json:"reason,omitempty"`
	View         string                      `json:"view"`
	PackagePath  string                      `json:"package_path,omitempty"`
	Filename     string                      `json:"filename,omitempty"`
	ActivityID   string                      `json:"activity_id,omitempty"`
	SourceSHA256 string                      `json:"source_sha256,omitempty"`
	Start        int                         `json:"start_byte"`
	End          int                         `json:"end_byte"`
	Expression   string                      `json:"expression,omitempty"`
}

type BodyConditionSource struct {
	Construction  *SourceFillConditionOrigin  `json:"construction,omitempty"`
	Declaration   *ConditionDeclarationSource `json:"declaration,omitempty"`
	Input         *ConditionInputSource       `json:"input_source,omitempty"`
	Status        string                      `json:"status"`
	Reason        string                      `json:"reason,omitempty"`
	View          string                      `json:"view"`
	SourceSHA256  string                      `json:"source_sha256"`
	BodySHA256    string                      `json:"body_sha256,omitempty"`
	Start         int                         `json:"start_byte"`
	End           int                         `json:"end_byte"`
	DecodedStart  int                         `json:"decoded_start"`
	DecodedEnd    int                         `json:"decoded_end"`
	SourceOrdinal int                         `json:"source_ordinal"`
	Expression    string                      `json:"expression,omitempty"`
	Relation      string                      `json:"relation,omitempty"`
	Fields        []ConditionSourceField      `json:"fields,omitempty"`
}

func UnmappedConditionSource(sourceSHA, reason string) *BodyConditionSource {
	return &BodyConditionSource{Status: "UNMAPPED", Reason: reason, View: "selected_gooo_source", SourceSHA256: sourceSHA}
}

// BindBodyConditionSources resolves forward origins once against the retained
// selected carrier. A mismatched body remains explicitly unmapped.
func BindBodyConditionSources(source []byte, origins []BodyConditionOrigin) ([]*BodyConditionSource, error) {
	file, diagnostics := ParseBodyFile("<selected-gooo-source>", source)
	if diagnostics.HasErrors() {
		return nil, diagnostics.Error()
	}
	model, _, err := resolveBodyModel(file)
	if err != nil {
		return nil, err
	}
	ids := map[string]string{}
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind {
			ids[node.Name] = string(node.ID)
		}
	}
	activities := map[string]*syntax.ActivityDecl{}
	for _, declaration := range file.Declarations {
		if a, ok := declaration.(*syntax.ActivityDecl); ok {
			activities[ids[a.Name]] = a
		}
	}
	result := make([]*BodyConditionSource, len(origins))
	fields := conditionFieldDeclarations(source, file, model)
	sourceSHA := digest(source)
	for i, origin := range origins {
		result[i] = bindBodyConditionSource(source, sourceSHA, activities[origin.ActivityID], origin)
		if result[i].Status == "MAPPED" {
			activity := activities[origin.ActivityID]
			result[i].Declaration = conditionDeclaration(sourceSHA, "activity", origin.ActivityID, activity.Name, activity.Span, len(source))
			for j := range result[i].Fields {
				field := &result[i].Fields[j]
				field.Declaration = fields[field.EntityID+"\x00"+field.FieldID]
			}
		}
	}
	return result, nil
}

func bindBodyConditionSource(source []byte, sourceSHA string, activity *syntax.ActivityDecl, origin BodyConditionOrigin) *BodyConditionSource {
	fail := func(reason string) *BodyConditionSource { return UnmappedConditionSource(sourceSHA, reason) }
	if activity == nil || !activity.ValueProgramPresent {
		return fail("ACTIVITY_BODY_UNAVAILABLE")
	}
	normalized, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil || digest([]byte(normalized)) != origin.BodySHA256 {
		return fail("SELECTED_BODY_DIFFERS")
	}
	if origin.Relation != "retained_condition_node" || origin.Start < 0 || origin.End < origin.Start || origin.End > len(normalized) || normalized[origin.Start:origin.End] != origin.Expression {
		return fail("FORWARD_ORIGIN_DIFFERS")
	}
	span := activity.ValueProgramSpan
	if span.Start.Offset < 0 || span.End.Offset > len(source) || span.End.Offset < span.Start.Offset {
		return fail("BODY_LITERAL_UNAVAILABLE")
	}
	start, end, err := SourceLiteralByteRange(string(source[span.Start.Offset:span.End.Offset]), activity.ValueProgram, origin.Start, origin.End)
	if err != nil {
		return fail("BODY_LITERAL_RANGE_UNAVAILABLE")
	}
	return &BodyConditionSource{Status: "MAPPED", View: "selected_gooo_source", SourceSHA256: sourceSHA, BodySHA256: origin.BodySHA256,
		Start: span.Start.Offset + start, End: span.Start.Offset + end, DecodedStart: origin.Start, DecodedEnd: origin.End, SourceOrdinal: origin.SourceOrdinal,
		Expression: activity.ValueProgram[origin.Start:origin.End], Relation: origin.Relation, Fields: append([]ConditionSourceField(nil), origin.Fields...)}
}

func EqualRetainedConditionOrigins(prior, expected []BodyConditionOrigin) bool {
	// Earlier receipts carry no source-origin claims. They remain unmapped.
	if prior == nil {
		return true
	}
	return reflect.DeepEqual(prior, expected)
}
