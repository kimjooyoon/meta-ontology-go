package sourceauthoritypromotion

import (
	"encoding/json"
	"testing"
)

func TestExternalEvidenceUsesSnakeCaseFields(t *testing.T) {
	assuranceJSON := []byte(`{"schema":"assurance","subject_sha":"sha","denominator_id":"denominator","denominator_digest":"digest","assurance_decision":"PARTIAL","candidate_decision":"ALLOW_LIMITED","denominator":[{"metric_id":"metric","priority":"P1","class":"DRIVER","proof_choice":"FOUNDATION","required_meta_operation":"bind"}],"obligations":[{"metric_id":"metric","status":"NOT_IMPLEMENTED","resolution":"NONE","meta_operation":"bind"}],"summary":{"denominator_total":12,"operating":6,"not_implemented":6,"implementation_coverage_bps":5000,"unknown_top_decisions":0,"unresolved_indicators":0,"violated_guardrails":0,"repository_writes":0}}`)
	var assurance assuranceDocument
	if err := json.Unmarshal(assuranceJSON, &assurance); err != nil {
		t.Fatal(err)
	}
	definition, obligation := assurance.Denominator[0], assurance.Obligations[0]
	if assurance.SubjectSHA != "sha" || assurance.DenominatorID != "denominator" || assurance.Summary.ImplementationCoverageBPS != 5000 || definition.RequiredMetaOperation != "bind" || obligation.MetaOperation != "bind" {
		t.Fatalf("snake_case assurance fields were not bound: %#v", assurance)
	}

	upstreamJSON := []byte(`{"schema":"upstream","subject_sha":"sha","decision":"PASS","resolution":"EXACT","denominator_id":"denominator","denominator_digest":"digest","repository_writes":0,"promotion_credit_bps":0,"summary":{"cases_total":3,"cases_passed":3,"exact_allow":1,"fail_closed":2,"coverage_bps":10000},"cases":[{"id":"exact","expected_observation":"SATISFIED","expected_resolution":"EXACT","expected_enforcement":"ALLOW","expected_reason":"SOURCE_SNAPSHOT_EXACT","passed":true,"receipt":{"subject_sha":"sha","observation":"SATISFIED","resolution":"EXACT","enforcement":"ALLOW","reason":"SOURCE_SNAPSHOT_EXACT","repository_writes":0,"promotion_credit_bps":0,"snapshot":{"digest":"snapshot","source_ref":"source","authority_ref":"authority","bytes":77,"authority":{"repository":"repo","revision":"revision","path":"README.md"},"selection":{"start_line":1,"end_line":1}},"indicators":[{"class":"OUTCOME","proof_choice":"FOUNDATION","satisfied":true}]}}]}`)
	var upstream upstreamDocument
	if err := json.Unmarshal(upstreamJSON, &upstream); err != nil {
		t.Fatal(err)
	}
	receipt := upstream.Cases[0].Receipt
	if upstream.SubjectSHA != "sha" || upstream.PromotionCreditBPS != 0 || upstream.Summary.CoverageBPS != 10000 || receipt.Snapshot.Authority.Revision != "revision" || receipt.Snapshot.Selection.EndLine != 1 || receipt.Indicators[0].ProofChoice != "FOUNDATION" {
		t.Fatalf("snake_case upstream fields were not bound: %#v", upstream)
	}
}
