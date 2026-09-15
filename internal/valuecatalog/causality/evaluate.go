package causality

func Evaluate(input []byte, expectedMode string) (Receipt, error) {
	report, mode, err := parseInputReport(input, expectedMode)
	if err != nil {
		return Receipt{}, err
	}

	registered := report.OperationClaimTransitions[:ClaimTotal]
	resolved := report.OperationClaimTransitions[ClaimTotal:]
	claimIDs := make([]string, ClaimTotal)
	for index := range registered {
		claimIDs[index] = registered[index].ClaimID
	}
	graph, err := buildGraph(claimIDs)
	if err != nil {
		return Receipt{}, err
	}

	unavailable := make([]bool, ClaimTotal)
	for index := range resolved {
		unavailable[index] = isUnavailableEvent(resolved[index].Event)
	}
	resolutions := make([]Resolution, 0, ClaimTotal)
	for index, transition := range resolved {
		if !unavailable[index] {
			resolutions = append(resolutions, Resolution{ClaimID: claimIDs[index], Axis: claimAxes[index], State: "DISCHARGED", Kind: "EVIDENCE_ACCEPTED", ObservedEvent: transition.Event, Coordinate: transition.Coordinate})
			continue
		}

		blockingEdges := incomingUnavailableEdges(graph, claimIDs[index], unavailable, claimIDs)
		if len(blockingEdges) == 0 {
			coordinate := nonEmptyCoordinate(transition.Coordinate)
			resolutions = append(resolutions, Resolution{ClaimID: claimIDs[index], Axis: claimAxes[index], State: "OPEN", Kind: "DIRECT_MISSING", ObservedEvent: transition.Event, Coordinate: coordinate, MissingEvidenceIDs: []string{missingEvidenceID(claimIDs[index])}, CausePath: []string{claimIDs[index]}, CauseTransitionDigest: transition.TransitionDigest, CauseCoordinate: &coordinate})
			continue
		}

		pathIndexes := primaryCausePath(index, unavailable)
		rootIndex := pathIndexes[0]
		causeCoordinate := nonEmptyCoordinate(resolved[rootIndex].Coordinate)
		resolution := Resolution{ClaimID: claimIDs[index], Axis: claimAxes[index], State: "OPEN", Kind: "DEPENDENCY_BLOCKED", ObservedEvent: transition.Event, Coordinate: Coordinate{Stage: "RESOLVE_DEPENDENCY", Step: claimAxes[index], Reason: "UPSTREAM_CLAIM_OPEN"}, MissingEvidenceIDs: []string{missingEvidenceID(claimIDs[rootIndex])}, CauseTransitionDigest: resolved[rootIndex].TransitionDigest, CauseCoordinate: &causeCoordinate}
		for _, edge := range blockingEdges {
			resolution.BlockedByClaimIDs = append(resolution.BlockedByClaimIDs, edge.FromClaimID)
			resolution.BlockedByEdgeIDs = append(resolution.BlockedByEdgeIDs, edge.EdgeID)
		}
		for _, pathIndex := range pathIndexes {
			resolution.CausePath = append(resolution.CausePath, claimIDs[pathIndex])
		}
		resolutions = append(resolutions, resolution)
	}

	metrics := deriveMetrics(graph, resolutions)
	receipt := Receipt{Schema: ReceiptSchema, Scope: ReceiptScope, Subject: Subject{InputReportSchema: report.Schema, InputReportDigest: digestBytes(input), TransitionHeadDigest: report.transitionHead(), GraphDigest: graph.Digest, SourceDigest: report.SourceDigest, SemanticIRDigest: report.CoreIRFingerprint, BindingStatus: "PARTIAL_UNKNOWN", MissingBindingEvidence: bindingEvidence(report.SourceDigest, report.CoreIRFingerprint)}, Graph: graph, Metrics: metrics, Indicators: buildIndicators(metrics), Resolutions: resolutions, Decision: decisionFor(mode, metrics)}
	receipt.ReceiptDigest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
