package transformationeffectverification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/transformationeffect"
)

func deriveCausalUnknownProjection(report generation.ReceiptReport) (transformationeffect.CausalUnknownProjection, error) {
	projection := transformationeffect.CausalUnknownProjection{Records: []transformationeffect.CausalUnknownRecord{}}
	failures, err := verificationFailureIndex(report.Failures)
	if err != nil {
		return transformationeffect.CausalUnknownProjection{}, err
	}
	seen := make(map[string]bool, len(report.Unknowns))
	for _, unknown := range report.Unknowns {
		if err := validateVerificationCausalUnknown(unknown); err != nil {
			return transformationeffect.CausalUnknownProjection{}, err
		}
		key := unknown.ActionIndicatorID + "\x00" + unknown.RequiredIndicatorID
		if seen[key] {
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("unknown obligation is duplicated")
		}
		seen[key] = true
		failure, hasFailure := failures[unknown.ActionIndicatorID]
		switch unknown.UnknownClass {
		case generation.ReceiptUnknownClassDirectMissing,
			generation.ReceiptUnknownClassMalformedEvidence,
			generation.ReceiptUnknownClassUnexpectedEvidence:
			if len(unknown.BlockedBy) != 0 {
				return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("direct unknown has a frontier")
			}
			projection.DirectUnknownCount++
		case generation.ReceiptUnknownClassDependencyBlocked:
			root := "operation-failure:" + unknown.ActionIndicatorID
			if !hasFailure || failure.Stage != unknown.Stage || failure.Step != unknown.Step ||
				failure.Reason != string(unknown.Reason) || failure.NextOperation != unknown.NextOperation ||
				len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != root {
				return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("dependency unknown is not bound")
			}
			projection.DependencyBlockedUnknownCount++
		default:
			return transformationeffect.CausalUnknownProjection{}, fmt.Errorf("unknown class is unsupported")
		}
		projection.Records = append(projection.Records, verificationCausalUnknownRecord(unknown))
	}
	sort.Slice(projection.Records, func(left, right int) bool {
		return causalUnknownKey(projection.Records[left]) < causalUnknownKey(projection.Records[right])
	})
	payload, err := json.Marshal(projection)
	if err != nil {
		return transformationeffect.CausalUnknownProjection{}, err
	}
	digest := sha256.Sum256(payload)
	projection.Digest = hex.EncodeToString(digest[:])
	return projection, nil
}
