package repositorytopology

func (s *inspection) buildReport(sourceJSON, rootOntology, bindingOntology []byte) Report {
	indicators := []Indicator{
		metric("subject.identity", "OUTCOME", "FOUNDATION", s.identityExact, boolInt(s.identityExact), 1),
		metric("ontology.binding", "DRIVER", "FOUNDATION", s.ontologyExact, boolInt(s.ontologyExact), 1),
		metric("rows.files", "DRIVER", "COHERENCE", s.fileRowsExact == len(s.source.Files), s.fileRowsExact, len(s.source.Files)),
		metric("rows.directories", "DRIVER", "COHERENCE", s.directoryRowsExact == len(s.source.Directories), s.directoryRowsExact, len(s.source.Directories)),
		metric("aggregate.go", "DRIVER", "COHERENCE", s.rootLanguageExact("go"), boolInt(s.rootLanguageExact("go")), 1),
		metric("aggregate.gooo", "DRIVER", "COHERENCE", s.rootLanguageExact("gooo"), boolInt(s.rootLanguageExact("gooo")), 1),
		metric("root.exemptions", "GUARDRAIL", "FOUNDATION", s.rootPolicyExact && s.rootTopology == 2 && s.rootREADME == 1, s.rootTopology+s.rootREADME, 3),
		metric("meta.binding", "GUARDRAIL", "COHERENCE", s.metaBound == len(s.source.Meta.Indicators) && s.bindingWitnesses == 1, s.metaBound, len(s.source.Meta.Indicators)),
		metric("vocabulary.decisions", "GUARDRAIL", "FOUNDATION", s.unknownDecisions == 0, s.unknownDecisions, 0),
		metric("mutation.free", "GUARDRAIL", "REGRESSION", s.duplicates == 0, s.duplicates, 0),
	}
	satisfied, failures := 0, []string{}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			satisfied++
		} else {
			failures = append(failures, indicator.ID)
		}
	}
	report := Report{
		Schema: "gooo/repository-topology-receipt/v1", ExecutionPolicy: "READ_ONLY_EXACT_HEAD",
		Repository: s.source.Repository, CommitSHA: s.source.CommitSHA,
		Summary: s.summary(satisfied), Files: s.source.Files, Directories: s.source.Directories,
		Indicators: indicators, Views: buildViews(indicators), Proofs: proofs(), Failures: failures,
		SourceMetricsDigest: digestBytes(sourceJSON), RootOntologyDigest: digestBytes(rootOntology),
		BindingOntologyDigest: digestBytes(bindingOntology), MutationAuthority: false,
	}
	report.Status, report.Decision, report.Resolution, report.Reason = "PASS", "PASS", "EXACT", "REPOSITORY_TOPOLOGY_EXACT"
	if s.lowerResolution {
		report.Status, report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "FAIL_CLOSED", "LOWER_RESOLUTION", "REPOSITORY_TOPOLOGY_DECISION_UNKNOWN"
	} else if satisfied != len(indicators) {
		report.Status, report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "FAIL_CLOSED", "INVARIANT_ONLY", "REPOSITORY_TOPOLOGY_INVARIANT_MISMATCH"
	}
	return report
}

func metric(id, class, proof string, satisfied bool, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, Satisfied: satisfied, Observed: observed, Expected: expected}
}

func boolInt(value bool) int { if value { return 1 }; return 0 }
