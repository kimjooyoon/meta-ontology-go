package sourceexecution

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func Execute(request Request) Receipt {
	sourceDigest := digestBytes([]byte(request.Source))
	if strings.TrimSpace(request.Filename) == "" || request.Source == "" || strings.TrimSpace(request.Entry) == "" {
		return rejected(request, sourceDigest, "REQUEST", "SOURCE_EXECUTION_REQUEST_INVALID",
			"filename, source, and entry activity are required")
	}
	file, diagnostics := syntax.ParseFile(request.Filename, request.Source)
	if file == nil || diagnostics.HasErrors() {
		return rejected(request, sourceDigest, "PARSE", "SOURCE_SYNTAX_INVALID",
			"source has syntax diagnostics")
	}
	if file.Package == nil || file.Namespace == nil {
		return rejected(request, sourceDigest, "PARSE", "SOURCE_HEADER_UNKNOWN",
			"package and namespace headers are required")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return rejected(request, sourceDigest, "LOWER", "SOURCE_SEMANTIC_INVALID",
			"source semantic lowering failed")
	}
	entry, code, message := resolveEntry(file, request.Entry)
	if code != "" {
		return rejected(request, sourceDigest, "EXECUTE", code, message)
	}
	semanticDigest := "sha256:" + ir.StableHash()
	receipt := Receipt{
		Schema: ReceiptSchema, Decision: "PASS", Reason: "SOURCE_ACTIVITY_EXECUTED",
		Resolution: "EXACT", Filename: request.Filename, SourceDigest: sourceDigest,
		SemanticDigest: semanticDigest, Entry: entry, Diagnostics: []Diagnostic{}, Effects: Effects{},
		Events: []Event{
			{Sequence: 1, Kind: "SOURCE_PARSED", Subject: sourceDigest},
			{Sequence: 2, Kind: "SEMANTIC_LOWERED", Subject: semanticDigest},
			{Sequence: 3, Kind: "ACTIVITY_INVOKED", Subject: entry.Activity},
			{Sequence: 4, Kind: "ENTITY_PRODUCED", Subject: entry.Output.ID},
		},
	}
	return seal(receipt)
}
