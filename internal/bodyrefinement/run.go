package bodyrefinement

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func Run(ctx context.Context, filename string, source []byte, feedback bodyexecution.CompositionCases, options Options) (Result, error) {
	result := Result{Schema: "gooo/body-refinement/v1", Status: "UNKNOWN", FeedbackStatus: "UNKNOWN", EvaluationStatus: "UNKNOWN", Activity: options.Activity,
		OriginalSource: string(source), PolicySource: string(options.PolicySource), SearchPolicy: options.SearchPolicy, FeedbackCases: feedback,
		EvaluationCases: options.Evaluation, SelectedRound: -1, Rounds: []Round{},
		Scope: "Gooo policy controls retained round, bounded source-budget revisions and promotion of explicit root-input counterexamples into source cases; feedback is adaptive, not held-out evidence; final evaluation is never sent to the policy; model-training exposure is unknown"}
	if options.SearchPolicy {
		result.Scope += "; search policy can advance through source-declared grammar/candidate-bound alternatives without revisiting a grammar/bound pair"
	}
	fail := func(err error) (Result, error) {
		result.Status, result.Failure = "FAIL_CLOSED", err.Error()
		return result, err
	}
	attempts, err := validate(ctx, filename, source, feedback, &options)
	if err != nil {
		return fail(err)
	}
	result.PolicyComposition, err = preparePolicy(ctx, options)
	if err != nil {
		return fail(err)
	}
	if err := runRounds(ctx, filename, source, feedback, options, attempts, &result); err != nil {
		return fail(err)
	}
	if err := finishRefinement(ctx, filename, options, &result); err != nil {
		return fail(err)
	}
	return result, nil
}

func runRounds(ctx context.Context, filename string, source []byte, feedback bodyexecution.CompositionCases,
	options Options, attempts int, result *Result) error {
	current, best := append([]byte(nil), source...), 0
	for index := 0; index < options.MaxRounds; index++ {
		round := Round{Source: string(current), Observation: Observation{Best: best, Attempts: attempts,
			Limit: options.MaxAttempts, Round: index + 1, RoundLimit: options.MaxRounds}}
		err := observeRound(ctx, filename, feedback, options, &round)
		if err == nil && options.SearchPolicy {
			err = observeSearch(ctx, filename, current, options, result.Rounds, &round)
		}
		if err == nil {
			err = decide(ctx, options, result.PolicyComposition, &round)
		}
		result.Rounds = append(result.Rounds, round)
		if err != nil {
			return err
		}
		if round.Decision.Retain {
			result.SelectedRound, best = index, round.Runtime.FinitePassed
		}
		if round.Decision.Action == "STOP" {
			result.StopReason = "POLICY_STOP"
			break
		}
		if index+1 == options.MaxRounds {
			result.StopReason = "ROUND_LIMIT"
			break
		}
		current, err = reviseRound(ctx, filename, current, options, &result.Rounds[index])
		attempts = round.Decision.NextAttempts
		if err != nil {
			return err
		}
	}
	return nil
}

func finishRefinement(ctx context.Context, filename string, options Options, result *Result) error {
	if result.SelectedRound < 0 {
		return fmt.Errorf("Gooo feedback policy retained no program")
	}
	selected := result.Rounds[result.SelectedRound]
	result.Status = "PROGRESS"
	if selected.Runtime.FiniteTotal > 0 && selected.Runtime.FinitePassed == selected.Runtime.FiniteTotal {
		result.Status = "PASS"
	}
	result.FeedbackStatus = result.Status
	if options.Evaluation != nil {
		evaluation, err := bodyexecution.ExecuteComposition(ctx, filename, []byte(selected.Source), selected.Composition, *options.Evaluation, options.GoBinary)
		result.Evaluation = &evaluation
		if err != nil {
			return err
		}
		result.EvaluationStatus = "PROGRESS"
		if evaluation.FiniteTotal > 0 && evaluation.FinitePassed == evaluation.FiniteTotal {
			result.EvaluationStatus = "PASS"
		} else {
			result.Status = "PROGRESS"
		}
	}
	return nil
}

func validate(ctx context.Context, filename string, source []byte, feedback bodyexecution.CompositionCases, options *Options) (int, error) {
	if options.MaxRounds < 1 || options.MaxRounds > 8 || options.MaxAttempts < 1 || options.MaxAttempts > 64 || feedback.Schema != "gooo/body-composition-cases/v1" {
		return 0, fmt.Errorf("refinement requires finite feedback cases, 1..8 rounds and 1..64 maximum attempts")
	}
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, options.Activity)
	if err != nil {
		return 0, err
	}
	if spec == nil || spec.FillPlan != nil || spec.MaxAttempts > options.MaxAttempts {
		return 0, fmt.Errorf("refinement activity requires a choice/search assembly within the requested attempt limit")
	}
	if options.SearchPolicy && spec.Search == nil {
		return 0, fmt.Errorf("search policy requires a source-owned integer search")
	}
	if spec.Search != nil && !options.SearchPolicy && options.MaxAttempts > spec.Search.MaxCandidates {
		options.MaxAttempts = spec.Search.MaxCandidates
	}
	if err := bodyexecution.ValidateCompositionCases(ctx, filename, source, feedback); err != nil {
		return 0, err
	}
	if options.Evaluation != nil {
		if options.Evaluation.Schema != "gooo/body-composition-cases/v1" {
			return 0, fmt.Errorf("final evaluation requires finite cases")
		}
		if err := bodyexecution.ValidateCompositionCases(ctx, filename, source, *options.Evaluation); err != nil {
			return 0, err
		}
	}
	return spec.MaxAttempts, nil
}

func observeRound(ctx context.Context, filename string, feedback bodyexecution.CompositionCases, options Options, round *Round) error {
	var err error
	source := []byte(round.Source)
	round.Composition, err = bodyexecution.GenerateComposition(ctx, filename, source, feedback, options.ModelPath)
	if err != nil {
		return err
	}
	round.Runtime, err = bodyexecution.ExecuteComposition(ctx, filename, source, round.Composition, feedback, options.GoBinary)
	round.Observation.Matched, round.Observation.Total = round.Runtime.FinitePassed, round.Runtime.FiniteTotal
	if err == nil {
		_, round.Observation.Counterexamples, err = bodycodegen.ExtendAssemblyCases(ctx, filename, source, options.Activity, counterexamples(*round, options.Activity))
	}
	return err
}
