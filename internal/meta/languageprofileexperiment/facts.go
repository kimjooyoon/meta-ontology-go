package languageprofileexperiment

import (
	"reflect"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

type observedFacts struct {
	profiles, samples, successful, sourceCoherence     int
	variants, wallObservations, allocationObservations int
	wallMin, wallMedian, wallMax                       int64
	allocMin, allocMedian, allocMax                    uint64
	go127Runtimes, unknownRejections, writes           int
	mutation, nonClaims, executableBound               bool
}

func observeFacts(input Input) observedFacts {
	values := []languageprofile.Receipt{input.First, input.Replay}
	facts := observedFacts{executableBound: validDigest(input.ExecutableDigest)}
	digests := map[string]bool{}
	walls := []int64{}
	allocations := []uint64{}
	for _, receipt := range values {
		if receipt.Decision == "PASS" {
			facts.profiles++
		}
		if strings.HasPrefix(receipt.Runner.GoVersion, input.Contract.RunnerGoPrefix) {
			facts.go127Runtimes++
		}
		facts.samples += len(receipt.Samples)
		for _, sample := range receipt.Samples {
			if sample.Decision == "PASS" {
				facts.successful++
			}
			if sample.ExecutionDigest != "" {
				digests[sample.ExecutionDigest] = true
			}
			if sample.WallNanoseconds > 0 {
				facts.wallObservations++
				walls = append(walls, sample.WallNanoseconds)
			}
			if sample.TotalAllocBytes > 0 {
				facts.allocationObservations++
				allocations = append(allocations, sample.TotalAllocBytes)
			}
		}
		facts.writes += receipt.Effects.RepositoryWrites
		facts.mutation = facts.mutation || receipt.Effects.MutationAuthority
	}
	facts.variants = len(digests)
	if input.First.SourceDigest != "" && input.First.SourceDigest == input.Replay.SourceDigest &&
		input.First.SemanticDigest == input.Replay.SemanticDigest &&
		reflect.DeepEqual(input.First.ProfiledEntry, input.Replay.ProfiledEntry) {
		facts.sourceCoherence = 1
	}
	facts.unknownRejections = boolInt(input.UnknownEntry.Decision == "FAIL_CLOSED" &&
		input.UnknownEntry.Resolution == "EXACT" && input.UnknownEntry.Reason == "SOURCE_ENTRY_UNKNOWN")
	facts.writes += input.UnknownEntry.Effects.RepositoryWrites
	facts.mutation = facts.mutation || input.UnknownEntry.Effects.MutationAuthority
	want := languageprofile.DefaultNonClaims()
	facts.nonClaims = reflect.DeepEqual(input.First.NotClaimed, want) &&
		reflect.DeepEqual(input.Replay.NotClaimed, want) && reflect.DeepEqual(input.UnknownEntry.NotClaimed, want)
	facts.wallMin, facts.wallMedian, facts.wallMax = int64Stats(walls)
	facts.allocMin, facts.allocMedian, facts.allocMax = uint64Stats(allocations)
	return facts
}

func int64Stats(values []int64) (int64, int64, int64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	slices.Sort(values)
	return values[0], values[len(values)/2], values[len(values)-1]
}

func uint64Stats(values []uint64) (uint64, uint64, uint64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	slices.Sort(values)
	return values[0], values[len(values)/2], values[len(values)-1]
}

func validDigest(value string) bool { return len(value) == 71 && strings.HasPrefix(value, "sha256:") }

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
