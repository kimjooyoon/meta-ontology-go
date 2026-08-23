package languagesemantic

func evaluateFlowStep21(flow *evaluateFlowState) {
	for _, definition = range flow.slot01.Cases {
		if definition.Kind == CaseSource {
			continue
		}
		result := CaseResult{Definition: definition}
		switch definition.Kind {
		case CaseLaw:
			if flow.slot15 != nil {
				result.Status = StatusUnresolved
				result.Evidence.Error = flow.slot15.Error()
			} else {
				satisfied := lawSatisfied(definition.Law, flow.slot14)
				result.Evidence.Law = &LawEvidence{Law: definition.Law, Satisfied: satisfied, Observation: flow.slot14}
				if satisfied {
					result.Status = StatusSatisfied
				} else {
					result.Status = StatusNotSatisfied
				}
			}
		case CaseUpstreamRejection:
			item, ok := flow.slot16[definition.UpstreamCase]
			if !ok {
				result.Status = StatusUnresolved
				result.Evidence.Error = "upstream syntax case is missing"
			} else {
				evidence := &UpstreamEvidence{CaseID: definition.UpstreamCase, ObservedDecision: item.Evidence.ObservedDecision, Diagnostics: append([]string(nil), item.Evidence.Diagnostics...)}
				result.Evidence.Upstream = evidence
				if item.Status == "SATISFIED" && item.Evidence.ObservedDecision == "FAIL_CLOSED" {
					result.Status = StatusSatisfied
				} else {
					result.Status = StatusNotSatisfied
				}
			}
		}
		result.Digest = caseDigest(result)
		flow.slot07 = append(flow.slot07, result)
	}
}
