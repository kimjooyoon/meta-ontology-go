package languagedelivery

func userObligations() []Obligation {
	items := baseUserObligations()
	matched := 0
	for index := range items {
		if items[index].ID != "USER-MULTI-FILE-EXECUTION" {
			continue
		}
		items[index].Evidence = rule(SourceUserJourney, EvidenceJourney, "run-package", "", 1)
		matched++
	}
	if matched != 1 {
		panic("languagedelivery: USER-MULTI-FILE-EXECUTION denominator drift")
	}
	return items
}
