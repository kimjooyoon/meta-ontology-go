package languageprofile

import "slices"

func summarize(requested int, samples []Sample) Summary {
	value := Summary{SamplesRequested: requested, SamplesObserved: len(samples)}
	digests := map[string]bool{}
	walls := make([]int64, 0, len(samples))
	allocations := make([]uint64, 0, len(samples))
	for _, sample := range samples {
		if sample.Decision == "PASS" {
			value.SuccessfulExecutions++
		}
		if sample.ExecutionDigest != "" {
			digests[sample.ExecutionDigest] = true
		}
		if sample.WallNanoseconds > 0 {
			value.WallObservations++
			walls = append(walls, sample.WallNanoseconds)
		}
		if sample.TotalAllocBytes > 0 {
			value.AllocationObservations++
			allocations = append(allocations, sample.TotalAllocBytes)
		}
	}
	value.ExecutionDigestVariants = len(digests)
	if len(walls) > 0 {
		slices.Sort(walls)
		value.WallMinNanoseconds, value.WallMedianNanoseconds = walls[0], walls[len(walls)/2]
		value.WallMaxNanoseconds = walls[len(walls)-1]
	}
	if len(allocations) > 0 {
		slices.Sort(allocations)
		value.TotalAllocMinBytes, value.TotalAllocMedianBytes = allocations[0], allocations[len(allocations)/2]
		value.TotalAllocMaxBytes = allocations[len(allocations)-1]
	}
	return value
}
