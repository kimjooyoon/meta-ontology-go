package languageprofileexperiment

func summarize(input Input, facts observedFacts, indicators []Indicator) Summary {
	satisfied := 0
	for _, item := range indicators {
		if item.Satisfied {
			satisfied++
		}
	}
	return Summary{
		Coordinates: Counter{Satisfied: satisfied, Total: ExpectedIndicators,
			BasisPoints: satisfied * 10000 / ExpectedIndicators},
		Profiles: facts.profiles, Samples: facts.samples, SuccessfulExecutions: facts.successful,
		SourceCoherence: facts.sourceCoherence, ExecutionDigestVariants: facts.variants,
		Resources: ResourceSummary{WallObservations: facts.wallObservations,
			AllocationObservations: facts.allocationObservations,
			WallMinNanoseconds: facts.wallMin, WallMedianNanoseconds: facts.wallMedian, WallMaxNanoseconds: facts.wallMax,
			TotalAllocMinBytes: facts.allocMin, TotalAllocMedianBytes: facts.allocMedian, TotalAllocMaxBytes: facts.allocMax},
		Compiler: CompilerSummary{ExecutableDigest: input.ExecutableDigest, Go127Runtimes: facts.go127Runtimes},
		UnknownEntryRejections: facts.unknownRejections,
		Effects: EffectSummary{RepositoryWrites: facts.writes, MutationAuthority: facts.mutation},
		NotClaimed: len(input.First.NotClaimed), Unknowns: 0,
	}
}
