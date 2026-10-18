package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type nativeConstructionElement struct {
	ReadStart    *int  `json:"read_start"`
	ReadEnd      *int  `json:"read_end"`
	ReadsDropped *int  `json:"reads_dropped"`
	Completed    *bool `json:"completed"`
	InputIndex   *int  `json:"input_index,omitempty"`
}

type nativeConstruction struct {
	Case         *int                        `json:"case_index"`
	Site         *int                        `json:"site_index"`
	ReadStart    *int                        `json:"read_start"`
	ReadEnd      *int                        `json:"read_end"`
	Completed    *bool                       `json:"completed"`
	Elements     []nativeConstructionElement `json:"elements"`
	ElementCount *int                        `json:"element_count,omitempty"`
}

func (graph compositionGraph) constructionRows(events []nativeConstruction, reads []nativeReadEvent, readSites []CompositionReadSite, dropped, readDrops []int, context *compositionConstructionContext, summary *CompositionReadSummary) error {
	cases := len(readDrops)
	if context == nil || events == nil || len(events) > constructionObservationLimit || len(dropped) != cases {
		return fmt.Errorf("construction protocol or context differs")
	}
	context.rows = make([][]CompositionConstruction, cases)
	context.dropped = dropped
	counts := make([]int, cases)
	caseReads := make([][]nativeReadEvent, cases)
	for _, read := range reads {
		if read.Case == nil || *read.Case < 0 || *read.Case >= cases {
			return fmt.Errorf("construction read case differs")
		}
		counts[*read.Case]++
		caseReads[*read.Case] = append(caseReads[*read.Case], read)
	}
	reserved, lastCase, lastRoot, lastStart := 0, -1, -1, -1
	for _, event := range events {
		if event.Case == nil || event.Site == nil || event.ReadStart == nil || event.ReadEnd == nil || event.Completed == nil || event.Elements == nil {
			return fmt.Errorf("construction coordinates are required")
		}
		c, s := *event.Case, *event.Site
		if c < 0 || c >= cases || s < 0 || s >= len(context.sites) || *event.ReadStart < 0 || *event.ReadEnd < *event.ReadStart || *event.ReadEnd > counts[c] {
			return fmt.Errorf("construction case, site or read range differs")
		}
		site := context.sites[s]
		root := graph.index(site.RootActivityID)
		if root < 0 || c < lastCase || c == lastCase && (root < lastRoot || *event.ReadStart < lastStart) {
			return fmt.Errorf("construction evaluation order differs")
		}
		count, valid := constructionSize(site, event.ElementCount)
		reserved += count
		if !valid || reserved > constructionElementLimit || len(event.Elements) > count || *event.Completed && len(event.Elements) != count || !*event.Completed && len(event.Elements) == 0 && site.Kind != "filter_values" {
			return fmt.Errorf("construction element count differs")
		}
		row := CompositionConstruction{SiteIndex: s, ReadStart: *event.ReadStart, ReadEnd: *event.ReadEnd, Completed: *event.Completed, Elements: []CompositionConstructionElement{}, ElementCount: event.ElementCount}
		cursor := row.ReadStart
		previousInput := -1
		for i, e := range event.Elements {
			if e.ReadStart == nil || e.ReadEnd == nil || e.ReadsDropped == nil || e.Completed == nil || *e.ReadEnd > row.ReadEnd || *e.ReadsDropped < 0 || *e.ReadsDropped > readDrops[c] {
				return fmt.Errorf("construction element read range differs")
			}
			if site.Kind == "filter_values" {
				if e.InputIndex == nil || *e.InputIndex <= previousInput || *e.InputIndex >= 256 || *e.ReadStart < cursor || *e.ReadStart >= *e.ReadEnd || *e.ReadStart >= len(caseReads[c]) {
					return fmt.Errorf("filter selection source position differs")
				}
				previousInput = *e.InputIndex
			} else if e.InputIndex != nil || *e.ReadStart != cursor || *e.ReadEnd < cursor {
				return fmt.Errorf("construction element read range differs")
			}
			if !*e.Completed && (row.Completed || i != len(event.Elements)-1) {
				return fmt.Errorf("construction continued after an interrupted element")
			}
			row.Elements = append(row.Elements, CompositionConstructionElement{ReadStart: *e.ReadStart, ReadEnd: *e.ReadEnd, ReadsDropped: *e.ReadsDropped, Completed: *e.Completed, InputIndex: e.InputIndex})
			cursor = *e.ReadEnd
		}
		if site.Kind != "filter_values" && (cursor != row.ReadEnd || !row.Completed && row.Elements[len(row.Elements)-1].Completed) {
			return fmt.Errorf("construction completion differs from evaluated elements")
		}
		if site.Kind == "filter_values" {
			if err := validateNativeFilterSelection(row, caseReads[c], site, readSites, readDrops[c]); err != nil {
				return err
			}
		}
		if !row.Completed {
			summary.InterruptedConstructions++
		}
		context.rows[c] = append(context.rows[c], row)
		lastCase, lastRoot, lastStart = c, root, row.ReadStart
	}
	for _, n := range dropped {
		if n < 0 || n > int(^uint(0)>>1)-summary.ConstructionDropped || n > 0 && len(events) < constructionObservationLimit && reserved <= constructionElementLimit-256 {
			return fmt.Errorf("construction claimed impossible dropped instances")
		}
		summary.ConstructionDropped += n
	}
	summary.RecordedConstructions = len(events)
	return nil
}

func validateNativeFilterSelection(row CompositionConstruction, reads []nativeReadEvent, site CompositionConstructionSite, readSites []CompositionReadSite, dropped int) error {
	filterSite, err := filterInputReadSite(site, readSites)
	if err != nil {
		return err
	}
	inputReads := 0
	inputLength := 0
	for i := row.ReadStart; i < row.ReadEnd; i++ {
		read := reads[i]
		if read.Site == nil || *read.Site != filterSite {
			continue
		}
		if read.Index == nil || read.Length == nil || *read.Index != inputReads || (inputReads > 0 && *read.Length != inputLength) {
			return fmt.Errorf("filter source read order differs")
		}
		inputLength = *read.Length
		inputReads++
	}
	if row.Completed && dropped == 0 && inputReads != inputLength {
		return fmt.Errorf("filter omitted a source read")
	}
	for _, element := range row.Elements {
		read := reads[element.ReadStart]
		if read.Site == nil || *read.Site != filterSite || read.Index == nil || element.InputIndex == nil || *read.Index != *element.InputIndex {
			return fmt.Errorf("filter output has no matching source read")
		}
	}
	return nil
}

func filterInputReadSite(site CompositionConstructionSite, readSites []CompositionReadSite) (int, error) {
	fset := token.NewFileSet()
	expression, err := parser.ParseExprFrom(fset, "filter.go", site.Expression, 0)
	if err != nil {
		return -1, fmt.Errorf("filter construction expression differs: %w", err)
	}
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return -1, fmt.Errorf("filter construction requires a call")
	}
	input, _, recognized := bodycodegen.CollectionFilterProjection(call)
	if !recognized {
		return -1, fmt.Errorf("filter construction no longer matches its lowering")
	}
	start := site.Start + fset.Position(input.Pos()).Offset
	end := site.Start + fset.Position(input.End()).Offset
	index := -1
	for i, candidate := range readSites {
		if candidate.Kind == "filter_value" && candidate.RootActivityID == site.RootActivityID && candidate.ActivityID == site.ActivityID && candidate.Start == start && candidate.End == end {
			if index >= 0 {
				return -1, fmt.Errorf("filter has multiple source read sites")
			}
			index = i
		}
	}
	if index < 0 {
		return -1, fmt.Errorf("filter has no source read site")
	}
	return index, nil
}

func (graph compositionGraph) constructedRead(event nativeReadEvent, c, readIndex, root int, site CompositionReadSite, context *compositionConstructionContext) (*CompositionConstructedOrigin, error) {
	o := event.Construction
	if context == nil || o.Instance == nil || o.Index == nil || *o.Instance < 0 || *o.Instance >= len(context.rows[c]) || *o.Index < 0 || event.Origin != nil || event.Reason != "" {
		return nil, fmt.Errorf("constructed read origin is incomplete or ambiguous")
	}
	created := context.rows[c][*o.Instance]
	owner := context.sites[created.SiteIndex]
	if !created.Completed || created.ReadEnd > readIndex || *o.Index >= len(created.Elements) || graph.index(owner.RootActivityID) > root || owner.ElementType != site.ElementType {
		return nil, fmt.Errorf("constructed read precedes its completed origin or differs in type")
	}
	return &CompositionConstructedOrigin{Instance: *o.Instance, Index: *o.Index}, nil
}

func (graph compositionGraph) bindConstructions(context *compositionConstructionContext, traces []CompositionTrace, sites []CompositionReadSite) error {
	if context == nil || context.rows == nil {
		return nil
	}
	for c, rows := range context.rows {
		for _, row := range rows {
			root := graph.index(context.sites[row.SiteIndex].RootActivityID)
			if root < 0 || root >= len(traces[c].Deliveries) || len(traces[c].Deliveries[root].BlockedBy) > 0 || !row.Completed && traces[c].Deliveries[root].Fault == nil {
				return fmt.Errorf("construction differs from activity delivery")
			}
			for _, read := range traces[c].Reads[row.ReadStart:row.ReadEnd] {
				if sites[read.SiteIndex].RootActivityID != context.sites[row.SiteIndex].RootActivityID {
					return fmt.Errorf("element evaluation claimed another activity's reads")
				}
			}
		}
		traces[c].Constructions = rows
		traces[c].ConstructionsDropped = context.dropped[c]
	}
	return nil
}
