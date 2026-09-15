package ciplanusecase

func applyProfileSummary(summary *Summary, input Input) {
	profiles := map[string]ProfileSample{}
	for _, sample := range input.Profile.Samples {
		profiles[sample.CaseID] = sample
	}
	for _, spec := range input.Contract.Cases {
		sample, ok := profiles[spec.ID]
		if !ok {
			continue
		}
		summary.ResourceSamples++
		if sample.WallMS > summary.MaxWallMS {
			summary.MaxWallMS = sample.WallMS
		}
		if sample.PeakRSSKiB > summary.MaxPeakRSSKiB {
			summary.MaxPeakRSSKiB = sample.PeakRSSKiB
		}
		if sample.ReceiptBytes > summary.MaxReceiptBytes {
			summary.MaxReceiptBytes = sample.ReceiptBytes
		}
	}
	if len(input.Profile.Samples) != input.Contract.Denominator {
		summary.ResourceSamples = -1
	}
}
