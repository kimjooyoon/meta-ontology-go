package toolchaincli

import cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"

func summarize(source Source, cases []CaseResult, registryDrift int) Summary {
	summary := Summary{Total: FixedTotal, RegistryDrift: registryDrift}
	if validDigest(source.ExecutableDigest) {
		summary.BinaryBindings = 1
	}
	for _, result := range cases {
		summary.Invocations += result.Invocations
		summary.StructuredOutputs += result.StructuredOutputs
		summary.LanguageOperations += result.LanguageOperations
		summary.DeclaredCommands += result.DeclaredCommands
		summary.RepositoryWrites += result.RepositoryWrites
		switch result.Status {
		case "SATISFIED":
			summary.Satisfied++
		case "UNRESOLVED":
			summary.Unresolved++
			continue
		default:
			summary.NotSatisfied++
		}
		summary.Executed++
		if result.Status == "SATISFIED" && result.Definition.Kind == "POSITIVE" {
			summary.PositivePaths++
		}
		if result.Status == "SATISFIED" && result.Definition.Kind == "GUARDRAIL" {
			summary.GuardrailRejections++
		}
		if result.ReplayMatched {
			summary.ReplayMatches++
		} else {
			summary.ReplayMismatches++
		}
		if !result.ExitMatched {
			summary.ExitMismatches++
		}
		if !result.StdoutMatched {
			summary.StdoutMismatches++
		}
		if !result.StderrMatched {
			summary.StderrMismatches++
		}
		for _, observation := range []cliruntime.Observation{result.First, result.Replay} {
			if observation.PeakRSSKiB > summary.PeakRSSKiB {
				summary.PeakRSSKiB = observation.PeakRSSKiB
			}
			if observation.PeakRSSKiB > 0 {
				summary.ResourceObservations++
			}
		}
	}
	summary.ReadinessBPS = summary.Satisfied * 10000 / FixedTotal
	return summary
}
