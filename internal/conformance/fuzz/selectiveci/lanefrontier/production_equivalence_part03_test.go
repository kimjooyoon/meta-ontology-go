package lanefrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
)

func compareCase(fixture Case) (pairedReceipt, error) {
	oracleResult := Evaluate(fixture.Input)
	productionInput := translateInput(fixture.Input)
	data, err := json.Marshal(productionInput)
	if err != nil {
		return pairedReceipt{}, err
	}
	productionOutput := production.ClassifyJSON(data)
	permuted := permutedCase(fixture)
	oraclePermutation := bytes.Equal(CanonicalJSON(fixture), CanonicalJSON(permuted)) && Evaluate(fixture.Input).CanonicalDigest == Evaluate(permuted.Input).CanonicalDigest
	permutedData, err := json.Marshal(translateInput(permuted.Input))
	if err != nil {
		return pairedReceipt{}, err
	}
	leftOutput, err := production.EncodeJSON(productionOutput)
	if err != nil {
		return pairedReceipt{}, err
	}
	rightOutput, err := production.EncodeJSON(production.ClassifyJSON(permutedData))
	if err != nil {
		return pairedReceipt{}, err
	}
	return pairedReceipt{CaseID: fixture.Name, SchemaIdentical: fixture.Input.Schema == productionInput.SchemaVersion, OracleCanonicalDigest: oracleResult.CanonicalDigest, ProductionDigest: productionOutput.CanonicalDigest, OracleVector: oracleVector(fixture.Input, oracleResult), ProductionVector: productionVector(fixture.Input, productionOutput), OraclePermutationEqual: oraclePermutation, ProductionPermutationEqual: bytes.Equal(leftOutput, rightOutput)}, nil
}
func translateInput(input Input) production.Input {
	schema := input.Schema
	if schema == SchemaV1 {
		schema = production.SchemaVersion
	}
	return production.Input{SchemaVersion: schema, RegistryDigest: productionDigest(input.RegistryDigest), BaseSHA: productionSHA(input.BaseSHA), LaneHeadSHA: productionSHA(input.LaneHeadSHA), LaneID: productionLaneID(input.LaneStableID), RegisteredBranch: input.RegisteredBranch, OwnedPathPrefixes: append([]string(nil), input.OwnedPathPrefixes...), ChangedPaths: append([]string(nil), input.ChangedPaths...), AheadCount: input.AheadCount, BehindCount: input.BehindCount, OpenPRCount: input.OpenPRCount, ActiveLeaseCount: input.ActiveLeaseCount}
}
func oracleVector(input Input, result Result) semanticVector {
	normalized := normalizedInput(input)
	reason := string(result.Reason)
	if result.Reason == Clean {
		reason = string(production.ReasonEligible)
	}
	return semanticVector{SchemaVersion: normalizedSchema(input.Schema), Decision: string(result.Decision), Reason: reason, RegistryDigest: input.RegistryDigest, BaseSHA: input.BaseSHA, LaneHeadSHA: input.LaneHeadSHA, LaneID: input.LaneStableID, RegisteredBranch: input.RegisteredBranch, OwnedPathPrefixes: normalized.OwnedPathPrefixes, ChangedPaths: normalized.ChangedPaths, AheadCount: input.AheadCount, BehindCount: input.BehindCount, OpenPRCount: input.OpenPRCount, ActiveLeaseCount: input.ActiveLeaseCount}
}
func productionVector(source Input, output production.Output) semanticVector {
	normalized := normalizedInput(source)
	return semanticVector{SchemaVersion: output.SchemaVersion, Decision: string(output.Decision), Reason: string(output.Reason), RegistryDigest: source.RegistryDigest, BaseSHA: source.BaseSHA, LaneHeadSHA: source.LaneHeadSHA, LaneID: source.LaneStableID, RegisteredBranch: source.RegisteredBranch, OwnedPathPrefixes: normalized.OwnedPathPrefixes, ChangedPaths: normalized.ChangedPaths, AheadCount: output.AheadCount, BehindCount: output.BehindCount, OpenPRCount: output.OpenPRCount, ActiveLeaseCount: output.ActiveLeaseCount}
}
func productionDigest(value string) string {
	if value == "" {
		return ""
	}
	if len(value) == 64 {
		if _, err := hex.DecodeString(value); err == nil {
			return value
		}
	}
	sum := sha256.Sum256([]byte("registry:" + value))
	return hex.EncodeToString(sum[:])
}
