package workspaceexecution

import (
	"context"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type OriginQuery struct {
	Follow      bool   `json:"follow,omitempty"`
	ActivityID  string `json:"activity_id,omitempty"`
	CandidateID string `json:"candidate_id,omitempty"`
	Limit       int    `json:"limit"`
}

type OriginUseReport struct {
	ValueFlow       *ProgramValueOrigins `json:"value_flow,omitempty"`
	Schema          string               `json:"schema"`
	Query           OriginQuery          `json:"query"`
	OriginsSeen     int                  `json:"origins_seen"`
	OriginsMatched  int                  `json:"origins_matched"`
	GraphsAvailable int                  `json:"graphs_available"`
	Truncated       bool                 `json:"truncated"`
	Rows            []RecordedOriginUse  `json:"rows"`
	Issues          []OriginQueryIssue   `json:"issues"`
	Scope           string               `json:"scope"`
}

type OriginQueryIssue struct {
	ActivityID  string `json:"activity_id"`
	CandidateID string `json:"candidate_id"`
	Reason      string `json:"reason"`
}

type RecordedOriginUse struct {
	ActivityID        string                    `json:"activity_id"`
	CandidateID       string                    `json:"candidate_id"`
	BodySHA256        string                    `json:"body_sha256"`
	AnalysisStatus    string                    `json:"analysis_status"`
	Use               bodycodegen.FillOriginUse `json:"use"`
	Expressions       []string                  `json:"expressions"`
	SnippetsTruncated bool                      `json:"snippets_truncated,omitempty"`
	Locations         []OriginUseLocation       `json:"locations"`
}

type OriginUseLocation struct {
	View         string `json:"view"`
	PackagePath  string `json:"package_path,omitempty"`
	Filename     string `json:"filename"`
	SourceSHA256 string `json:"source_sha256"`
	ActivityID   string `json:"activity_id"`
	Start        int    `json:"start_byte"`
	End          int    `json:"end_byte"`
}

// InspectRecordedOrigins reads retained construction records without selecting
// candidates, executing bodies or reinterpreting past runtime observations.
func InspectRecordedOrigins(ctx context.Context, result Result, query OriginQuery) (OriginUseReport, error) {
	out := OriginUseReport{Schema: "gooo/recorded-origin-uses/v1", Query: query,
		Scope: "one shortest recorded static path per condition/return root and filled hole; selected construction only; no body, candidate or model execution"}
	if ctx == nil || query.Limit < 1 || query.Limit > 128 {
		return out, fmt.Errorf("origin query requires context and a limit of 1..128")
	}
	if result.Schema != "gooo/workspace-body-execution/v1" {
		return out, fmt.Errorf("origin query requires a saved workspace execution")
	}
	if query.Follow {
		flow, err := InspectSelectedValueOrigins(ctx, result, query)
		if err != nil {
			return out, err
		}
		out.ValueFlow = &flow
	}
	generations, err := originQueryGenerations(result)
	if err != nil {
		return out, err
	}
	seen := map[string]bool{}
	steps := 0
	for _, generation := range generations {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		fill := generation.Report.BodyFill
		if fill == nil {
			continue
		}
		origin := fill.SourceOrigin
		if origin == nil {
			key := "missing:" + generation.Report.ActivityID + "\x00" + fill.SelectedCandidateID + "\x00" + generation.Report.SourceDigest
			if seen[key] {
				continue
			}
			seen[key] = true
			out.OriginsSeen++
			if query.ActivityID != "" && generation.Report.ActivityID != query.ActivityID || query.CandidateID != "" && fill.SelectedCandidateID != query.CandidateID {
				continue
			}
			out.OriginsMatched++
			out.Issues = append(out.Issues, OriginQueryIssue{generation.Report.ActivityID, fill.SelectedCandidateID, "NO_RETAINED_SOURCE_ORIGIN"})
			continue
		}
		key := origin.ActivityID + "\x00" + origin.SourceSHA256 + "\x00" + origin.SelectedBodySHA256 + "\x00" + origin.CandidateID
		if seen[key] {
			continue
		}
		seen[key] = true
		out.OriginsSeen++
		if query.ActivityID != "" && origin.ActivityID != query.ActivityID || query.CandidateID != "" && origin.CandidateID != query.CandidateID {
			continue
		}
		out.OriginsMatched++
		if origin.DataFlow == nil {
			out.Issues = append(out.Issues, OriginQueryIssue{origin.ActivityID, origin.CandidateID, "NO_RETAINED_VALUE_GRAPH"})
			continue
		}
		out.GraphsAvailable++
		for _, reason := range origin.DataFlow.Reasons {
			out.Issues = append(out.Issues, OriginQueryIssue{origin.ActivityID, origin.CandidateID, reason})
		}
		uses, truncated, err := bodycodegen.SourceFillUses(*origin, query.Limit)
		if err != nil {
			return out, err
		}
		out.Truncated = out.Truncated || truncated
		body, err := originQueryBody(generation, *origin)
		if err != nil {
			return out, err
		}
		for _, use := range uses {
			if len(out.Rows) >= query.Limit || steps+len(use.Path) > 2048 {
				out.Truncated = true
				break
			}
			row := RecordedOriginUse{ActivityID: origin.ActivityID, CandidateID: origin.CandidateID,
				BodySHA256: origin.SelectedBodySHA256, AnalysisStatus: origin.DataFlow.Status, Use: use,
				Locations: originUseLocations(result, *origin, use.HoleID)}
			for _, step := range use.Path {
				span := step.Node.Decoded
				if span.Start < 0 || span.End < span.Start || span.End > len(body) {
					return out, fmt.Errorf("recorded dependency is outside its body")
				}
				snippet, cut := originQuerySnippet(body[span.Start:span.End])
				row.Expressions = append(row.Expressions, snippet)
				row.SnippetsTruncated = row.SnippetsTruncated || cut
			}
			out.Rows = append(out.Rows, row)
			steps += len(use.Path)
		}
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		a, b := out.Rows[i], out.Rows[j]
		if a.ActivityID != b.ActivityID {
			return a.ActivityID < b.ActivityID
		}
		if a.Use.RootID != b.Use.RootID {
			return a.Use.RootID < b.Use.RootID
		}
		return a.Use.HoleID < b.Use.HoleID
	})
	return out, nil
}

func originQueryGenerations(result Result) ([]bodycodegen.Result, error) {
	if len(result.BodyFills)+len(result.Composition.Steps)+len(result.Composition.Preparations) > 128 {
		return nil, fmt.Errorf("origin query exceeds 128 recorded construction steps")
	}
	var out []bodycodegen.Result
	for _, step := range result.BodyFills {
		out = append(out, step.Generation)
	}
	for _, step := range result.Composition.Steps {
		out = append(out, step.Generation)
	}
	for _, step := range result.Composition.Preparations {
		out = append(out, step.Generation)
	}
	return out, nil
}
