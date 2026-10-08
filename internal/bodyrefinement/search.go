package bodyrefinement

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// SearchObservation extends the opt-in policy input. The next alternative is
// the first source-declared grammar/bound pair not visited in this run.
type SearchObservation struct {
	Attempted        int    `json:"search_attempted"`
	Retained         int    `json:"search_retained"`
	Omitted          int    `json:"search_omitted"`
	Grammar          string `json:"grammar"`
	CandidateLimit   int    `json:"candidate_limit"`
	NextSearchID     string `json:"next_search_id"`
	NextGrammar      string `json:"next_grammar"`
	NextAttemptLimit int    `json:"next_attempt_limit"`
}

func observeSearch(ctx context.Context, filename string, source []byte, options Options, prior []Round, round *Round) error {
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, options.Activity)
	if err != nil {
		return err
	}
	if !bodycodegen.IsSourceIRSearch(spec) {
		return fmt.Errorf("search policy requires a source-owned integer search")
	}
	search := &SearchObservation{Grammar: spec.Search.Grammar, CandidateLimit: spec.Search.MaxCandidates}
	seen := map[string]bool{searchKey(search): true}
	for _, old := range prior {
		if old.Search != nil {
			seen[searchKey(old.Search)] = true
		}
	}
	for _, alternative := range spec.SearchAlternatives {
		key := fmt.Sprintf("%s/%d", alternative.Grammar, alternative.MaxCandidates)
		if !seen[key] {
			search.NextSearchID, search.NextGrammar = alternative.ID, alternative.Grammar
			search.NextAttemptLimit = min(options.MaxAttempts, alternative.MaxCandidates)
			break
		}
	}
	steps := append([]bodyexecution.CompositionStep(nil), round.Composition.Preparations...)
	steps = append(steps, round.Composition.Steps...)
	for _, step := range steps {
		if step.Generation.Report.Activity != options.Activity || step.Generation.Report.BodySearch == nil {
			continue
		}
		r := step.Generation.Report.BodySearch
		if r.CandidateGeneration == nil {
			return fmt.Errorf("search policy is missing generated-space observations")
		}
		search.Attempted, search.Retained, search.Omitted = r.AttemptedCandidates, r.CandidateCount, r.CandidateGeneration.CandidatesOmitted
		round.Search, round.Observation.Limit = search, min(options.MaxAttempts, spec.Search.MaxCandidates)
		return nil
	}
	return fmt.Errorf("search policy target has no search construction observation")
}

func searchKey(search *SearchObservation) string {
	return fmt.Sprintf("%s/%d", search.Grammar, search.CandidateLimit)
}

func validSearchDecision(options Options, round *Round) bool {
	s, d := round.Search, round.Decision
	return options.SearchPolicy && s != nil && s.NextSearchID != "" && d.NextSearchID == s.NextSearchID &&
		d.NextAttempts >= 1 && d.NextAttempts <= s.NextAttemptLimit
}
