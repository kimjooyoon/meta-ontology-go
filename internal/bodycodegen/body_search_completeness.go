package bodycodegen

func bodySearchCompletenessDimensions(search *IRBodySearchReceipt, generatedDigest string) ([]CompletenessDimension, []string) {
	training := completenessDimension("search_training_accuracy", search.TrainingPassed, search.TrainingTotal,
		"training cases matched by the finally selected body in the bounded evaluator",
		"Training cases drive candidate search; their score does not establish unseen behavior or global candidate optimality.",
		[]string{"body_search.training_suite_sha256:" + search.TrainingSuiteSHA256,
			"body_search.evaluator:" + search.Evaluator, "generated_digest:" + generatedDigest}, false)
	if search.TrainingAccuracyPercent == nil {
		training.Status = "UNKNOWN"
		training.Reason = "The final emitted body has no completed training score."
	} else if search.TrainingTotal > 0 && search.TrainingPassed == 0 {
		training.Status = "PROGRESS"
	}
	holdout := completenessDimension("search_holdout_accuracy", search.HoldoutPassed, search.HoldoutTotal,
		"post-selection holdout cases matched by the selected body in the bounded evaluator",
		"Holdout inputs are disjoint from training, excluded from choice requests, and never used to change the selected candidate.",
		[]string{"body_search.holdout_suite_sha256:" + search.HoldoutSuiteSHA256,
			"body_search.evaluator:" + search.Evaluator, "generated_digest:" + generatedDigest}, false)
	if search.HoldoutAccuracyPercent == nil {
		holdout.Status = "UNKNOWN"
		if search.HoldoutError != "" {
			holdout.Reason = "Holdout could not be scored by the bounded evaluator: " + search.HoldoutError
		}
	} else if search.HoldoutTotal > 0 && search.HoldoutPassed == 0 {
		holdout.Status = "PROGRESS"
	}
	coverage := completenessDimension("search_candidate_observation", search.AttemptedCandidates, search.CandidateCount,
		"candidate bodies attempted within the declared search budget",
		"A tested subset cannot establish the best score available among all declared candidates; rejected bodies remain in the attempt trace.",
		[]string{"body_search.attempts", "body_search.ir_plan_sha256:" + search.IRPlanSHA256}, false)
	scoring := completenessDimension("search_candidate_scoring", search.EvaluatedCandidates, search.CandidateCount,
		"candidate bodies with completed finite-suite scoring",
		"Parse, typecheck, and evaluator rejections are attempted but unscored; their accuracy is unavailable.",
		[]string{"body_search.attempts.scoring_completed", "body_search.ir_plan_sha256:" + search.IRPlanSHA256}, false)
	if search.AttemptedCandidates > 0 && search.EvaluatedCandidates == 0 {
		scoring.Status = "PROGRESS"
	}
	core := []string{training.ID}
	dimensions := []CompletenessDimension{training, holdout, coverage, scoring}
	if search.CandidateGeneration != nil {
		generated := search.CandidateGeneration
		grammar := completenessDimension("search_candidate_grammar_coverage", generated.CandidatesRetained,
			generated.CandidatesEnumerated,
			"retained candidates divided by unique expressions enumerated in the named bounded grammar",
			"Full grammar coverage only closes enumeration for this candidate grammar; it does not establish intent or whole-domain correctness.",
			[]string{"body_search.candidate_generation.grammar:" + generated.Grammar,
				"body_search.candidate_generation.candidate_set_sha256:" + generated.CandidateSetSHA256}, false)
		if !generated.GrammarComplete {
			grammar.Status = "PROGRESS"
		}
		dimensions = append(dimensions, grammar)
		core = append(core, grammar.ID)
	}
	if search.HoldoutTotal > 0 {
		core = append(core, holdout.ID)
	}
	return dimensions, core
}

func bodySearchProviderSummary(search *IRBodySearchReceipt) (string, string) {
	mode, provider := "deterministic_fallback", "deterministic"
	for _, attempt := range search.Attempts {
		if attempt.Decision != nil && attempt.Decision.Mode == "laya" {
			mode, provider = "laya", "laya"
		}
	}
	return mode, provider
}
