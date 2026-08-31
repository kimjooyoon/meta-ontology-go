package languagedebugexperiment

type View struct {
	Audience     string   `json:"audience"`
	Resolution   string   `json:"resolution"`
	Satisfied    int      `json:"satisfied"`
	Total        int      `json:"total"`
	BasisPoints  int      `json:"basis_points"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Report struct {
	Schema              string               `json:"schema"`
	SubjectSHA          string               `json:"subject_sha"`
	Decision            string               `json:"decision"`
	Reason              string               `json:"reason"`
	Resolution          string               `json:"resolution"`
	Summary             Summary              `json:"summary"`
	Replay              ReplayEvidence       `json:"replay"`
	RuntimeObservations []RuntimeObservation `json:"runtime_observations"`
	Build               Measurement          `json:"build"`
	EvaluatorBuild      Measurement          `json:"evaluator_build"`
	Test                Measurement          `json:"test"`
	Graph               GraphObservation     `json:"graph"`
	Indicators          []Indicator          `json:"indicators"`
	Views               []View               `json:"views"`
	UnknownCases        []Uncertainty        `json:"unknown_cases,omitempty"`
	RefutedCases        []Refutation         `json:"refuted_cases,omitempty"`
	RepositoryWrites    int                  `json:"repository_writes"`
	MutationAuthority   bool                 `json:"mutation_authority"`
	Digest              string               `json:"digest"`
}

func buildViews(indicators []Indicator) []View {
	return []View{
		buildView("USER", "USER_VISIBLE", indicators[:4]),
		buildView("TOOL_AUTHOR", "TOOL_CONTRACT", indicators[:9]),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators),
	}
}

func buildView(audience, resolution string, indicators []Indicator) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indicators)}
	for _, indicator := range indicators {
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	view.BasisPoints = basisPoints(view.Satisfied, view.Total)
	return view
}
