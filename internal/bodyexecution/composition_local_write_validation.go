package bodyexecution

import "fmt"

func validateLocalWrites(t CompositionTrace, sites []CompositionLocalSite, deliveries map[string]CompositionDelivery, bind func(int, string, string) error) ([]int, error) {
	completed := []int{}
	lastRead, lastUse := -1, -1
	for index, w := range t.LocalWrites {
		if w.SiteIndex < 0 || w.SiteIndex >= len(sites) || w.ReadStart < 0 || w.ReadEnd < w.ReadStart || w.ReadEnd > len(t.Reads) ||
			w.UseStart < 0 || w.UseEnd < w.UseStart || w.UseEnd > len(t.LocalUses) || w.ReadsDropped < 0 || w.ReadsDropped > t.ReadsDropped ||
			w.UsesDropped < 0 || w.UsesDropped > t.LocalUseDropped || w.ReadStart < lastRead || w.UseStart < lastUse {
			return nil, fmt.Errorf("local assignment coordinates differ")
		}
		site := sites[w.SiteIndex]
		if err := bind(w.Frame, site.RootActivityID, site.ActivityID); err != nil {
			return nil, err
		}
		if !w.Completed && deliveries[site.RootActivityID].Fault == nil {
			return nil, fmt.Errorf("interrupted assignment has no activity fault")
		}
		if w.Completed {
			completed = append(completed, index)
		}
		lastRead, lastUse = w.ReadStart, w.UseStart
	}
	return completed, nil
}
