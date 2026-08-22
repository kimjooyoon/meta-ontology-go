package languageconcept

import "io/fs"

type observation struct {
	concepts, code, useCases, metrics int
	operating, conformed, unbound     int
	novelty, writes                   int
}

func evaluate(repository fs.FS, concepts []Concept) Report {
	report := Report{Schema: ReportSchema, Concepts: concepts}
	seen := make(map[string]bool)
	observed := observation{concepts: len(concepts)}
	for _, item := range concepts {
		bound := item.ID != "" && !seen[item.ID] && item.Problem != "" &&
			item.PositiveEffect != "" && item.MetaOperation != ""
		seen[item.ID] = true
		codeBound, missing := codeBindings(repository, item)
		report.MissingBindings = append(report.MissingBindings, missing...)
		useCaseBound := validUseCases(item.UseCases)
		metricBound := len(item.MetricBindings) > 0
		if codeBound { observed.code++ }
		if useCaseBound { observed.useCases++ }
		if metricBound { observed.metrics++ }
		if item.Stage == "OPERATING" { observed.operating++ } else if item.Stage == "CONFORMED" { observed.conformed++ } else { bound = false }
		if item.NoveltyClaim || item.Rarity != "UNCOMMON_COMBINATION" { observed.novelty++ }
		if !bound || !codeBound || !useCaseBound || !metricBound { observed.unbound++ }
	}
	ready := observed.concepts > 0 && observed.unbound == 0 && observed.novelty == 0 && observed.writes == 0
	if ready {
		report.Decision, report.Reason = "PASS", "LANGUAGE_CONCEPT_CATALOG_BOUND"
	} else {
		report.Decision, report.Reason = "FAIL_CLOSED", "LANGUAGE_CONCEPT_CATALOG_UNBOUND"
	}
	return finish(report, observed, ready)
}

func codeBindings(repository fs.FS, item Concept) (bool, []string) {
	if len(item.CodeBindings) == 0 { return false, []string{item.ID + ":no-code-binding"} }
	missing := make([]string, 0)
	for _, path := range item.CodeBindings {
		if !fs.ValidPath(path) { missing = append(missing, item.ID+":"+path); continue }
		if _, err := fs.Stat(repository, path); err != nil { missing = append(missing, item.ID+":"+path) }
	}
	return len(missing) == 0, missing
}

func validUseCases(cases []UseCase) bool {
	if len(cases) == 0 { return false }
	for _, item := range cases {
		if item.ID == "" || item.Trigger == "" || item.ExpectedOutcome == "" { return false }
	}
	return true
}
