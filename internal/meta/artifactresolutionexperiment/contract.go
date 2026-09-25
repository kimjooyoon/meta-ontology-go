package artifactresolutionexperiment

const ContractSchema = "gooo/artifact-resolution-experiment-contract/v1"
const ReportSchema = "gooo/artifact-resolution-experiment-report/v1"
const ExpectedIndicators = 13
const ExpectedNonClaims = 4

type Contract struct {
	Schema               string   `json:"schema"`
	ID                   string   `json:"id"`
	ManifestSchema       string   `json:"manifest_schema"`
	InterfaceSchema      string   `json:"interface_schema"`
	ManifestDefinitions  int      `json:"manifest_definitions"`
	InterfaceDefinitions int      `json:"interface_definitions"`
	RegisteredEmitters   int      `json:"registered_emitters"`
	Indicators           int      `json:"indicators"`
	NotClaimedCount      int      `json:"not_claimed_count"`
	NotClaimed           []string `json:"not_claimed"`
}
