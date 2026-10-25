package main

import (
	"encoding/json"
	"fmt"
)

// The raw proposal hash remains in the observed process output. Replay binds
// its address-normalized proposal, so disposable directories do not become
// semantic identity. Every other receipt field remains in the projection.
func canonicalGoooSplitReceipt(root string, raw []byte, proposalDigest string) (json.RawMessage, error) {
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(canonicalProcessBytes(root, raw), &receipt); err != nil {
		return nil, err
	}
	var application map[string]json.RawMessage
	if err := json.Unmarshal(receipt["application"], &application); err != nil {
		return nil, err
	}
	if application == nil || !validBareDigest(trimDigestPrefix(proposalDigest)) {
		return nil, fmt.Errorf("replay requires an application and exact normalized proposal digest")
	}
	application["proposal_sha256"], _ = json.Marshal(proposalDigest)
	encoded, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}
	receipt["application"] = encoded
	return json.Marshal(receipt)
}
