package replay

func lawsFlowStep25(flow *lawsFlowState) {
	{
		flow.result0, flow.result1 = LawObservation{AnchorPath: flow.slot00, PresentationChanged: flow.slot02.Canonical() != flow.slot06.Canonical(), PresentationInvariant: flow.slot02.StableHash() == flow.slot06.StableHash(), CandidateRecorded: len(flow.slot09.Graph.Candidates()) == 1, CandidateNonAuthoritative: flow.slot08.StableHash() == flow.slot09.StableHash(), DeterministicRecorded: len(flow.slot11.Graph.DeterministicFacts()) == 1, DeterministicAuthoritative: flow.slot08.StableHash() != flow.slot11.StableHash(), StructureSemanticHash: flow.slot08.StableHash(), PresentationSemanticHash: flow.slot06.StableHash(), CandidateSemanticHash: flow.slot09.StableHash(), DeterministicSemanticHash: flow.slot11.StableHash(), CandidateCanonicalChanged: flow.slot08.Canonical() != flow.slot09.Canonical(), DeterministicCanonicalChanged: flow.slot08.Canonical() != flow.slot11.Canonical()}, nil
		flow.done = true
		return
	}
}
