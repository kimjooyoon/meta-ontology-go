package foundationseed

import (
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorselection"
)

func observeSource(input predecessorresolution.Report, expectedHead string) Source {
	summary := input.Summary
	source := Source{
		CurrentHeadSHA: input.CurrentHeadSHA, ImmediatePredecessorSHA: input.ImmediatePredecessorSHA,
		ResolutionDigest: input.ReportDigest,
		ResolutionValid:  predecessorresolution.Validate(input) == nil,
		HeadBound:        input.CurrentHeadSHA == expectedHead,
		ObservedAttempts: summary.ObservedAttempts, MissingAttempts: summary.MissingAttempts,
		SearchLimit: summary.SearchLimit, SelectedAncestors: summary.SelectedAncestors,
		ValidCandidates: summary.ValidCandidates, AmbiguousCandidates: summary.AmbiguousCandidates,
		RepositoryWrites: summary.RepositoryWrites, ReadinessDeltaClaims: summary.ReadinessDeltaClaims,
		Contiguous: contiguous(input), AuthorityDenied: true,
	}
	source.SearchComplete = source.ObservedAttempts == predecessorresolution.SearchLimit &&
		source.SearchLimit == predecessorresolution.SearchLimit
	source.MissingComplete = source.MissingAttempts == predecessorresolution.SearchLimit
	source.ExactExhaustion = exactExhaustion(input, source)
	return source
}

func exactExhaustion(input predecessorresolution.Report, source Source) bool {
	summary := input.Summary
	return source.ResolutionValid && source.HeadBound && source.SearchComplete &&
		source.MissingComplete && source.Contiguous &&
		input.Decision == predecessorresolution.DecisionFailClosed &&
		input.Reason == predecessorresolution.ReasonExhausted &&
		input.BlockingSelectionReason == "" && input.Selected == nil &&
		summary.SelectedAncestors == 0 && summary.SelectedDepth == -1 &&
		summary.ValidCandidates == 0 && summary.AmbiguousCandidates == 0 &&
		summary.RepositoryWrites == 0 && summary.ReadinessDeltaClaims == 0 &&
		exactMissingAttempts(input)
}

func exactMissingAttempts(input predecessorresolution.Report) bool {
	if len(input.Attempts) != predecessorresolution.SearchLimit {
		return false
	}
	for index, attempt := range input.Attempts {
		expected, err := predecessorselection.Select(predecessorselection.Input{
			Repository: input.Repository, CurrentHeadSHA: input.CurrentHeadSHA,
			PredecessorSHA: attempt.AncestorSHA, Branch: canonicalBranch,
			Workflow: canonicalWorkflow,
		})
		if err != nil || !reflect.DeepEqual(attempt.Selection, expected.Report) {
			return false
		}
		if index+1 < len(input.Attempts) {
			if attempt.ParentSHA != input.Attempts[index+1].AncestorSHA {
				return false
			}
		} else if attempt.ParentSHA != "" {
			return false
		}
	}
	return true
}

func contiguous(input predecessorresolution.Report) bool {
	for index := 1; index < len(input.Attempts); index++ {
		if input.Attempts[index-1].ParentSHA != input.Attempts[index].AncestorSHA {
			return false
		}
	}
	return len(input.Attempts) > 0
}
