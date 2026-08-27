package semanticdeltareceiptconsumer

func classDecision(structural StructuralDelta, claims ClaimDelta) (string, string, string) {
	return classDecisionWithComponents(structural, SemanticComponentDelta{}, claims)
}

func classDecisionWithComponents(structural StructuralDelta, components SemanticComponentDelta, claims ClaimDelta) (string, string, string) {
	if hasSemanticDelta(structural, components, claims) {
		if len(components.Added)+len(components.Removed)+len(components.Changed) > 0 && len(structural.AddedNodes)+len(structural.RemovedNodes)+len(structural.AddedFacts)+len(structural.RemovedFacts)+len(claims.Added)+len(claims.Removed)+len(claims.Changed) == 0 {
			return classChanged, decisionDelta, reasonComponentDelta
		}
		return classChanged, decisionDelta, reasonMeaning
	}
	return classPreserved, decisionFixedPoint, reasonTextualOnly
}

func semanticDecision(class string) string {
	if class == classPreserved {
		return semanticPreserved
	}
	return semanticChanged
}
