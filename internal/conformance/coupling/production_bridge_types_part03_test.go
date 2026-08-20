//go:build detector_bridge

package coupling

type bridgeSubjectVector struct {
	Oracle   oracleBridgeVector
	Producer productionVector
}

func oracleBridgeVectorFromResult(output Output) oracleBridgeVector {
	return oracleBridgeVector{
		Schema: output.Schema, InputDigest: output.InputDigest, Decision: output.Decision, Reason: output.Reason,
		AcceptedSurfaces: append([]string(nil), output.AcceptedSurfaces...), ChangedSurfaces: append([]string(nil), output.ChangedSurfaces...), ReceiptSurfaces: append([]string(nil), output.ReceiptSurfaces...),
		SemanticBeforeDigest: output.SemanticBeforeDigest, SemanticAfterDigest: output.SemanticAfterDigest, SemanticDeltaDigest: output.SemanticDeltaDigest, PathClosureDigest: output.PathClosureDigest,
		ObservationCounts: output.ObservationCounts, Resources: output.Resources, CanonicalOutputDigest: output.CanonicalOutputDigest, ReplayDigest: output.ReplayDigest,
	}
}
