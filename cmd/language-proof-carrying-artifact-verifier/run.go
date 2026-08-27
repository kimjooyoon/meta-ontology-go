package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	verifier "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageproofartifactverifier"
)

type options struct {
	head, contract, valid, tampered, coherentTampered, missing, byteOnly, wrongRecipe string
	source, operation, recipe, independence, writeSet, coherentOperation, output, check string
	semanticArtifact, semanticSource, semanticOperation, commentArtifact, commentSource, commentOperation string
}

func run(args []string) int {
	flags := flag.NewFlagSet("language-proof-carrying-artifact-verifier", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "proof-carrying contract")
	flags.StringVar(&value.valid, "valid", "", "valid carried artifact")
	flags.StringVar(&value.tampered, "tampered", "", "tampered carried artifact")
	flags.StringVar(&value.coherentTampered, "coherent-tamper", "", "coherently resealed tampered artifact")
	flags.StringVar(&value.missing, "missing", "", "artifact with missing evidence")
	flags.StringVar(&value.byteOnly, "byte-only", "", "artifact evaluated without external evidence")
	flags.StringVar(&value.wrongRecipe, "wrong-recipe", "", "independent recipe mutation")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.operation, "operation", "", "operation receipt")
	flags.StringVar(&value.recipe, "recipe", "", "canonical recipe")
	flags.StringVar(&value.independence, "independence", "", "verifier dependency evidence")
	flags.StringVar(&value.writeSet, "write-set", "", "repository write-set observation")
	flags.StringVar(&value.coherentOperation, "coherent-operation", "", "coherently resealed tampered operation receipt")
	flags.StringVar(&value.semanticArtifact, "semantic-artifact", "", "semantic intervention artifact")
	flags.StringVar(&value.semanticSource, "semantic-source", "", "semantic intervention source")
	flags.StringVar(&value.semanticOperation, "semantic-operation", "", "semantic intervention operation receipt")
	flags.StringVar(&value.commentArtifact, "comment-artifact", "", "comment-only intervention artifact")
	flags.StringVar(&value.commentSource, "comment-source", "", "comment-only intervention source")
	flags.StringVar(&value.commentOperation, "comment-operation", "", "comment-only intervention operation receipt")
	flags.StringVar(&value.output, "output", "", "verification report output")
	flags.StringVar(&value.check, "check", "", "existing report to validate")
	if flags.Parse(args) != nil {
		return 2
	}
	if value.check != "" {
		report, err := verifier.LoadReport(value.check)
		if err != nil || verifier.Validate(report) != nil {
			return 1
		}
		return 0
	}
	if value.head == "" || value.contract == "" || value.valid == "" || value.tampered == "" ||
		value.missing == "" || value.byteOnly == "" || value.wrongRecipe == "" || value.source == "" ||
		value.operation == "" || value.recipe == "" || value.independence == "" || value.output == "" {
		return 2
	}
	read := func(path string) ([]byte, bool) {
		raw, err := os.ReadFile(path)
		return raw, err == nil
	}
	contractRaw, ok := read(value.contract)
	if !ok {
		return 2
	}
	contract, err := verifier.DecodeContract(contractRaw)
	if err != nil {
		return 2
	}
	valid, validOK := read(value.valid)
	tampered, tamperedOK := read(value.tampered)
	coherentTampered, coherentTamperedOK := read(value.coherentTampered)
	missing, missingOK := read(value.missing)
	byteOnly, byteOnlyOK := read(value.byteOnly)
	wrongRecipe, wrongRecipeOK := read(value.wrongRecipe)
	source, sourceOK := read(value.source)
	operation, operationOK := read(value.operation)
	recipe, recipeOK := read(value.recipe)
	independenceRaw, independenceOK := read(value.independence)
	writeSetRaw, writeSetOK := read(value.writeSet)
	coherentOperation, coherentOperationOK := read(value.coherentOperation)
	semanticArtifact, semanticArtifactOK := read(value.semanticArtifact)
	semanticSource, semanticSourceOK := read(value.semanticSource)
	semanticOperation, semanticOperationOK := read(value.semanticOperation)
	commentArtifact, commentArtifactOK := read(value.commentArtifact)
	commentSource, commentSourceOK := read(value.commentSource)
	commentOperation, commentOperationOK := read(value.commentOperation)
	if !validOK || !tamperedOK || !coherentTamperedOK || !missingOK || !byteOnlyOK || !wrongRecipeOK || !sourceOK || !operationOK || !recipeOK || !independenceOK ||
		!writeSetOK || !coherentOperationOK || !semanticArtifactOK || !semanticSourceOK || !semanticOperationOK ||
		!commentArtifactOK || !commentSourceOK || !commentOperationOK {
		return 2
	}
	var independence verifier.IndependenceEvidence
	if err := json.Unmarshal(independenceRaw, &independence); err != nil {
		return 2
	}
	writeSet, err := verifier.DecodeWriteSet(writeSetRaw)
	if err != nil {
		return 2
	}
	interventions := []verifier.InterventionInput{
		{ID: "semantic-source-intervention", Kind: "SEMANTIC", Before: verifier.SubjectInput{Artifact: valid, Source: source, Operation: operation, Recipe: recipe}, After: verifier.SubjectInput{Artifact: semanticArtifact, Source: semanticSource, Operation: semanticOperation, Recipe: recipe}},
		{ID: "comment-only-intervention", Kind: "NONSEMANTIC", Before: verifier.SubjectInput{Artifact: valid, Source: source, Operation: operation, Recipe: recipe}, After: verifier.SubjectInput{Artifact: commentArtifact, Source: commentSource, Operation: commentOperation, Recipe: recipe}},
	}
	report := verifier.Evaluate(verifier.Input{Contract: contract, HeadSHA: value.head, ValidArtifact: valid,
		TamperedArtifact: tampered, CoherentTamperedArtifact: coherentTampered, MissingArtifact: missing, ByteOnlyArtifact: byteOnly, WrongRecipe: wrongRecipe,
		Source: source, Operation: operation, Recipe: recipe, Independence: independence, WriteSet: writeSet,
		CoherentOperation: coherentOperation, Interventions: interventions})
	if err := verifier.WriteReport(value.output, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("proof-carrying verifier: %s %d/%d authority=%s transitions=%d\n", report.ConformanceDecision,
		report.Summary.CasesSatisfied, report.Summary.CasesTotal, report.ArtifactUseAuthority, len(report.Transitions))
	if report.ConformanceDecision != "PASS" {
		return 1
	}
	return 0
}
