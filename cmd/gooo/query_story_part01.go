package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func runQueryStory(options queryOptions, ir semantic.IR, filename string, jsonMode bool, stdout, stderr io.Writer) int {
	graph, err := queryengine.FromSemanticIR(ir)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.engine", err.Error(), syntax.Span{})
	}
	store := provenance.New(options.ledgerPath)
	snapshot, err := store.ReadOnly(provenance.ReadOptions{})
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "provenance.read_only", err.Error(), syntax.Span{})
	}
	id, err := queryengine.ParseID(options.storyID)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.story_id", err.Error(), syntax.Span{})
	}
	story, err := graph.StoryWithEvidenceBinding(id, snapshot)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.story", err.Error(), syntax.Span{})
	}
	if err := story.Validate(); err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.story_validation", err.Error(), syntax.Span{})
	}
	payload, err := json.Marshal(story)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.story_encode", err.Error(), syntax.Span{})
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "query.story_size", errDiagnosticLimit.Error(), syntax.Span{})
	}
	if _, err := fmt.Fprintf(stdout, "%s\n", payload); err != nil {
		return exitFailure
	}
	return exitOK
}
