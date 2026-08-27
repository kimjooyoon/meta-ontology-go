package audienceresolution

import "reflect"

const (
	ContractSchema = "gooo/audience-resolution-contract/v1"
	LedgerSchema   = "gooo/audience-resolution-ledger/v1"
	ReceiptSchema  = "gooo/audience-resolution-receipt/v1"
	SourceKind     = "gooo"
	Subject        = "audience-resolution-fixture"
	IndicatorTotal = 12
)

type Contract struct {
	Schema                 string             `json:"schema"`
	ID                     string             `json:"id"`
	SourcePath             string             `json:"source_path"`
	SourceDeclarationCount int                `json:"source_declaration_count"`
	Fixed                  FixedDenominator   `json:"fixed"`
	Audiences              []AudienceContract `json:"audiences"`
	NotClaimed             []string           `json:"not_claimed"`
}

type FixedDenominator struct {
	Indicators      int            `json:"indicators"`
	Records         int            `json:"records"`
	Counterexamples int            `json:"counterexamples"`
	Classes         map[string]int `json:"classes"`
	ProofChoices    map[string]int `json:"proof_choices"`
}

type AudienceContract struct {
	Audience    string   `json:"audience"`
	Resolution  string   `json:"resolution"`
	Coordinates []string `json:"coordinates"`
}

// CanonicalContract is the only denominator accepted by this experiment.
// The coordinate sets are deliberately nested: each higher audience sees more
// of the same ledger, never a different decision.
func CanonicalContract() Contract {
	user := []string{"ledger.coverage", "user.coordinates", "projection.shared-decision", "projection.resolution"}
	author := append(append([]string{}, user...), "source.binding", "ledger.replay", "author.coordinates", "governor.coordinates")
	governor := append(append([]string{}, author...), "projection.nesting", "counterexample.omission", "counterexample.contradiction", "receipt.seal")
	return Contract{
		Schema: ContractSchema, ID: "audience-resolution-v1",
		SourcePath: "examples/audience-resolution/main.gooo", SourceDeclarationCount: 22,
		Fixed: FixedDenominator{Indicators: IndicatorTotal, Records: IndicatorTotal, Counterexamples: 2,
			Classes:      map[string]int{"OUTCOME": 4, "DRIVER": 5, "GUARDRAIL": 3},
			ProofChoices: map[string]int{"FOUNDATION": 6, "COHERENCE": 3, "REGRESSION": 3}},
		Audiences: []AudienceContract{
			{Audience: "USER", Resolution: "USER_VISIBLE_COORDINATES", Coordinates: user},
			{Audience: "TOOL_AUTHOR", Resolution: "TOOL_CONTRACT_COORDINATES", Coordinates: author},
			{Audience: "GOVERNOR", Resolution: "GOVERNOR_FULL_LEDGER", Coordinates: governor},
		},
		NotClaimed: []string{
			"confidentiality against covert channels",
			"human comprehension",
			"authorization enforcement",
			"semantic correctness of producer claims",
			"cross-run performance determinism",
		},
	}
}

func ContractValid(value Contract) bool { return reflect.DeepEqual(value, CanonicalContract()) }
