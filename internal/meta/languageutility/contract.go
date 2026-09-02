package languageutility

import "fmt"

const (
	ContractSchema    = "gooo/language-utility-contract/v1"
	ObservationSchema = "gooo/language-utility-observation/v1"
	ReportSchema      = "gooo/language-utility-report/v1"
	MetaOperation     = "measure-language-utility"
)

var CanonicalStages = []StageSpec{
	{ID: "SOURCE_PRESENT", ProofChoice: "foundation"},
	{ID: "SYNTAX_ACCEPTED", ProofChoice: "foundation"},
	{ID: "SEMANTIC_ACCEPTED", ProofChoice: "coherence"},
	{ID: "OUTCOME_OBSERVED", ProofChoice: "coherence"},
	{ID: "DETERMINISTIC_REPLAY", ProofChoice: "regression"},
	{ID: "RESOURCE_OBSERVED", ProofChoice: "regression"},
	{ID: "USER_ARTIFACT_VERIFIED", ProofChoice: "coherence"},
}

type StageSpec struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}

type UseCaseSpec struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Floors struct {
	ClosedCells      int `json:"closed_cells"`
	CompleteUseCases int `json:"complete_use_cases"`
}

type Contract struct {
	Schema     string        `json:"schema"`
	ID         string        `json:"id"`
	Stages     []StageSpec   `json:"stages"`
	UseCases   []UseCaseSpec `json:"use_cases"`
	Floors     Floors        `json:"floors"`
	NotClaimed []string      `json:"not_claimed"`
}

func ValidateContract(value Contract) error {
	if value.Schema != ContractSchema || value.ID == "" {
		return fmt.Errorf("language utility contract identity is invalid")
	}
	if len(value.Stages) != len(CanonicalStages) || len(value.UseCases) != 6 {
		return fmt.Errorf("language utility denominator must be 6 x 7")
	}
	for index, stage := range value.Stages {
		if stage != CanonicalStages[index] {
			return fmt.Errorf("stage %d is not canonical", index)
		}
	}
	seen := map[string]bool{}
	for _, useCase := range value.UseCases {
		if useCase.ID == "" || useCase.Label == "" || seen[useCase.ID] {
			return fmt.Errorf("use case %q is invalid", useCase.ID)
		}
		seen[useCase.ID] = true
	}
	if value.Floors.ClosedCells < 0 || value.Floors.ClosedCells > 42 ||
		value.Floors.CompleteUseCases < 0 || value.Floors.CompleteUseCases > 6 {
		return fmt.Errorf("language utility floors are invalid")
	}
	return nil
}
