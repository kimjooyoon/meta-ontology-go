package sourcesplit

import (
	"bytes"
	"errors"
	"os"
	"slices"

	recipeauthority "github.com/kimjooyoon/meta-ontology-go/internal/meta/functionextractorrecipe"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractionplan"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func observeExtraction(root, head, metricsDigest string, action generation.Action,
	recipes []recipeauthority.ExtractionRecipe, recipeErr error) generation.CandidatePrerequisite {
	observed := generation.NewCandidatePrerequisite(action, head, "", metricsDigest, "EXTRACTION_PLAN_READY")
	observed.Stage, observed.Step = "prepare-extraction", "stage-without-writing"
	if recipeErr != nil {
		return extractionPrerequisiteFailure(observed, recipeErr)
	}
	subject, err := sourcepolicy.ParseSourceSubject(action.Subject)
	if err != nil {
		return extractionPrerequisiteFailure(observed, err)
	}
	target, err := secureSourcePath(root, subject.Path)
	if err != nil {
		return extractionPrerequisiteFailure(observed, err)
	}
	before, err := os.ReadFile(target)
	if err != nil {
		return extractionPrerequisiteFailure(observed, err)
	}
	observed.SourceDigest = sourceDigest(before)
	prepared, prepareErr := extractionplan.Prepare(root, []string{subject.Path}, recipes)
	after, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(before, after) {
		return extractionPrerequisiteFailure(observed, projectionextractor.Failure{
			Stage: "prepare-extraction", Step: "reread-source", Reason: "SOURCE_CHANGED_DURING_PLANNING",
			UnknownClass: "STALE", NextOperation: "observe-source-prerequisites"})
	}
	return adjudicateExtractionPreparation(observed, subject.Path, prepared, prepareErr)
}

func adjudicateExtractionPreparation(observed generation.CandidatePrerequisite, logical string,
	prepared extractionplan.Result, err error) generation.CandidatePrerequisite {
	if err == nil {
		err = prepared.Failures[logical]
	}
	if err != nil {
		return extractionPrerequisiteFailure(observed, err)
	}
	operations := prepared.Operations[logical]
	if recipe, exists := prepared.Recipes[logical]; exists {
		operations = []string{recipe.Operation}
	}
	if !slices.Contains(operations, string(sourcepolicy.OperationExtractFunction)) || len(prepared.Files[logical].Data) == 0 {
		return extractionPrerequisiteFailure(observed, projectionextractor.Failure{
			Stage: "prepare-extraction", Step: "bind-operation", Reason: "EXTRACTION_NOT_REQUIRED",
			UnknownClass: "KNOWN_CONTRADICTION", NextOperation: "reconcile-source-metrics"})
	}
	return observed
}

func extractionPrerequisiteFailure(observed generation.CandidatePrerequisite, err error) generation.CandidatePrerequisite {
	observed.State, observed.Reason = "UNKNOWN", "EXTRACTION_PLANNING_UNAVAILABLE"
	observed.UnknownClass, observed.NextOperation = "DIRECT_MISSING", "observe-source-prerequisites"
	observed.Diagnostics = []string{err.Error()}
	if native, ok := errors.AsType[projectionextractor.Failure](err); ok {
		observed.Stage, observed.Step, observed.Reason = native.Stage, native.Step, native.Reason
		observed.UnknownClass, observed.NextOperation = native.UnknownClass, native.NextOperation
		observed.BlockedBy = append([]string(nil), native.BlockedBy...)
		observed.Diagnostics = append([]string(nil), native.Diagnostics...)
		if native.UnknownClass == "KNOWN_CONTRADICTION" {
			observed.State = "BLOCKED"
		}
	}
	return observed
}
