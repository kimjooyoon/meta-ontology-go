package bodyexecution

import "fmt"

type nativeConstructionElement struct {
	ReadStart    *int  `json:"read_start"`
	ReadEnd      *int  `json:"read_end"`
	ReadsDropped *int  `json:"reads_dropped"`
	Completed    *bool `json:"completed"`
}

type nativeConstruction struct {
	Case      *int                        `json:"case_index"`
	Site      *int                        `json:"site_index"`
	ReadStart *int                        `json:"read_start"`
	ReadEnd   *int                        `json:"read_end"`
	Completed *bool                       `json:"completed"`
	Elements  []nativeConstructionElement `json:"elements"`
}

func (graph compositionGraph) constructionRows(events []nativeConstruction, reads []nativeReadEvent, dropped, readDrops []int, context *compositionConstructionContext, summary *CompositionReadSummary) error {
	cases := len(readDrops)
	if context == nil || events == nil || len(events) > constructionObservationLimit || len(dropped) != cases {
		return fmt.Errorf("construction protocol or context differs")
	}
	context.rows = make([][]CompositionConstruction, cases)
	context.dropped = dropped
	counts := make([]int, cases)
	for _, read := range reads {
		if read.Case == nil || *read.Case < 0 || *read.Case >= cases {
			return fmt.Errorf("construction read case differs")
		}
		counts[*read.Case]++
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
		reserved += len(site.Elements)
		if reserved > constructionElementLimit || len(event.Elements) > len(site.Elements) || *event.Completed && len(event.Elements) != len(site.Elements) || !*event.Completed && len(event.Elements) == 0 {
			return fmt.Errorf("construction element count differs")
		}
		row := CompositionConstruction{SiteIndex: s, ReadStart: *event.ReadStart, ReadEnd: *event.ReadEnd, Completed: *event.Completed, Elements: []CompositionConstructionElement{}}
		cursor := row.ReadStart
		for i, e := range event.Elements {
			if e.ReadStart == nil || e.ReadEnd == nil || e.ReadsDropped == nil || e.Completed == nil || *e.ReadStart != cursor || *e.ReadEnd < cursor || *e.ReadEnd > row.ReadEnd || *e.ReadsDropped < 0 || *e.ReadsDropped > readDrops[c] {
				return fmt.Errorf("construction element read range differs")
			}
			if !*e.Completed && (row.Completed || i != len(event.Elements)-1) {
				return fmt.Errorf("construction continued after an interrupted element")
			}
			row.Elements = append(row.Elements, CompositionConstructionElement{ReadStart: *e.ReadStart, ReadEnd: *e.ReadEnd, ReadsDropped: *e.ReadsDropped, Completed: *e.Completed})
			cursor = *e.ReadEnd
		}
		if cursor != row.ReadEnd || !row.Completed && row.Elements[len(row.Elements)-1].Completed {
			return fmt.Errorf("construction completion differs from evaluated elements")
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
