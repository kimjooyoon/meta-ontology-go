package opentofuobservation

import "fmt"

const (
	ContractSchema     = "gooo/opentofu-observation-contract/v1"
	ObservationSchema  = "gooo/opentofu-observation/v1"
	ReportSchema       = "gooo/opentofu-observation-report/v1"
	MetaOperation      = "observe-released-opentofu-cli"
	ExpectedReleaseID  = "opentofu-v1.12.6"
	ExpectedAssetURL   = "https://github.com/opentofu/opentofu/releases/download/v1.12.6/tofu_1.12.6_linux_amd64.tar.gz"
	ExpectedAssetSHA   = "sha256:50a6106fa4de523d09c87af85f3db1dd47535fc005727fdca6852146476b88ec"
	ExpectedAssetSize  = 34646566
	ExpectedSumsSHA    = "sha256:6988e0cb8f4e9ebfa3b0999e44841549741b22d9b38873cb5b89074f1cddcb1c"
	ExpectedGo         = "go1.27.0"
	DecisionPass       = "PASS"
	DecisionUnknown    = "UNKNOWN"
	DecisionRefuted    = "REFUTED"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"
)

type CellSpec struct {
	ID            string `json:"id"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Indicator     string `json:"indicator"`
}

var fixedCells = []CellSpec{
	{"OPENTOFU_RELEASE_PIN", "PinOpenTofuRelease", "FOUNDATION", "DRIVER"},
	{"ASSET_CHECKSUM", "VerifyOpenTofuAssetChecksum", "FOUNDATION", "GUARDRAIL"},
	{"CLI_VERSION_JSON", "ObserveOpenTofuVersionJSON", "FOUNDATION", "OUTCOME"},
	{"FIXTURE_INPUT_DIGEST", "PinOpenTofuFixtureInput", "FOUNDATION", "DRIVER"},
	{"PLAN_JSON_SCHEMA", "ValidateOpenTofuPlanJSON", "COHERENCE", "OUTCOME"},
	{"TEST_EVENT_INVENTORY", "AccountOpenTofuTestEvents", "COHERENCE", "OUTCOME"},
	{"COMMAND_RUNTIME_RECEIPT", "RecordOpenTofuCommandRuntime", "COHERENCE", "DRIVER"},
	{"PEAK_RSS_OBSERVATION", "ObserveOpenTofuPeakRSS", "COHERENCE", "DRIVER"},
	{"DETERMINISTIC_PLAN_REPLAY", "ReplayOpenTofuPlanJSON", "REGRESSION", "GUARDRAIL"},
	{"DETERMINISTIC_TEST_REPLAY", "ReplayOpenTofuTestEvents", "REGRESSION", "GUARDRAIL"},
	{"REUSE_ELIGIBILITY", "EvaluateOpenTofuTestReuse", "REGRESSION", "GUARDRAIL"},
	{"HUMAN_REPORT", "PublishOpenTofuObservationReport", "REGRESSION", "OUTCOME"},
}

var fixedPaths = []string{"P1 RELEASE_IDENTITY", "P2 PLAN_JSON", "P3 TEST_JSON"}

func Cells() []CellSpec { return append([]CellSpec(nil), fixedCells...) }

func Paths() []string { return append([]string(nil), fixedPaths...) }

func ValidateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.ID == "" || len(contract.Cells) != len(fixedCells) {
		return fmt.Errorf("OpenTofu contract identity or denominator is invalid")
	}
	for index, cell := range contract.Cells {
		if cell != fixedCells[index] {
			return fmt.Errorf("OpenTofu contract cell %d is not canonical", index)
		}
	}
	return nil
}

type Contract struct {
	Schema       string     `json:"schema"`
	ID           string     `json:"id"`
	Cells        []CellSpec `json:"cells"`
	NotClaimed   []string   `json:"not_claimed"`
	GraphProgram string     `json:"graph_program"`
}
