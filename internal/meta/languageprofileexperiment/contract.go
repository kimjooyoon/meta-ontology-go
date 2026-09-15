package languageprofileexperiment

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

const (
	ContractSchema     = "gooo/language-profile-experiment-contract/v1"
	ReportSchema       = "gooo/language-profile-experiment-report/v1"
	ExpectedIndicators = 13
	ExpectedNonClaims  = 4
)

type Contract struct {
	Schema            string `json:"schema"`
	ID                string `json:"contract_id"`
	ProfileSchema     string `json:"profile_schema"`
	Profiles          int    `json:"profiles"`
	SamplesPerProfile int    `json:"samples_per_profile"`
	RunnerGoPrefix    string `json:"runner_go_prefix"`
	Indicators        int    `json:"indicators"`
	NotClaimedCount   int    `json:"not_claimed_count"`
}

func ExpectedContract() Contract {
	return Contract{Schema: ContractSchema, ID: "billing-language-profile-v1",
		ProfileSchema: languageprofile.ReceiptSchema, Profiles: 2, SamplesPerProfile: 5,
		RunnerGoPrefix: "go1.27", Indicators: ExpectedIndicators, NotClaimedCount: ExpectedNonClaims}
}

func contractReason(contract Contract) string {
	if !reflect.DeepEqual(contract, ExpectedContract()) {
		return "PROFILE_EXPERIMENT_CONTRACT_DRIFT"
	}
	return ""
}
