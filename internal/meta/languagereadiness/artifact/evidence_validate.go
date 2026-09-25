package artifact

import (
	"encoding/hex"
	"strings"
)

func validPromotionDigest(receipt Receipt) bool {
	return validObligationDigest(
		receipt, "AUTONOMY-CHANGE-PROPOSAL", receipt.ProposalPromotionDigest,
	)
}

func validGuardedCapabilityDigest(receipt Receipt) bool {
	if receipt.Schema == LegacySchema && receipt.GuardedCapabilityDigest != "" {
		return false
	}
	return validObligationDigest(
		receipt, "AUTONOMY-GUARDED-PROMOTION", receipt.GuardedCapabilityDigest,
	)
}

func validObligationDigest(receipt Receipt, obligationID, evidenceDigest string) bool {
	value := strings.TrimPrefix(evidenceDigest, "sha256:")
	_, err := hex.DecodeString(value)
	for _, result := range receipt.Snapshot.Obligations {
		if result.ID == obligationID && result.Status == "SATISFIED" {
			return strings.HasPrefix(evidenceDigest, "sha256:") && len(value) == 64 && err == nil
		}
	}
	return evidenceDigest == ""
}
