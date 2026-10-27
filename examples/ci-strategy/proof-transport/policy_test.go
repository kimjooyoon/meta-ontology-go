package prooftransport

import "testing"

func completeObservation(scope R_TransportScope) R_TransportObservation {
	return R_TransportObservation{F_base_url: scope.F_base_url, F_material_digest: scope.F_manifest_digest,
		F_observed_at: 100, F_now: 101, F_subject_status: 200, F_subject_digest: scope.F_subject_digest,
		F_manifest_status: 200, F_manifest_digest: scope.F_manifest_digest, F_payload_status: 200,
		F_payload_digest: scope.F_payload_digest, F_config_status: 200, F_config_digest: scope.F_config_digest,
		F_referrers_status: 200, F_referrers_url: scope.F_referrers_url, F_referrers_digest: "sha256:retained-index",
		F_referrers_media_type: "application/vnd.oci.image.index.v1+json", F_descriptors: []string{scope.F_manifest_digest}}
}

func TestSourceOwnedTransportChoices(t *testing.T) {
	scope := Plan("transport-original-build-proof")
	for _, row := range []struct {
		name, state string
		write       bool
		change      func(*R_TransportObservation)
	}{
		{"present", "ALREADY_PRESENT", false, func(*R_TransportObservation) {}},
		{"present-on-partial-page", "ALREADY_PRESENT", false, func(o *R_TransportObservation) { o.F_next_link = "next page" }},
		{"known-absent", "READY_TO_RESTORE", true, func(o *R_TransportObservation) { o.F_manifest_status = 404; o.F_descriptors = nil }},
		{"partial-absence", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) {
			o.F_manifest_status = 404
			o.F_descriptors = nil
			o.F_next_link = "next page"
		}},
		{"contradictory-index", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_manifest_status = 404 }},
		{"missing-discovery", "REVISE_DISCOVERY", false, func(o *R_TransportObservation) { o.F_descriptors = nil }},
		{"legacy-present", "ALREADY_PRESENT", false, func(o *R_TransportObservation) { o.F_referrers_url = scope.F_fallback_url }},
		{"legacy-absence", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) {
			o.F_referrers_url = scope.F_fallback_url
			o.F_manifest_status = 404
			o.F_descriptors = nil
		}},
		{"stale", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_now = 161 }},
		{"future", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_now = 99 }},
		{"read-error", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_read_error = "timeout" }},
		{"unavailable-index", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_referrers_status = 401 }},
		{"wrong-index-type", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_referrers_media_type = "text/html" }},
		{"wrong-index-origin", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_referrers_url = "https://another.invalid/index" }},
		{"wrong-material", "REFUTED_MATERIAL", false, func(o *R_TransportObservation) { o.F_material_digest = "changed" }},
		{"wrong-repository", "REFUTED_MATERIAL", false, func(o *R_TransportObservation) { o.F_base_url = "https://another.invalid/v2/repo" }},
		{"wrong-subject", "UNRESOLVED_MATERIAL", false, func(o *R_TransportObservation) { o.F_subject_digest = "changed" }},
		{"wrong-payload", "UNRESOLVED_MATERIAL", false, func(o *R_TransportObservation) { o.F_payload_digest = "changed" }},
		{"missing-config", "UNRESOLVED_MATERIAL", false, func(o *R_TransportObservation) { o.F_config_status = 404 }},
		{"changed-manifest", "REFRESH_OBSERVATION", false, func(o *R_TransportObservation) { o.F_manifest_digest = "changed" }},
	} {
		t.Run(row.name, func(t *testing.T) {
			o := completeObservation(scope)
			row.change(&o)
			got := Evaluate("transport-original-build-proof", o)
			if got.F_state != row.state || got.F_write_manifest != row.write || got.F_effect_id != scope.F_effect_id {
				t.Fatal(got)
			}
		})
	}
	if got := Evaluate("publish-image", completeObservation(scope)); got.F_state != "INVALID_PROPOSAL" || got.F_write_manifest {
		t.Fatal(got)
	}
}
