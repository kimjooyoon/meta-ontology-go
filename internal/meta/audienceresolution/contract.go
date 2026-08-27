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
	Schema     string             `json:"schema"`
	ID         string             `json:"id"`
	SourcePath string             `json:"source_path"`
	Fixed      FixedDenominator   `json:"fixed"`
	Audiences  []AudienceContract `json:"audiences"`
	NotClaimed []string           `json:"not_claimed"`
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

// CanonicalContract describes the fixture's fixed report shape. The semantic
// policy itself is not authoritative here: it is derived from the .gooo
// source after syntax.ParseFile -> bidir.Lower. Keeping this JSON contract
// separate makes a source policy intervention observable in the projections.
func CanonicalContract() Contract {
	user := []string{"ledger.coverage", "user.coordinates", "projection.shared-decision", "projection.resolution"}
	author := append(append([]string{}, user...), "source.binding", "ledger.replay", "author.coordinates", "governor.coordinates")
	governor := append(append([]string{}, author...), "projection.nesting", "counterexample.omission", "counterexample.contradiction", "receipt.seal")
	return Contract{
		Schema: ContractSchema, ID: "audience-resolution-v1",
		SourcePath: "examples/audience-resolution/main.gooo",
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

func ContractValid(value Contract) bool {
	want := CanonicalContract()
	return value.Schema == want.Schema && value.ID == want.ID &&
		value.SourcePath == want.SourcePath && reflect.DeepEqual(value.Fixed, want.Fixed) &&
		reflect.DeepEqual(value.NotClaimed, want.NotClaimed)
}
