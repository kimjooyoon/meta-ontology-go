package recoveryorigin

import "testing"

func TestOriginCounterexamples(t *testing.T) {
	type mutate func(*R_RecoveryScope, *R_ReceiptOrigin, *R_ReceiptOrigin)
	cases := []struct {
		name   string
		change mutate
	}{
		{"unrelated prior plan", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { s.F_prior_plan_id = "urn:unrelated:plan" }},
		{"unrelated failed activity", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { s.F_failed_activity_id = "urn:unrelated:activity" }},
		{"missing receipt", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_read_error = "missing receipt" }},
		{"wrong publication run", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { p.F_run_id++ }},
		{"wrong readback run", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_run_id++ }},
		{"wrong attempt", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { p.F_attempt++ }},
		{"wrong job attempt", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_job_attempt++ }},
		{"wrong artifact attempt", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_artifact_name += "-2" }},
		{"stale source scope", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { s.F_source = "another-source" }},
		{"changed workflow", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { p.F_workflow_sha256 = "changed" }},
		{"wrong repository", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_repository = "other/repository" }},
		{"artifact from another run", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_artifact_run_id++ }},
		{"changed archive", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_archive_sha256 = "changed" }},
		{"publication did not complete", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { p.F_step_conclusions[4] = "failure" }},
		{"failure claim contradicted", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { p.F_step_conclusions[5] = "success" }},
		{"duplicate step", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) {
			p.F_step_numbers = append(p.F_step_numbers, 5)
			p.F_step_conclusions = append(p.F_step_conclusions, "success")
			p.F_step_references = append(p.F_step_references, p.F_step_references[4])
		}},
		{"same run is not independent readback", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) {
			r.F_run_id = p.F_run_id
			s.F_readback_run = p.F_run_id
			r.F_job_run_id = p.F_run_id
			r.F_artifact_run_id = p.F_run_id
		}},
		{"future receipt", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_completed_at = s.F_observed_at + 1 }},
		{"readback predates publication", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_completed_at = p.F_completed_at - 1 }},
		{"different observed subject", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_subjects[0] = "different" }},
		{"different image config", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) { r.F_image_ids[0] = "different" }},
		{"ambiguous images", func(s *R_RecoveryScope, p, r *R_ReceiptOrigin) {
			r.F_image_ids = append(r.F_image_ids, r.F_image_ids[0])
		}},
	}
	observed, passed := 0, 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			observed++
			root, scope := historyFixture(t)
			p, _ := Read(root, "publication", scope.F_publish_job)
			r, _ := Read(root, "readback", scope.F_readback_job)
			tc.change(&scope, &p, &r)
			got := Evaluate(scope, p, r)
			if got.F_state != "UNRESOLVED_ORIGIN" || got.F_binding_state != "REFRESH_ORIGIN" ||
				len(got.F_completed_effect_ids) != 0 || got.F_execution_authorized {
				t.Fatalf("unbound history became completed effect: %+v", got)
			}
			passed++
		})
	}
	t.Logf("independently specified origin counterexamples=%d/%d executed; full population=%d", passed, observed, len(cases))
}

func TestEffectIdentitySurvivesPresentationChanges(t *testing.T) {
	root, scope := historyFixture(t)
	p, _ := Read(root, "publication", scope.F_publish_job)
	r, _ := Read(root, "readback", scope.F_readback_job)
	for _, label := range []string{"publish again", "새로운 이름으로 재시도"} {
		scope.F_proposed_display_name, scope.F_proposed_operation = label, "publish-image"
		got := Evaluate(scope, p, r)
		if got.F_state != "REJECTED_REPLAY" || got.F_binding_state != "BOUND_HISTORICAL_OBSERVATION" || got.F_execution_authorized {
			t.Fatal("display name hid an already observed effect", got)
		}
	}
	scope.F_proposed_operation = "transport-original-build-proof"
	scope.F_proposed_subject = "different-subject"
	if got := Evaluate(scope, p, r); got.F_state != "INSUFFICIENT_EVIDENCE" || got.F_binding_state != "REVISE_SCOPE" {
		t.Fatal("proposal widened declared subject", got)
	}
	scope.F_proposed_subject, scope.F_proposed_operation = scope.F_subject, "arbitrary-operation"
	if got := Evaluate(scope, p, r); got.F_state != "INVALID_PROPOSAL" || got.F_binding_state != "REVISE_SCOPE" {
		t.Fatal("unknown effect acquired a canonical identity", got)
	}
}
