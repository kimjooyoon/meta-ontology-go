package languagecomparison

import "slices"

func summarize(requested int, samples []Sample) Summary {
	goooWalls, goWalls := []int64{}, []int64{}
	goooAllocations, goAllocations := []uint64{}, []uint64{}
	goooDigests, goDigests := map[string]bool{}, map[string]bool{}
	equivalent := 0
	for _, sample := range samples {
		goooWalls = append(goooWalls, sample.Gooo.WallNanoseconds)
		goWalls = append(goWalls, sample.Go.WallNanoseconds)
		goooAllocations = append(goooAllocations, sample.Gooo.TotalAllocBytes)
		goAllocations = append(goAllocations, sample.Go.TotalAllocBytes)
		if sample.GoooOutputDigest != "" {
			goooDigests[sample.GoooOutputDigest] = true
		}
		if sample.GoOutputDigest != "" {
			goDigests[sample.GoOutputDigest] = true
		}
		if sample.GoooOutputDigest != "" && sample.GoooOutputDigest == sample.GoOutputDigest &&
			sample.GoooDecision == "PASS" && sample.GoDecision == "PASS" {
			equivalent++
		}
	}
	gooo := summarizeSide(goooWalls, goooAllocations)
	baseline := summarizeSide(goWalls, goAllocations)
	return Summary{SamplesRequested: requested, SamplesObserved: len(samples), EquivalentOutputSamples: equivalent,
		GoooOutputDigestVariants: len(goooDigests), GoOutputDigestVariants: len(goDigests), Gooo: gooo, Go: baseline,
		GoToGoooWallRatioPPM: ratioPPM(uint64(max(0, baseline.WallMedianNanoseconds)),
			uint64(max(0, gooo.WallMedianNanoseconds))),
		GoToGoooAllocRatioPPM: ratioPPM(baseline.TotalAllocMedianBytes, gooo.TotalAllocMedianBytes)}
}

func summarizeSide(walls []int64, allocations []uint64) SideSummary {
	slices.Sort(walls)
	slices.Sort(allocations)
	if len(walls) == 0 || len(allocations) == 0 {
		return SideSummary{}
	}
	return SideSummary{WallMinNanoseconds: walls[0], WallMedianNanoseconds: walls[len(walls)/2],
		WallMaxNanoseconds: walls[len(walls)-1], TotalAllocMinBytes: allocations[0],
		TotalAllocMedianBytes: allocations[len(allocations)/2], TotalAllocMaxBytes: allocations[len(allocations)-1]}
}

func ratioPPM(numerator, denominator uint64) uint64 {
	if numerator == 0 || denominator == 0 {
		return 0
	}
	return numerator * 1_000_000 / denominator
}
