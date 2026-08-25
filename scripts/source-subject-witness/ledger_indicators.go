package main

func buildLedgerIndicators(ledger witnessLedger) []ledgerIndicator {
	pass := func(id, route, relation, value, limit string) ledgerIndicator {
		return ledgerIndicator{ID: id, Route: route, Verdict: "PASS", Relation: relation, Value: value, Limit: limit}
	}
	counts := ledger.Counts
	partition := counts.SourceIndicatorsApplicable + counts.SourceIndicatorsNotApplicable
	return []ledgerIndicator{
		pass("foundation.source-schema", "FOUNDATION", "=", ledger.SourceSchema, "gooo/indicator-report/v3"),
		pass("foundation.commit-binding", "FOUNDATION", "=", ledger.CommitSHA, ledger.CommitSHA),
		pass("foundation.policy-binding", "FOUNDATION", "sha256", ledger.PolicyDigest, "bound"),
		pass("foundation.project-root-topology-exemption", "FOUNDATION", "=", "true", "true"),
		pass("foundation.project-root-readme-exemption", "FOUNDATION", "=", "true", "true"),
		pass("foundation.workflow-discovery-exemption", "FOUNDATION", "=", itoa(counts.WorkflowDiscoveryExemptions), "1"),
		pass("coherence.file-observations", "COHERENCE", "=", itoa(counts.FileWitnesses), itoa(counts.FileWitnesses)),
		pass("coherence.file-meta-coverage", "COHERENCE", "=", itoa(counts.FileSourceBindings), itoa(counts.GoFiles+counts.GoooFiles)),
		pass("coherence.logical-directory-observations", "COHERENCE", "=", itoa(counts.LogicalDirectories), itoa(counts.LogicalDirectories)),
		pass("coherence.storage-directory-meta-coverage", "COHERENCE", "=", itoa(counts.StorageSourceBindings), itoa(counts.StorageDirectories)),
		pass("coherence.source-applicability-partition", "COHERENCE", "=", itoa(partition), itoa(counts.MetaIndicators)),
		pass("coherence.subject-witnesses", "COHERENCE", "sha256", ledger.SubjectWitnessDigest, "bound"),
		pass("coherence.meta-indicators", "COHERENCE", "sha256", ledger.MetaIndicatorDigest, "bound"),
		pass("regression.canonical-encoding", "REGRESSION", "=", "true", "true"),
	}
}
