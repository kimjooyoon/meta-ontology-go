package toolchainrelease

import "sort"

func groupEvidence(evidence []PlatformEvidence, expectedHead string,
	summary *Summary) (map[string][]PlatformEvidence, int) {
	expected := map[string]bool{}
	for _, target := range targetRegistry {
		expected[target.ID] = true
	}
	grouped := map[string][]PlatformEvidence{}
	unexpected := 0
	for _, item := range evidence {
		if item.Receipt.HeadSHA != expectedHead {
			summary.HeadMismatches++
		}
		id := item.Receipt.Platform.ID
		if !expected[id] {
			unexpected++
			continue
		}
		grouped[id] = append(grouped[id], item)
	}
	return grouped, unexpected
}

func uniquePlatforms(grouped map[string][]PlatformEvidence) (int, int) {
	systems, architectures := map[string]bool{}, map[string]bool{}
	for _, items := range grouped {
		if len(items) == 1 {
			systems[items[0].Receipt.Platform.GOOS] = true
			architectures[items[0].Receipt.Platform.GOARCH] = true
		}
	}
	return len(systems), len(architectures)
}

func receiptSetDigest(evidence []PlatformEvidence) string {
	digests := make([]string, 0, len(evidence))
	for _, item := range evidence {
		digests = append(digests, item.Receipt.ReceiptDigest)
	}
	sort.Strings(digests)
	digest, _ := digestValue(digests)
	return digest
}
