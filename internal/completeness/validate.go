package completeness

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Validate checks internal accounting and source binding. Evidence strings are
// references, not semantic proof of a product or an external permission grant.
func Validate(r *CompletenessReceipt) error {
	if r == nil || r.Schema != CompletenessReceiptSchema || r.ProfileID == "" || r.Decision == "" || r.DecisionBasis == "" {
		return fmt.Errorf("receipt schema or profile is unbound")
	}
	if !DeclarationMatchesProjection() {
		return fmt.Errorf("receipt declaration and generated structure differ")
	}
	if r.AggregateCompletenessScore != nil {
		return fmt.Errorf("completeness is not an aggregate score")
	}
	actual, err := json.Marshal(r.Scope["receipt_declaration"])
	if err != nil {
		return err
	}
	expected, err := json.Marshal(ContractBinding())
	if err != nil {
		return err
	}
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("receipt declaration binding differs")
	}
	counts := map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	ids := map[string]bool{}
	unresolved := 0
	for _, d := range r.Dimensions {
		if _, ok := counts[d.Status]; !ok {
			return fmt.Errorf("unknown dimension state %q", d.Status)
		}
		if d.ID == "" || ids[d.ID] || d.Reason == "" || d.Unit == "" || len(d.Evidence) == 0 {
			return fmt.Errorf("dimension identity or evidence is incomplete")
		}
		if d.Numerator < 0 || d.Denominator < 0 || d.Numerator > d.Denominator {
			return fmt.Errorf("dimension counters are invalid")
		}
		ids[d.ID] = true
		counts[d.Status]++
		if d.Status == "PASS" {
			if d.Denominator == 0 || d.Numerator != d.Denominator {
				return fmt.Errorf("PASS has no complete observed denominator")
			}
			continue
		}
		if unresolved >= len(r.UnresolvedClaims) {
			return fmt.Errorf("unresolved claim is missing")
		}
		c := r.UnresolvedClaims[unresolved]
		if c.ID != d.ID || c.Status != d.Status || c.Reason != d.Reason || c.NextOperation == "" {
			return fmt.Errorf("unresolved claim changed identity, state, cause or next operation")
		}
		unresolved++
	}
	if len(ids) == 0 || len(r.CoreDimensions) == 0 || unresolved != len(r.UnresolvedClaims) {
		return fmt.Errorf("receipt dimensions or unresolved frontier are incomplete")
	}
	coreIDs := map[string]bool{}
	for _, id := range r.CoreDimensions {
		if !ids[id] || coreIDs[id] {
			return fmt.Errorf("core dimension is absent or duplicated")
		}
		coreIDs[id] = true
	}
	if len(counts) != len(r.StatusCounts) {
		return fmt.Errorf("status count keys differ")
	}
	for k, v := range counts {
		if r.StatusCounts[k] != v {
			return fmt.Errorf("status count %s differs", k)
		}
	}
	if counts["FAIL_CLOSED"] > 0 && r.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("receipt decision hides a failed axis")
	}
	if r.Decision == "FAIL_CLOSED" && (r.FailClosedReason == nil || *r.FailClosedReason == "") {
		return fmt.Errorf("failed receipt lost its cause")
	}
	if r.Decision != "FAIL_CLOSED" && r.FailClosedReason != nil {
		return fmt.Errorf("receipt decision contradicts its failure cause")
	}
	if unresolved == 0 {
		if r.FirstUnresolved != nil {
			return fmt.Errorf("spurious first unresolved claim")
		}
	} else if r.FirstUnresolved == nil || *r.FirstUnresolved != r.UnresolvedClaims[0] {
		return fmt.Errorf("first unresolved stage changed")
	}
	return nil
}
