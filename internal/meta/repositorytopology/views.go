package repositorytopology

func (s *inspection) summary(satisfied int) Summary {
	return Summary{
		Coordinates: Counter{Satisfied: satisfied, Total: 10, BasisPoints: satisfied * 1000},
		Rows: RowSummary{FilesObserved: len(s.source.Files), FilesExact: s.fileRowsExact, DirectoriesObserved: len(s.source.Directories), DirectoriesExact: s.directoryRowsExact, DuplicatePaths: s.duplicates},
		Languages: LanguageSummary{GoFiles: s.goFiles, GoooFiles: s.goooFiles, GoLines: s.goLines, GoooLines: s.goooLines},
		Meta: MetaSummary{Indicators: len(s.source.Meta.Indicators), BoundIndicators: s.metaBound, BindingWitnesses: s.bindingWitnesses, UnknownDecisions: s.unknownDecisions},
		Root: RootSummary{TopologyExemptions: s.rootTopology, READMEExemptions: s.rootREADME},
	}
}

func (s *inspection) rootLanguageExact(language string) bool {
	for _, directory := range s.source.Directories {
		if directory.Path != "." {
			continue
		}
		if language == "go" {
			return directory.GoFiles == s.goFiles && directory.GoLines == s.goLines
		}
		return directory.GoooFiles == s.goooFiles && directory.GoooLines == s.goooLines
	}
	return false
}

func buildViews(indicators []Indicator) []AudienceView {
	specs := []struct{ audience, resolution string; ids []string }{
		{"READER", "COUNT_ROWS", []string{"rows.files", "rows.directories", "aggregate.go", "aggregate.gooo"}},
		{"IMPLEMENTER", "META_BOUND", []string{"ontology.binding", "rows.files", "rows.directories", "aggregate.go", "aggregate.gooo", "meta.binding", "vocabulary.decisions"}},
		{"GOVERNOR", "FULL_RECEIPT", indicatorIDs(indicators)},
	}
	views := make([]AudienceView, 0, len(specs))
	for _, spec := range specs {
		satisfied := 0
		for _, id := range spec.ids {
			for _, indicator := range indicators {
				if indicator.ID == id && indicator.Satisfied { satisfied++ }
			}
		}
		decision := "FAIL_CLOSED"
		if satisfied == len(spec.ids) { decision = "PASS" }
		views = append(views, AudienceView{Audience: spec.audience, Decision: decision, Resolution: spec.resolution, Satisfied: satisfied, Total: len(spec.ids), BasisPoints: satisfied * 10000 / len(spec.ids), IndicatorIDs: spec.ids})
	}
	return views
}

func indicatorIDs(indicators []Indicator) []string {
	ids := make([]string, len(indicators))
	for i, indicator := range indicators { ids[i] = indicator.ID }
	return ids
}

func proofs() []Proof {
	return []Proof{
		{Choice: "FOUNDATION", Claim: "exact head, ontology, and explicit root exemptions", Evidence: "source identity plus two ontology digests"},
		{Choice: "COHERENCE", Claim: "file rows, directory rows, language sums, and meta bindings agree", Evidence: "complete source metric row joins"},
		{Choice: "REGRESSION", Claim: "duplicates and repository mutation authority remain zero", Evidence: "negative CI cases plus read-only command boundary"},
	}
}
