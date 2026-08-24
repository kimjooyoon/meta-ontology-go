package main

import "fmt"

func reconstruct(subjectSHA string, input transaction) (rawReceipt, error) {
	if !validSHA(subjectSHA) || input.Schema != transactionSchema || input.TransactionID == "" || len(input.RawReconstructions) != 0 {
		return rawReceipt{}, fmt.Errorf("raw reconstruction input identity is malformed")
	}
	selfMinting, roleConflicts := countSelfMinting(input.AuthorityRoutes), countRoleConflicts(input.RoleBindings)
	unknownLaundering, unknownTop := countUnknown(input.DecisionTransitions)
	snapshotBPS, snapshotPaths, err := observeSnapshots(subjectSHA, input.SnapshotBindings)
	if err != nil { return rawReceipt{}, err }
	authorityObserved, rolesObserved, decisionsObserved := len(input.AuthorityRoutes) > 0, len(input.RoleBindings) > 0, len(input.DecisionTransitions) > 0
	evidenceObserved := boolInt(authorityObserved) + boolInt(rolesObserved) + boolInt(decisionsObserved)
	observation := rawObservation{
		EvidenceGroupsObserved: evidenceObserved, EvidenceGroupsTotal: 3,
		SelfMintingPaths: observed(authorityObserved, selfMinting), RoleConflictPaths: observed(rolesObserved, roleConflicts),
		UnknownLaunderingPaths: observed(decisionsObserved, unknownLaundering), UnknownTopDecisions: observed(decisionsObserved, unknownTop),
		SnapshotBindingsObserved: len(input.SnapshotBindings), SnapshotBindingsRequired: len(snapshotEvidenceIDs),
		ExactSnapshotBindingBPS: snapshotBPS, SnapshotMismatchPaths: snapshotPaths,
	}
	observation.CandidateDecision, observation.CandidateReason, observation.CandidateResolution = decide(observation)
	input.RawReconstructions = nil
	transactionDigest, err := digest(input)
	if err != nil { return rawReceipt{}, err }
	return rawReceipt{Schema: receiptSchema, VerifierID: verifierID, SubjectSHA: subjectSHA, DenominatorDigest: denominatorDigest, RawTransactionDigest: transactionDigest, Observation: observation}, nil
}

func decide(observation rawObservation) (string, string, string) {
	if observation.EvidenceGroupsObserved != observation.EvidenceGroupsTotal || observation.SelfMintingPaths == nil || observation.RoleConflictPaths == nil || observation.UnknownLaunderingPaths == nil || observation.ExactSnapshotBindingBPS == nil {
		return failClosed, reasonEvidence, unknown
	}
	if observation.UnknownTopDecisions != nil && *observation.UnknownTopDecisions > 0 { return failClosed, reasonUnknown, invariantOnly }
	if *observation.ExactSnapshotBindingBPS < 10000 { return block, reasonSnapshot, exact }
	if *observation.SelfMintingPaths > 0 || *observation.RoleConflictPaths > 0 || *observation.UnknownLaunderingPaths > 0 { return block, reasonGovernance, exact }
	return allowLimited, reasonClear, exact
}

func boolInt(value bool) int { if value { return 1 }; return 0 }
