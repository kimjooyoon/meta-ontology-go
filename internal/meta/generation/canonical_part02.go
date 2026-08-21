package generation

import (
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

type canonicalInput struct {
	SchemaVersion      string                   `json:"schema_version"`
	BaseSHA            string                   `json:"base_sha"`
	HeadSHA            string                   `json:"head_sha"`
	Policy             sourcepolicy.Policy      `json:"policy"`
	Indicators         []sourcepolicy.Indicator `json:"indicators"`
	Registry           []Binding                `json:"registry"`
	RequestedK         uint32                   `json:"requested_k"`
	MinimumIndependent uint32                   `json:"minimum_independent"`
}

func finish(plan Plan) Plan {
	if plan.Selected == nil {
		plan.Selected = []Action{}
	}
	if plan.UnselectedIndicatorIDs == nil {
		plan.UnselectedIndicatorIDs = []string{}
	}
	if plan.UnknownIndicatorIDs == nil {
		plan.UnknownIndicatorIDs = []string{}
	}
	if plan.Shortfall == nil {
		plan.Shortfall = []string{}
	}
	sort.Strings(plan.UnselectedIndicatorIDs)
	sort.Strings(plan.UnknownIndicatorIDs)
	sort.Strings(plan.Shortfall)
	unsigned := plan
	unsigned.PlanDigest, unsigned.ReplayDigest = "", ""
	plan.PlanDigest = digestJSON(unsigned)
	plan.ReplayDigest = digestPair(plan.InputDigest, plan.PlanDigest)
	return plan
}

func Encode(plan Plan) ([]byte, error) {
	payload, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func duplicateIndicators(indicators []sourcepolicy.Indicator) bool {
	for index := 1; index < len(indicators); index++ {
		if indicatorKey(indicators[index-1]) == indicatorKey(indicators[index]) {
			return true
		}
	}
	return false
}
