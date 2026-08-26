package languagedelivery

func bindProfileObligation(items []Obligation) []Obligation {
	matched := 0
	for index := range items {
		if items[index].ID != "TOOL-PROFILER" {
			continue
		}
		items[index].Evidence = rule(SourceProfile, EvidenceProfile, "", "profiles", 2)
		matched++
	}
	if matched != 1 {
		panic("languagedelivery: TOOL-PROFILER denominator drift")
	}
	return bindDebugObligation(items)
}
