package languagedelivery

func bindDebugObligation(items []Obligation) []Obligation {
	matched := 0
	for index := range items {
		if items[index].ID != "TOOL-DEBUGGER" {
			continue
		}
		items[index].Evidence = rule(SourceDebug, EvidenceDebug, "", "paused_sessions", 2)
		matched++
	}
	if matched != 1 {
		panic("languagedelivery: TOOL-DEBUGGER denominator drift")
	}
	return items
}
