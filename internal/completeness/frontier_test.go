package completeness

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLaterFailureCannotReplaceEarlierUnknownFrontier(t *testing.T) {
	r := testReceipt()
	reason := "later observed boundary failure"
	r.Dimensions = append(r.Dimensions, CompletenessDimension{ID: "execution_boundary", Status: "FAIL_CLOSED", Denominator: 1, Unit: "executions", Reason: reason, Evidence: []string{"independent negative observation"}})
	r.UnresolvedClaims = append(r.UnresolvedClaims, UnresolvedCompletenessClaim{ID: "execution_boundary", Status: "FAIL_CLOSED", Reason: reason, NextOperation: "preserve failed execution"})
	r.StatusCounts["FAIL_CLOSED"]++
	if Validate(&r) == nil {
		t.Fatal("progress decision hid a failure")
	}
	r.Decision = "FAIL_CLOSED"
	r.FailClosedReason = &reason
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if observed.FirstUnresolved.ID != "reverse_observation_coverage" || observed.FirstUnresolved.Status != "UNKNOWN" {
		t.Fatal("later failure erased first unknown stage")
	}
}

func TestDecodeBoundsNestedScope(t *testing.T) {
	r := testReceipt()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	deep := strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)
	data = []byte(strings.Replace(string(data), `"large_integer":9007199254740993`, `"large_integer":`+deep, 1))
	if _, err := Decode(data); err == nil {
		t.Fatal("over-depth scope accepted")
	}
}

func TestDecodeRejectsCaseInsensitiveOverwritingAlias(t *testing.T) {
	r := testReceipt()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), `"profile_id":`, `"PROFILE_ID":"rewritten","profile_id":`, 1))
	if _, err := Decode(data); err == nil {
		t.Fatal("extra case-insensitive alias accepted outside the declared schema")
	}
}
