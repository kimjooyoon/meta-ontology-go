package query

import "encoding/json"

type canonicalDerivedRule struct {
	Schema string        `json:"schema"`
	Rule   DerivedRuleID `json:"rule"`
}

// CanonicalJSON returns the versioned rule identity used by replay receipts.
func (rule DerivedRuleID) CanonicalJSON() ([]byte, error) {
	normalized, err := ParseDerivedRule(rule)
	if err != nil {
		return nil, err
	}
	return json.Marshal(canonicalDerivedRule{
		Schema: DerivedRuleSchemaVersion,
		Rule:   normalized,
	})
}

// CanonicalDigest hashes the normalized versioned rule identity.
func (rule DerivedRuleID) CanonicalDigest() (string, error) {
	canonical, err := rule.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}
