package toolchainusecases

func indicators(report Report, registryDrift int) []Indicator {
	summary := report.Summary
	return []Indicator{
		metric("readiness-bps", "OUTCOME", "COHERENCE", report.Resolution, summary.ReadinessBPS, 10000),
		metric("executed-cases", "DRIVER", "FOUNDATION", report.Resolution, summary.Executed, totalCases),
		metric("pass-paths", "DRIVER", "COHERENCE", report.Resolution, summary.PassPaths, 1),
		metric("fail-closed-paths", "DRIVER", "REGRESSION", report.Resolution, summary.FailClosedPaths, 2),
		metric("unresolved.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, summary.Unresolved, 0),
		metric("repository-writes.guardrail", "GUARDRAIL", "REGRESSION", report.Resolution, report.RepositoryWrites, 0),
		metric("mutation-authority.guardrail", "GUARDRAIL", "REGRESSION", report.Resolution, boolInt(report.MutationAuthorized), 0),
		metric("registry-drift.guardrail", "GUARDRAIL", "FOUNDATION", report.Resolution, registryDrift, 0),
	}
}

func metric(id, class, proof, resolution string, value, target int) Indicator {
	return Indicator{MetricID: "gooo.metric.toolchain.executable-use-cases-" + id + ".v1",
		Class: class, ProofChoice: proof, Producer: "toolchainusecases.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: "execute-versioned-use-cases",
		Resolution: resolution, Value: value, Target: target, Satisfied: value == target}
}

func proofs(report Report, registryDrift int) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-use-case-registry-and-artifact",
			EvidenceDigest: digestJSON(report.Source), Passed: registryDrift == 0},
		{Choice: "COHERENCE", MetaOperation: "replay-canonical-success-and-rejections",
			EvidenceDigest: digestJSON(report.Cases), Passed: report.Summary.Satisfied == totalCases},
		{Choice: "REGRESSION", MetaOperation: "reject-tamper-writes-and-authority",
			EvidenceDigest: digestJSON(report.Summary), Passed: report.Summary.FailClosedPaths == 2 &&
				report.RepositoryWrites == 0 && !report.MutationAuthorized},
	}
}
