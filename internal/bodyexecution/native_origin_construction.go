package bodyexecution

import "fmt"

func validateRecordedConstruction(runtime CompositionRuntime, trace CompositionTrace, created CompositionConstruction, reserved *int) error {
	if created.SiteIndex < 0 || created.SiteIndex >= len(runtime.ConstructionSites) || created.ReadStart < 0 || created.ReadEnd < created.ReadStart || created.ReadEnd > len(trace.Reads) {
		return fmt.Errorf("recorded construction site or read window differs")
	}
	site := runtime.ConstructionSites[created.SiteIndex]
	count, valid := constructionSize(site, created.ElementCount)
	*reserved += count
	if !valid || *reserved > constructionElementLimit || len(created.Elements) > count || created.Completed && len(created.Elements) != count || !created.Completed && len(created.Elements) == 0 && site.Kind != "filter_values" {
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
	if site.Kind == "filter_values" {
		return validateRecordedFilterSelection(runtime, trace, site, created)
	}
	return validateRecordedElements(created, trace.ReadsDropped)
}

func validateRecordedElements(created CompositionConstruction, dropped int) error {
	cursor := created.ReadStart
	for i, element := range created.Elements {
		if element.InputIndex != nil || element.ReadStart != cursor || element.ReadEnd < cursor || element.ReadEnd > created.ReadEnd || element.ReadsDropped < 0 || element.ReadsDropped > dropped {
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

func validateRecordedFilterSelection(runtime CompositionRuntime, trace CompositionTrace, site CompositionConstructionSite, created CompositionConstruction) error {
	filterSite, err := filterInputReadSite(site, runtime.ReadSites)
	if err != nil {
		return err
	}
	inputReads, inputLength := 0, 0
	for i := created.ReadStart; i < created.ReadEnd; i++ {
		read := trace.Reads[i]
		if read.SiteIndex != filterSite {
			continue
		}
		if read.Index != inputReads || inputReads > 0 && read.Length != inputLength {
			return fmt.Errorf("recorded filter source read order differs")
		}
		inputLength = read.Length
		inputReads++
	}
	if created.Completed && trace.ReadsDropped == 0 && inputReads != inputLength {
		return fmt.Errorf("recorded filter omitted a source read")
	}
	cursor, previousInput := created.ReadStart, -1
	for _, element := range created.Elements {
		if element.InputIndex == nil || *element.InputIndex <= previousInput || *element.InputIndex >= 256 ||
			element.ReadStart < cursor || element.ReadStart >= element.ReadEnd || element.ReadEnd > created.ReadEnd ||
			element.ReadsDropped < 0 || element.ReadsDropped > trace.ReadsDropped ||
			trace.Reads[element.ReadStart].SiteIndex != filterSite || trace.Reads[element.ReadStart].Index != *element.InputIndex || !element.Completed {
			return fmt.Errorf("recorded filter output source differs")
		}
		cursor, previousInput = element.ReadEnd, *element.InputIndex
	}
	return nil
}
