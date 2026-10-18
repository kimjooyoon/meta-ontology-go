package bodyexecution

import "fmt"

func validateRecordedConstruction(runtime CompositionRuntime, trace CompositionTrace, created CompositionConstruction, reserved *int) error {
	if created.SiteIndex < 0 || created.SiteIndex >= len(runtime.ConstructionSites) || created.ReadStart < 0 || created.ReadEnd < created.ReadStart || created.ReadEnd > len(trace.Reads) {
		return fmt.Errorf("recorded construction site or read window differs")
	}
	site := runtime.ConstructionSites[created.SiteIndex]
	count, valid := constructionSize(site, created.ElementCount)
	*reserved += count
	if !valid || *reserved > constructionElementLimit || len(created.Elements) > count || created.Completed && len(created.Elements) != count || !created.Completed && len(created.Elements) == 0 {
		return fmt.Errorf("recorded construction element population differs")
	}
	_, delivery := recordedOriginDelivery(trace, site.RootActivityID)
	if delivery == nil || len(delivery.BlockedBy) > 0 || !created.Completed && delivery.Fault == nil {
		return fmt.Errorf("recorded construction differs from root delivery")
	}
	for _, read := range trace.Reads[created.ReadStart:created.ReadEnd] {
		if read.SiteIndex < 0 || read.SiteIndex >= len(runtime.ReadSites) || runtime.ReadSites[read.SiteIndex].RootActivityID != site.RootActivityID {
			return fmt.Errorf("construction window names another root's reads")
		}
	}
	return validateRecordedElements(created, trace.ReadsDropped)
}

func validateRecordedElements(created CompositionConstruction, dropped int) error {
	cursor := created.ReadStart
	for i, element := range created.Elements {
		if element.ReadStart != cursor || element.ReadEnd < cursor || element.ReadEnd > created.ReadEnd || element.ReadsDropped < 0 || element.ReadsDropped > dropped {
			return fmt.Errorf("recorded element read window differs")
		}
		if !element.Completed && (created.Completed || i != len(created.Elements)-1) {
			return fmt.Errorf("recorded constructor continued after an interrupted element")
		}
		cursor = element.ReadEnd
	}
	if cursor != created.ReadEnd || !created.Completed && created.Elements[len(created.Elements)-1].Completed {
		return fmt.Errorf("recorded construction completion differs from its elements")
	}
	return nil
}
