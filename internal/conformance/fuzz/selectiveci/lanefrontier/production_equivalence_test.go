package lanefrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
)

type semanticVector struct {
	SchemaVersion     string   `json:"schema_version"`
	Decision          string   `json:"decision"`
	Reason            string   `json:"reason"`
	RegistryDigest    string   `json:"registry_digest"`
	BaseSHA           string   `json:"base_sha"`
	LaneHeadSHA       string   `json:"lane_head_sha"`
	LaneID            string   `json:"lane_id"`
	RegisteredBranch  string   `json:"registered_branch"`
	OwnedPathPrefixes []string `json:"owned_path_prefixes"`
	ChangedPaths      []string `json:"changed_paths"`
	AheadCount        int64    `json:"ahead_count"`
	BehindCount       int64    `json:"behind_count"`
	OpenPRCount       int64    `json:"open_pr_count"`
	ActiveLeaseCount  int64    `json:"active_lease_count"`
}

type pairedReceipt struct {
	CaseID                     string         `json:"case_id"`
	SchemaIdentical            bool           `json:"schema_identical"`
	OracleCanonicalDigest      string         `json:"oracle_canonical_digest"`
	ProductionDigest           string         `json:"production_canonical_digest"`
	OracleVector               semanticVector `json:"oracle_vector"`
	ProductionVector           semanticVector `json:"production_vector"`
	OraclePermutationEqual     bool           `json:"oracle_permutation_equal"`
	ProductionPermutationEqual bool           `json:"production_permutation_equal"`
}

func TestProductionLaneFrontierEquivalence(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 16 {
		t.Fatalf("lane frontier corpus cases=%d, want 16", len(corpus.Cases))
	}
	seen := map[string]bool{}
	receipts := make([]pairedReceipt, 0, len(corpus.Cases))
	mismatches := 0
	for _, fixture := range corpus.Cases {
		if seen[fixture.Name] {
			t.Fatalf("duplicate lane frontier case %q", fixture.Name)
		}
		seen[fixture.Name] = true
		receipt, err := compareCase(fixture)
		if err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s error=%v", fixture.Name, err)
			continue
		}
		receipts = append(receipts, receipt)
		if err := validatePartition(receipt.OracleVector); err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s partition=%v", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector), err)
		}
		if err := validatePartition(receipt.ProductionVector); err != nil {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s partition=%v", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector), err)
		}
		if !reflect.DeepEqual(receipt.OracleVector, receipt.ProductionVector) || !receipt.OraclePermutationEqual || !receipt.ProductionPermutationEqual {
			mismatches++
			t.Errorf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO case=%s oracle=%s production=%s", fixture.Name, vectorJSON(receipt.OracleVector), vectorJSON(receipt.ProductionVector))
		}
	}
	if len(seen) != len(corpus.Cases) || len(receipts)+mismatches < len(corpus.Cases) {
		t.Fatalf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO consumed=%d receipts=%d cases=%d", len(seen), len(receipts), len(corpus.Cases))
	}
	digest := pairedReceiptDigest(receipts)
	t.Logf("lane frontier paired receipt digest=%s mismatch_count=%d", digest, mismatches)
	if mismatches != 0 {
		t.Fatalf("LANE_FRONTIER_COUNTEREXAMPLE_NO_GO mismatch_count=%d paired_receipt_digest=%s", mismatches, digest)
	}
}

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

func productionSHA(value string) string {
	if value == "" {
		return ""
	}
	if len(value) == 40 || len(value) == 64 {
		if _, err := hex.DecodeString(value); err == nil {
			return value
		}
	}
	sum := sha256.Sum256([]byte("sha:" + value))
	return hex.EncodeToString(sum[:20])
}

func productionLaneID(value string) string {
	if value == "" {
		return ""
	}
	return "urn:lane-frontier:translated/" + productionDigest(value)[:16]
}

func normalizedSchema(schema string) string {
	if schema == SchemaV1 {
		return production.SchemaVersion
	}
	return schema
}

func validatePartition(vector semanticVector) error {
	unknownReasons := map[string]bool{string(UnknownSchema): true, string(MissingInput): true, string(InvalidCount): true, string(AmbiguousOwner): true}
	if vector.Decision == string(Unknown) && !unknownReasons[vector.Reason] {
		return fmt.Errorf("UNKNOWN reason %q is outside the four allowed classes", vector.Reason)
	}
	if vector.Decision == string(Eligible) && vector.Reason != string(production.ReasonEligible) {
		return fmt.Errorf("ELIGIBLE reason %q is not ELIGIBLE", vector.Reason)
	}
	if vector.Decision != string(Unknown) && vector.Decision != string(Ineligible) && vector.Decision != string(Eligible) {
		return fmt.Errorf("unknown decision %q", vector.Decision)
	}
	return nil
}

func permutedCase(fixture Case) Case {
	permuted := fixture
	permuted.Input.OwnedPathPrefixes = reverseCopy(fixture.Input.OwnedPathPrefixes)
	permuted.Input.ChangedPaths = reverseCopy(fixture.Input.ChangedPaths)
	return permuted
}

func reverseCopy(values []string) []string {
	result := append([]string(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func vectorJSON(vector semanticVector) string {
	data, _ := json.Marshal(vector)
	return string(data)
}

func pairedReceiptDigest(receipts []pairedReceipt) string {
	data, _ := json.Marshal(struct {
		Cases []pairedReceipt `json:"cases"`
	}{receipts})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
