package languagesemantic

type syntaxCase struct {
	Definition struct {
		ID string `json:"id"`
	} `json:"definition"`
	Evidence struct {
		ObservedDecision string   `json:"observed_decision"`
		Diagnostics      []string `json:"diagnostics"`
	} `json:"evidence"`
	Status string `json:"status"`
}
