package observereffect

import (
	"bytes"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	observerOutputPolicyOpen  = "OUTPUT_OPEN"
	observerOutputPolicyClose = "OUTPUT_REQUIRES_RESEALED_OBSERVATION"
)

type semanticInterventionCase struct {
	Name        string
	Kind        string
	Mutation    string
	Needle      string
	Replacement string
	Suffix      string
}

var semanticInterventionCases = []semanticInterventionCase{
	{
		Name:        "semantic-output-policy-intervention",
		Kind:        "SEMANTIC_CAUSAL",
		Mutation:    "replace the semantic output policy URI with a reseal-required policy URI",
		Needle:      "gooo://observer-effect/policy/output/open",
		Replacement: "gooo://observer-effect/policy/output/reseal-required",
	},
	{
		Name:     "comment-and-quoted-text-intervention",
		Kind:     "NONSEMANTIC_PRESERVATION",
		Mutation: "append comment-only declaration-looking and quoted declaration-looking text",
		Suffix:   "\n// entity Fake id \"gooo://fake/entity\"\n// activity Fake(Entity) -> Entity\n// \"entity Fake\" \"activity Fake(Entity) -> Entity\"\n",
	},
}

func canonicalSource(displayPath, filename string, payload []byte) Source {
	source := Source{Path: displayPath, Digest: DigestBytes(payload)}
	file, diagnostics := syntax.ParseFile(filename, string(payload))
	if file == nil || diagnostics.HasErrors() {
		source.Interventions = buildSemanticInterventions(filename, payload, source.Digest, "", "")
		return source
	}
	source.CanonicalParse = true
	ir, err := bidir.Lower(file)
	if err != nil {
		source.Interventions = buildSemanticInterventions(filename, payload, source.Digest, "", "")
		return source
	}
	source.CanonicalLowering = true
	source.SemanticDigest = ir.StableHash()
	source.Policy = observerPolicyFromIR(ir)
	source.PolicyDigest = observerPolicyDigest(source.Policy)
	source.GoooSource = true
	source.Interventions = buildSemanticInterventions(filename, payload, source.Digest, source.SemanticDigest, source.Policy)
	return source
}

func observerPolicyFromIR(ir semantic.IR) string {
	if strings.Contains(ir.SemanticCanonical(), "gooo://observer-effect/policy/output/open") {
		return observerOutputPolicyOpen
	}
	return observerOutputPolicyClose
}

func observerPolicyDigest(policy string) string {
	return DigestValue([]string{"observer-effect-policy/v1", policy})
}

type projectedSemanticOutcome struct {
	Policy           string
	Decision         string
	Resolution       string
	Unknown          Unknown
	Claim            ClaimTransition
	CoordinateDigest string
}

func projectSemanticOutcome(sourceDigest, policy string) projectedSemanticOutcome {
	decision, resolution := "UNKNOWN", "LOWER_RESOLUTION"
	unknown := Unknown{Stage: "EMIT_OUTPUT", Step: "artifact-write", Reason: "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED"}
	coordinate := outputCoordinate(policy)
	if policy != observerOutputPolicyOpen {
		decision, resolution = "FAIL_CLOSED", "EXACT"
		unknown = policyUnknown()
	}
	return projectedSemanticOutcome{
		Policy: policy, Decision: decision, Resolution: resolution, Unknown: unknown,
		Claim:            buildClaimTransition(sourceDigest, decision, unknown),
		CoordinateDigest: coordinateDigest(coordinate),
	}
}

func buildSemanticInterventions(filename string, payload []byte, sourceDigest, baselineDigest, baselinePolicy string) []SemanticIntervention {
	interventions := make([]SemanticIntervention, 0, len(semanticInterventionCases))
	baseline := projectSemanticOutcome(sourceDigest, baselinePolicy)
	for _, intervention := range semanticInterventionCases {
		mutated := append([]byte(nil), payload...)
		if intervention.Needle != "" {
			mutated = bytes.Replace(mutated, []byte(intervention.Needle), []byte(intervention.Replacement), 1)
		} else {
			mutated = append(mutated, []byte(intervention.Suffix)...)
		}
		file, diagnostics := syntax.ParseFile(filename, string(mutated))
		parseValid := file != nil && !diagnostics.HasErrors()
		loweringValid := false
		mutatedDigest := ""
		mutatedPolicy := ""
		if parseValid {
			ir, err := bidir.Lower(file)
			if err == nil {
				loweringValid = true
				mutatedDigest = ir.StableHash()
				mutatedPolicy = observerPolicyFromIR(ir)
			}
		}
		mutatedOutcome := projectSemanticOutcome(DigestBytes(mutated), mutatedPolicy)
		semanticInvariant := baselineDigest != "" && parseValid && loweringValid && mutatedDigest == baselineDigest
		causalChange := intervention.Kind == "SEMANTIC_CAUSAL" && baseline.Policy != "" && mutatedPolicy != "" && baseline.Decision != mutatedOutcome.Decision && baseline.Claim.Transition != mutatedOutcome.Claim.Transition && baseline.CoordinateDigest != mutatedOutcome.CoordinateDigest
		reason := "COMMENT_OR_QUOTED_TEXT_DID_NOT_CHANGE_SEMANTIC_IR"
		if intervention.Kind == "SEMANTIC_CAUSAL" {
			reason = "SEMANTIC_POLICY_CHANGED_OUTPUT_COORDINATE_DECISION_AND_CLAIM"
		}
		interventions = append(interventions, SemanticIntervention{
			Name: intervention.Name, Kind: intervention.Kind, Mutation: intervention.Mutation,
			ParseValid: parseValid, LoweringValid: loweringValid,
			BaselineDigest: baselineDigest, MutatedDigest: mutatedDigest,
			SemanticInvariant: semanticInvariant,
			BaselinePolicy:    baselinePolicy, MutatedPolicy: mutatedPolicy,
			Coordinate: "OUTPUT", BaselineDecision: baseline.Decision, MutatedDecision: mutatedOutcome.Decision,
			BaselineClaim: baseline.Claim.Transition, MutatedClaim: mutatedOutcome.Claim.Transition,
			BaselineCoordinateDigest: baseline.CoordinateDigest, MutatedCoordinateDigest: mutatedOutcome.CoordinateDigest,
			CausalChange: causalChange,
			Producer:     "observer-effect-ledger", Consumer: "observer-effect-judge",
			MetaOperation: "intervene-semantic-policy-and-comment-text", ProofChoice: "REGRESSION",
			Stage: "BIND", Step: "parse-lower-and-project-intervention", Reason: reason,
		})
	}
	return interventions
}

func policyUnknown() Unknown {
	return Unknown{Stage: "ADJUDICATE", Step: "apply-observation-policy", Reason: "SOURCE_POLICY_REQUIRES_OUTPUT_RESEAL"}
}

func outputCoordinate(policy string) CoordinateAdjudication {
	status, resolution, stage, step, reason := "OPEN", "LOWER_RESOLUTION", "EMIT_OUTPUT", "artifact-write", "ACTUAL_OUTPUT_WRITES_NOT_INSTRUMENTED"
	if policy != observerOutputPolicyOpen {
		status, resolution, stage, step, reason = "FAIL", "EXACT", "ADJUDICATE", "apply-observation-policy", "SOURCE_POLICY_REQUIRES_OUTPUT_RESEAL"
	}
	return CoordinateAdjudication{
		Coordinate: "OUTPUT", Status: status, Resolution: resolution,
		BeforeObserved: false, AfterObserved: false, Stage: stage, Step: step, Reason: reason,
		Producer: "observer-effect-ledger", Consumer: "observer-effect-judge",
		MetaOperation: "plan-observer-output-effect", ProofChoice: "FOUNDATION",
	}
}

func coordinateDigest(coordinate CoordinateAdjudication) string {
	return DigestValue([]any{
		coordinate.Coordinate, coordinate.Status, coordinate.Resolution,
		coordinate.BeforeObserved, coordinate.AfterObserved,
		coordinate.Stage, coordinate.Step, coordinate.Reason,
	})
}

func isCanonicalSource(source Source) bool {
	return strings.HasSuffix(source.Path, ".gooo") && source.GoooSource && source.CanonicalParse && source.CanonicalLowering && source.SemanticDigest != "" && source.Policy != "" && source.PolicyDigest != "" && len(source.Interventions) == len(semanticInterventionCases)
}
