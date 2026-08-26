package bindingcoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"strconv"
)

func generatedPartitionID(partition Partition, ordinal int) string {
	seed := partition.BindingID + "\x00" + partition.Polarity + "\x00" + partition.Stage + "\x00" + partition.Reason
	sum := sha256.Sum256([]byte(seed))
	return "urn:bindingcoverage:partition/" + hex.EncodeToString(sum[:8]) + "/" + strconv.Itoa(ordinal)
}
func expectedProductionReason(name, oracleReason string) (production.Reason, error) {
	switch name {
	case "selective-ci-9-binding-complete":
		return requireReason(oracleReason, "COMPLETE", production.ReasonComplete)
	case "selective-ci-9-missing-lane-registry-mismatch":
		return requireReason(oracleReason, "MISSING_MISMATCH", production.ReasonMissingMismatch)
	case "missing-match":
		return requireReason(oracleReason, "MISSING_MATCH", production.ReasonMissingMatch)
	case "zero-denominator":
		return requireReason(oracleReason, "ZERO_DENOMINATOR", production.ReasonZeroDenominator)
	case "duplicate-binding-id":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonDuplicateID)
	case "duplicate-binding-polarity":
		return requireReason(oracleReason, "DUPLICATE_PARTITION_POLARITY", production.ReasonDuplicatePolarity)
	case "dangling-binding":
		return requireReason(oracleReason, "DANGLING_PARTITION", production.ReasonUnknownReference)
	case "bad-stable-id":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidID)
	case "bad-snapshot-digest":
		return requireReason(oracleReason, "STALE_OR_BAD_DIGEST", production.ReasonInvalidDigest)
	case "unsupported-kind":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidEnum)
	case "bad-token":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonInvalidToken)
	case "stale-partition-metadata":
		return requireReason(oracleReason, "STALE_PARTITION", production.ReasonStalePartition)
	case "unknown-schema":
		return requireReason(oracleReason, "UNKNOWN_SCHEMA", production.ReasonUnknownSchema)
	case "complete-two-bindings", "permuted-complete-two-bindings", "overflow-boundary", "precedence-metadata-as-data", "shared-endpoint-references":
		return requireReason(oracleReason, "COMPLETE", production.ReasonComplete)
	case "valid-but-unequal-snapshot-digests":
		return requireReason(oracleReason, "STALE_OR_BAD_DIGEST", production.ReasonSnapshotMismatch)
	case "self-link-binding":
		return requireReason(oracleReason, "UNKNOWN_BINDING", production.ReasonSelfLink)
	default:
		return "", fmt.Errorf("unregistered corpus case %q", name)
	}
}
func requireReason(actual, expected string, productionReason production.Reason) (production.Reason, error) {
	if actual != expected {
		return "", fmt.Errorf("oracle reason %q, expected %q", actual, expected)
	}
	return productionReason, nil
}
func productionInputReceipt(input production.Input, output production.Output) (uint64, string, error) {
	canonical, err := marshalProductionInput(input)
	if err != nil {
		return 0, "", err
	}
	sum := sha256.Sum256(canonical)
	digest := hex.EncodeToString(sum[:])
	if output.InputBytes != uint64(len(canonical)) || output.InputDigest != digest {
		return uint64(len(canonical)), digest, fmt.Errorf("production input receipt mismatch: output=%d/%s recomputed=%d/%s", output.InputBytes, output.InputDigest, len(canonical), digest)
	}
	return uint64(len(canonical)), digest, nil
}
