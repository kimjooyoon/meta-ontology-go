package toolchainrelease

type packageProfileFill struct {
	InputSHA string `json:"input_source_sha256"`
	Activity struct {
		Package  string `json:"package_path"`
		Activity string
		Lowered  string `json:"lowered_name"`
	}
	Generation struct {
		Source     string
		GoooSource string `json:"gooo_source"`
		Report     struct {
			Activity, Decision string
			ID                 string                  `json:"activity_id"`
			Generated          string                  `json:"generated_digest"`
			Replay             string                  `json:"replay_digest"`
			Typechecked        bool                    `json:"typecheck_passed"`
			Replayed           bool                    `json:"deterministic_replay"`
			Fill               *packageProfileFillData `json:"body_fill"`
		}
	}
}

type packageProfileFillData struct {
	Decision *struct {
		Schema, Mode, Selected, Provider string
		Fallback                         string `json:"fallback_reason"`
		Request                          string `json:"request_sha256"`
	}
	ExternalKnown *bool `json:"external_provider_calls_known"`
	Schema        string
	Original      string               `json:"original_source_digest"`
	Plan          string               `json:"ir_plan_sha256"`
	Selected      string               `json:"selected_candidate_id"`
	Holes         []jointSmokeFillHole `json:"hole_fills"`
	Passed        *int                 `json:"test_cases_passed"`
	Total         *int                 `json:"test_cases_total"`
	HoldoutPassed *int                 `json:"holdout_cases_passed"`
	HoldoutTotal  *int                 `json:"holdout_cases_total"`
	LocalCalls    *int                 `json:"local_model_predictions"`
	ExternalCalls *int                 `json:"external_provider_calls"`
}
