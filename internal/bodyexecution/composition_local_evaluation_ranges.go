package bodyexecution

import "fmt"

func validateLocalWindows(t CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite, reads []CompositionReadSite, constructions []CompositionConstructionSite) error {
	useRoot := func(start, end int, root string) error {
		for _, u := range t.LocalUses[start:end] {
			if uses[u.SiteIndex].RootActivityID != root {
				return fmt.Errorf("local evaluation window crosses root activity")
			}
		}
		return nil
	}
	for _, w := range t.LocalWrites {
		root := sites[w.SiteIndex].RootActivityID
		if err := useRoot(w.UseStart, w.UseEnd, root); err != nil {
			return err
		}
		for _, read := range t.Reads[w.ReadStart:w.ReadEnd] {
			if read.SiteIndex < 0 || read.SiteIndex >= len(reads) || reads[read.SiteIndex].RootActivityID != root {
				return fmt.Errorf("local assignment read window crosses root activity")
			}
		}
	}
	windows := make([][]*CompositionLocalElementWindow, len(t.Constructions))
	for i, c := range t.Constructions {
		windows[i] = make([]*CompositionLocalElementWindow, len(c.Elements))
	}
	for index := range t.LocalElements {
		e := &t.LocalElements[index]
		if e.Construction < 0 || e.Construction >= len(windows) || e.Element < 0 || e.Element >= len(windows[e.Construction]) ||
			e.UseStart < 0 || e.UseEnd < e.UseStart || e.UseEnd > len(t.LocalUses) || e.UsesDropped < 0 || e.UsesDropped > t.LocalUseDropped {
			return fmt.Errorf("local constructor window coordinates differ")
		}
		if windows[e.Construction][e.Element] != nil {
			return fmt.Errorf("duplicate local constructor window")
		}
		windows[e.Construction][e.Element] = e
	}
	for ci, elements := range windows {
		site := t.Constructions[ci].SiteIndex
		if site < 0 || site >= len(constructions) {
			return fmt.Errorf("local constructor site differs")
		}
		for i, e := range elements {
			if e == nil || i > 0 && constructions[site].Kind != "filter_values" && e.UseStart != elements[i-1].UseEnd ||
				i > 0 && constructions[site].Kind == "filter_values" && e.UseStart < elements[i-1].UseEnd {
				return fmt.Errorf("local constructor windows are missing or noncontiguous")
			}
			if err := useRoot(e.UseStart, e.UseEnd, constructions[site].RootActivityID); err != nil {
				return err
			}
		}
	}
	return nil
}
