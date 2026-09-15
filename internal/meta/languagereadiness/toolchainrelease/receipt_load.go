package toolchainrelease

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadEvidence(directory string) ([]PlatformEvidence, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	evidence := []PlatformEvidence{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".receipt.json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		receipt, err := decodeStrict[PlatformReceipt](raw)
		if err != nil {
			return nil, err
		}
		if err := validateTopDecision(receipt); err != nil {
			return nil, err
		}
		item := PlatformEvidence{Receipt: receipt, ReceiptVerified: receiptDigestValid(receipt)}
		item.ArchivePath, item.ArchiveVerified = verifyReceiptArchive(directory, receipt)
		evidence = append(evidence, item)
	}
	sort.Slice(evidence, func(i, j int) bool {
		return evidence[i].Receipt.Platform.ID < evidence[j].Receipt.Platform.ID
	})
	return evidence, nil
}

func validateTopDecision(receipt PlatformReceipt) error {
	if receipt.Decision == DecisionFailClosed {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_PLATFORM_FAILED")
	}
	if receipt.Decision != DecisionPass {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_DECISION_UNKNOWN")
	}
	if receipt.Resolution != ResolutionExact {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_RESOLUTION_LOWERED")
	}
	return nil
}
