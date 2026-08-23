package operationconformance

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type BehavioralCase struct {
	ID                        string   `json:"id"`
	Mutation                  string   `json:"mutation"`
	ExpectedDecision          Decision `json:"expected_decision"`
	ExpectedIndicator         string   `json:"expected_indicator,omitempty"`
	ExpectedIndicatorDecision Decision `json:"expected_indicator_decision,omitempty"`
}

type BehavioralCorpus struct {
	Schema, ContractID, Version string
	CaseCount                   int              `json:"case_count"`
	Cases                       []BehavioralCase `json:"cases"`
}

var fixedBehavioralCases = []BehavioralCase{
	{"baseline-pass", "NONE", DecisionPass, "", ""},
	{"atomic-direct-write", "DIRECT_WRITE", DecisionBlock, fixedIndicators[0].ID, DecisionFail},
	{"filename-domain-drift", "FILENAME_DOMAIN", DecisionBlock, fixedIndicators[1].ID, DecisionFail},
	{"header-byte-drift", "HEADER", DecisionBlock, fixedIndicators[2].ID, DecisionFail},
	{"import-identity-drift", "IMPORT", DecisionBlock, fixedIndicators[3].ID, DecisionFail},
	{"initialization-order-drift", "ORDER", DecisionBlock, fixedIndicators[4].ID, DecisionFail},
	{"package-name-drift", "PACKAGE", DecisionBlock, fixedIndicators[5].ID, DecisionFail},
	{"incomplete-evidence", "EVIDENCE_MISSING", DecisionBlock, fixedIndicators[0].ID, DecisionUnknown},
}

func ParseBehavioralCorpus(raw []byte) (BehavioralCorpus, error) {
	var corpus BehavioralCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		return corpus, err
	}
	if corpus.Schema != BehavioralCorpusSchema || corpus.ContractID != ContractID ||
		corpus.Version != BehavioralCorpusVersion || corpus.CaseCount != 8 ||
		!reflect.DeepEqual(corpus.Cases, fixedBehavioralCases) {
		return corpus, fmt.Errorf("FAIL_CLOSED: SplitGo behavioral corpus drift")
	}
	return corpus, nil
}
