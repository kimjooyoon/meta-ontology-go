package bodycodegen

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type BodyReadSource = BodyExpressionSource

// BodyExpressionSource binds decoded syntax to the selected Gooo source literal.
type BodyExpressionSource struct {
	Status       string `json:"status"`
	Reason       string `json:"reason,omitempty"`
	View         string `json:"view"`
	SourceSHA256 string `json:"source_sha256"`
	BodySHA256   string `json:"body_sha256,omitempty"`
	Start        int    `json:"start_byte"`
	End          int    `json:"end_byte"`
	DecodedStart int    `json:"decoded_start"`
	DecodedEnd   int    `json:"decoded_end"`
	Expression   string `json:"expression,omitempty"`
}

func UnmappedReadSource(sourceSHA, reason string) *BodyReadSource {
	return &BodyReadSource{Status: "UNMAPPED", Reason: reason, View: "selected_gooo_source", SourceSHA256: sourceSHA}
}

func BindBodyReadSources(source []byte, origins []BodyReadOrigin) ([]*BodyReadSource, error) {
	activities, err := bodySourceActivities(source)
	if err != nil {
		return nil, err
	}
	result := make([]*BodyReadSource, len(origins))
	sourceSHA := digest(source)
	for i, origin := range origins {
		result[i] = bindBodyReadSource(source, sourceSHA, activities[origin.ActivityID], origin)
	}
	return result, nil
}

func bindBodyReadSource(source []byte, sourceSHA string, a *syntax.ActivityDecl, origin BodyReadOrigin) *BodyReadSource {
	fail := func(reason string) *BodyReadSource { return UnmappedReadSource(sourceSHA, reason) }
	if a == nil || !a.ValueProgramPresent {
		return fail("ACTIVITY_BODY_UNAVAILABLE")
	}
	if origin.Kind != "index" && origin.Kind != "range_value" && origin.Kind != "map_value" {
		return fail("FORWARD_ORIGIN_DIFFERS")
	}
	return bindBodyExpressionSource(source, sourceSHA, a, origin.BodySHA256, BodyElementOrigin{origin.Start, origin.End, origin.Expression})
}

func bindBodyExpressionSource(source []byte, sourceSHA string, a *syntax.ActivityDecl, bodySHA string, origin BodyElementOrigin) *BodyExpressionSource {
	fail := func(reason string) *BodyExpressionSource { return UnmappedReadSource(sourceSHA, reason) }
	if a == nil || !a.ValueProgramPresent {
		return fail("ACTIVITY_BODY_UNAVAILABLE")
	}
	normalized, mapping, err := rewriteBodySurface(a.ValueProgram)
	if err != nil || digest([]byte(normalized)) != bodySHA {
		return fail("SELECTED_BODY_DIFFERS")
	}
	if origin.Start < 0 || origin.End <= origin.Start || origin.End > len(normalized) || normalized[origin.Start:origin.End] != origin.Expression {
		return fail("FORWARD_ORIGIN_DIFFERS")
	}
	span := a.ValueProgramSpan
	if span.Start.Offset < 0 || span.End.Offset > len(source) || span.End.Offset < span.Start.Offset {
		return fail("BODY_LITERAL_UNAVAILABLE")
	}
	decoded := mapping.originalRange(FillOriginRange{Start: origin.Start, End: origin.End})
	start, end, err := SourceLiteralByteRange(string(source[span.Start.Offset:span.End.Offset]), a.ValueProgram, decoded.Start, decoded.End)
	if err != nil {
		return fail("BODY_LITERAL_RANGE_UNAVAILABLE")
	}
	return &BodyReadSource{Status: "MAPPED", View: "selected_gooo_source", SourceSHA256: sourceSHA, BodySHA256: bodySHA,
		Start: span.Start.Offset + start, End: span.Start.Offset + end, DecodedStart: decoded.Start, DecodedEnd: decoded.End, Expression: a.ValueProgram[decoded.Start:decoded.End]}
}

func bodySourceActivities(source []byte) (map[string]*syntax.ActivityDecl, error) {
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
	return activities, nil
}
