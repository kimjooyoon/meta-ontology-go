package main

import (
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type fixPlan struct {
	SchemaVersion string              `json:"schema_version"`
	Status        string              `json:"status"`
	SourceDigest  string              `json:"source_digest"`
	GraphHash     string              `json:"graph_hash,omitempty"`
	IR            graphIRStatus       `json:"ir"`
	Diagnostics   []fixPlanDiagnostic `json:"diagnostics"`
	Evidence      graphReferenceState `json:"evidence"`
	Provenance    graphReferenceState `json:"provenance"`
	Projection    graphStatus         `json:"projection"`
	Lowering      graphStatus         `json:"lowering"`
	Output        graphStatus         `json:"output"`
	Repairs       graphStatus         `json:"repairs"`
	GraphPatch    graphStatus         `json:"graph_patch"`
	Workspace     graphStatus         `json:"workspace"`
	SemanticLoop  graphStatus         `json:"semantic_loop"`
	Authorities   graphAuthorities    `json:"authorities"`
}

type fixPlanDiagnostic struct {
	RepairID      string      `json:"repair_id"`
	Phase         string      `json:"phase"`
	Severity      string      `json:"severity"`
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Span          fixPlanSpan `json:"span"`
	Applicability string      `json:"applicability"`
	Status        string      `json:"status"`
}

type fixPlanSpan struct {
	File  string          `json:"file"`
	Start fixPlanPosition `json:"start"`
	End   fixPlanPosition `json:"end"`
}

type fixPlanPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

func fixPlanSpanFromSyntax(span syntax.Span) fixPlanSpan {
	return fixPlanSpan{
		File:  span.Filename,
		Start: fixPlanPosition{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   fixPlanPosition{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}

func fileSpan(file *syntax.File) fixPlanSpan {
	if file == nil {
		return fixPlanSpan{}
	}
	return fixPlanSpanFromSyntax(file.Span)
}

func (span fixPlanSpan) canonical() string {
	values := []string{
		span.File,
		strconv.Itoa(span.Start.Offset), strconv.Itoa(span.Start.Line), strconv.Itoa(span.Start.Column),
		strconv.Itoa(span.End.Offset), strconv.Itoa(span.End.Line), strconv.Itoa(span.End.Column),
	}
	return strings.Join(values, "\x00")
}
