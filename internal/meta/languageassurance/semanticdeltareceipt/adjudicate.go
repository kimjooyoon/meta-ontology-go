package semanticdeltareceipt

import "reflect"

// Adjudicate is deliberately independent of Produce's decision. It parses the
// raw pair again, rebuilds the three layers, and only then checks the receipt.
func Adjudicate(input Input, receipt Receipt) Verdict {
	verdict := Verdict{Producer: receipt.Producer, Consumer: Consumer, Stage: "adjudicate", Step: "replay-receipt"}
	beforeSource, beforeErr := independentProject(input.Before)
	afterSource, afterErr := independentProject(input.After)
	text := textualDelta(input.Before, input.After)
	if receipt.Schema != ReceiptSchema || receipt.Producer != Producer || receipt.Consumer != Consumer ||
		receipt.MetaOperation != MetaOperation || receipt.SubjectSHA != input.SubjectSHA || receipt.CaseID != input.CaseID ||
		receipt.Before.SourceDigest != digestBytes(input.Before) || receipt.After.SourceDigest != digestBytes(input.After) ||
		receipt.TextualDelta != text || receipt.RepositoryWrites != 0 || !receiptDigestValid(receipt) {
		return mismatchVerdict()
	}
	if beforeErr != nil || afterErr != nil {
		verdict.Decision, verdict.Resolution, verdict.Classification, verdict.Reason = DecisionFailClosed, ResolutionUnknown, ClassIndeterminate, ReasonUnavailable
		verdict.Passed = receipt.Decision == verdict.Decision && receipt.Resolution == verdict.Resolution &&
			receipt.Classification == verdict.Classification && receipt.Reason == verdict.Reason &&
			receipt.StructuralDelta.Status == "UNKNOWN" && receipt.SemanticClaimDelta.Status == "UNKNOWN"
		return verdict
	}
	structural, err := structuralDelta(beforeSource, afterSource)
	if err != nil {
		return mismatchVerdict()
	}
	claims := claimDelta(beforeSource, afterSource)
	classification, decision, resolution, reason := ClassPreserved, DecisionFixedPoint, ResolutionExact, ReasonTextualOnly
	if hasSemanticDelta(structural, claims) {
		classification, decision, reason = ClassChanged, DecisionDelta, ReasonMeaning
	}
	verdict.Decision, verdict.Resolution, verdict.Classification, verdict.Reason = decision, resolution, classification, reason
	verdict.Passed = receipt.Decision == decision && receipt.Resolution == resolution && receipt.Classification == classification &&
		receipt.Reason == reason && reflect.DeepEqual(receipt.StructuralDelta, structural) &&
		reflect.DeepEqual(receipt.SemanticClaimDelta, claims) && reflect.DeepEqual(receipt.ClaimTransitions, transitions(beforeSource, afterSource, classification, reason))
	return verdict
}

func receiptDigestValid(receipt Receipt) bool {
	digest := receipt.ReceiptDigest
	copy := receipt
	copy.ReceiptDigest = ""
	return digest != "" && digest == digestValue(copy)
}

func independentProject(raw []byte) (projectedSource, error) {
	// This parser uses declaration slices and positional signatures instead of
	// the producer's field scanner. It is intentionally small and conservative.
	lines := splitIndependentLines(raw)
	entities := map[string]string{}
	activities := make([]activityDecl, 0)
	hasPackage, hasNamespace := false, false
	for _, line := range lines {
		if line == "" || line[0] == '/' {
			continue
		}
		if len(line) >= 8 && line[:8] == "package " {
			hasPackage = true
			continue
		}
		if len(line) >= 10 && line[:10] == "namespace " {
			hasNamespace = true
			continue
		}
		if len(line) >= 7 && line[:7] == "entity " {
			name, id, ok := independentEntity(line)
			if !ok || hasEntity(entities, name) {
				return projectedSource{}, errIndependentSource
			}
			entities[name] = id
			continue
		}
		if len(line) >= 9 && line[:9] == "activity " {
			activity, ok := independentActivity(line)
			if !ok {
				return projectedSource{}, errIndependentSource
			}
			activities = append(activities, activity)
			continue
		}
		return projectedSource{}, errIndependentSource
	}
	if !hasPackage || !hasNamespace || len(entities) == 0 || len(activities) == 0 {
		return projectedSource{}, errIndependentSource
	}
	result := projectedSource{}
	for name, id := range entities {
		result.nodes = append(result.nodes, Node{ID: id, Kind: "ENTITY"})
		_ = name
	}
	for _, activity := range activities {
		id := activityID(activity.name)
		result.nodes = append(result.nodes, Node{ID: id, Kind: "ACTIVITY"})
		for index, input := range activity.inputs {
			entityID, ok := entities[input]
			if !ok {
				return projectedSource{}, errIndependentSource
			}
			fact := Fact{Subject: id, Predicate: "gooo:uses", Object: entityID}
			result.facts = append(result.facts, fact)
			result.claims = append(result.claims, claimFor(id, "uses", index, fact))
		}
		entityID, ok := entities[activity.output]
		if !ok {
			return projectedSource{}, errIndependentSource
		}
		fact := Fact{Subject: id, Predicate: "gooo:generates", Object: entityID}
		result.facts = append(result.facts, fact)
		result.claims = append(result.claims, claimFor(id, "generates", 0, fact))
	}
	sortProjected(&result)
	return result, nil
}

func hasEntity(entities map[string]string, name string) bool {
	_, exists := entities[name]
	return exists
}
