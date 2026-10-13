package workspaceexecution

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func originQueryBody(generation bodycodegen.Result, origin bodycodegen.SourceFillOrigin) (string, error) {
	if len(generation.GoooSource) > 128<<10 || sourceSHA256([]byte(generation.GoooSource)) != origin.SelectedSourceSHA256 {
		return "", fmt.Errorf("recorded selected source differs from its origin")
	}
	file, diagnostics := bodycodegen.ParseBodyFile(origin.Filename, []byte(generation.GoooSource))
	if diagnostics.HasErrors() {
		return "", diagnostics.Error()
	}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok || activity.Name != generation.Report.Activity {
			continue
		}
		sha, err := bodycodegen.NormalizedConditionBodySHA256(activity.ValueProgram)
		if err != nil || sha != origin.SelectedBodySHA256 {
			return "", fmt.Errorf("recorded selected body differs from its origin")
		}
		return activity.ValueProgram, nil
	}
	return "", fmt.Errorf("recorded origin activity is absent from selected source")
}

func originQuerySnippet(expression string) (string, bool) {
	if len(expression) <= 256 {
		return expression, false
	}
	end := 256
	for end > 0 && !utf8.RuneStart(expression[end]) {
		end--
	}
	return expression[:end], true
}

func originUseLocations(result Result, origin bodycodegen.SourceFillOrigin, hole string) []OriginUseLocation {
	var locations []OriginUseLocation
	for _, fragment := range origin.Fragments {
		if fragment.HoleID != hole {
			continue
		}
		span := fragment.ExpressionSource
		if fragment.Kind == "derived_grammar" {
			span = fragment.GrammarSource
		}
		if span != nil {
			locations = append(locations, OriginUseLocation{View: origin.View, Filename: origin.Filename,
				SourceSHA256: origin.SourceSHA256, ActivityID: origin.ActivityID, Start: span.Start, End: span.End})
		}
	}
	// Original-file locations were retained on condition links. Reuse only the
	// same assembly input, activity and selected candidate; expression text is
	// not an identity key. Return-only bodies retain the assembly-input view.
	for _, site := range result.Runtime.ConditionSites {
		if site.Source == nil || site.Source.Construction == nil {
			continue
		}
		link := site.Source.Construction
		if link.ActivityID != origin.ActivityID || link.SourceSHA256 != origin.SourceSHA256 || link.CandidateID != origin.CandidateID || link.Input == nil {
			continue
		}
		for _, fragment := range link.Input.Fragments {
			if fragment.HoleID != hole {
				continue
			}
			span := fragment.ExpressionSource
			if fragment.Kind == "derived_grammar" {
				span = fragment.GrammarSource
			}
			if span != nil {
				locations = append(locations, OriginUseLocation{View: link.Input.View, PackagePath: link.Input.PackagePath,
					Filename: link.Input.Filename, SourceSHA256: link.Input.SourceSHA256, ActivityID: link.Input.ActivityID, Start: span.Start, End: span.End})
			}
		}
	}
	sort.Slice(locations, func(i, j int) bool { return fmt.Sprint(locations[i]) < fmt.Sprint(locations[j]) })
	var distinct []OriginUseLocation
	for _, location := range locations {
		if len(distinct) == 0 || distinct[len(distinct)-1] != location {
			distinct = append(distinct, location)
		}
	}
	return distinct
}
