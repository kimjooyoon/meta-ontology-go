package repositorytopology

import (
	"bytes"
	"strings"
)

func (s *inspection) inspectOntologies(root, binding []byte) {
	rootTokens := []string{"package rootreadme", "namespace rootreadme", "ProjectRoot", "READMEObservation", "FoundationExemption", "ObserveProjectRootREADME", "BindRootREADMEExemption"}
	bindingTokens := []string{"package ", "namespace ", "entity ", "activity "}
	s.ontologyExact = containsAll(root, rootTokens) && containsAll(binding, bindingTokens)
}

func containsAll(source []byte, tokens []string) bool {
	for _, token := range tokens {
		if !bytes.Contains(source, []byte(token)) {
			return false
		}
	}
	return true
}

func (s *inspection) inspectMeta() {
	policy := s.source.Meta.Policy
	s.rootPolicyExact = policy.ExemptProjectRootTopology && policy.ExemptProjectRootREADME
	for _, indicator := range s.source.Meta.Indicators {
		known := knownApplicability(indicator.Applicability) && knownProof(indicator.ProofChoice) && knownDecision(indicator.Decision)
		if !known {
			s.unknownDecisions++
			s.lowerResolution = true
		}
		if indicator.Producer != "" && indicator.Consumer != "" && indicator.MetaOperation != "" && knownProof(indicator.ProofChoice) {
			s.metaBound++
		}
		if indicator.MetricID == "gooo.metric.meta.unbound-indicators.v1" && indicator.Value == 0 && indicator.Satisfied && strings.Contains(indicator.Detail, "examples/meta-binding-coverage/main.gooo") {
			s.bindingWitnesses++
		}
		if indicator.Subject != "." || indicator.Applicability != "NOT_APPLICABLE" {
			continue
		}
		if indicator.Family == "topology" && indicator.ApplicabilityReason == "ROOT_TOPOLOGY_EXEMPT" {
			s.rootTopology++
		}
		if indicator.Family == "documentation" && indicator.ApplicabilityReason == "ROOT_README_EXEMPT" {
			s.rootREADME++
		}
	}
}

func knownApplicability(value string) bool { return value == "APPLICABLE" || value == "NOT_APPLICABLE" }
func knownProof(value string) bool { return value == "foundation" || value == "coherence" || value == "regression" }
func knownDecision(value string) bool { return value == "PASS" || value == "FAIL" || value == "NOT_APPLICABLE" }
