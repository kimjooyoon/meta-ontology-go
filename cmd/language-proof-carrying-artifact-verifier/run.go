package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	verifier "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageproofartifactverifier"
)

func readBundle(path string) verifier.Bundle {
	if path == "" {
		return verifier.Bundle{}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return verifier.Bundle{}
	}
	bundle, err := verifier.DecodeBundle(raw)
	if err != nil {
		return verifier.Bundle{}
	}
	return bundle
}

type options struct {
	head, contract, valid, tampered, coherentTampered, missing, byteOnly, wrongRecipe                        string
	source, operation, recipe, independence, writeSet, coherentOperation, output, check                      string
	semanticArtifact, semanticSource, semanticOperation, commentArtifact, commentSource, commentOperation    string
	recipeOnly, missingAttachment, wrongAttachmentDigest, unrelatedTampered, staleHead, unauthorizedConsumer string
	claimProposition, claimDependency, claimProofChoice, claimTarget, unauthorizedBundle                     string
	bundle, packBundle, bundleInputs, checkout                                                               string
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
	flags.StringVar(&value.recipeOnly, "recipe-only", "", "artifact for recipe-only mutation")
	flags.StringVar(&value.missingAttachment, "missing-attachment", "", "artifact tested without operation attachment")
	flags.StringVar(&value.wrongAttachmentDigest, "wrong-attachment-digest", "", "operation attachment with stale digest")
	flags.StringVar(&value.unrelatedTampered, "unrelated-tamper", "", "unrelated evidence tamper")
	flags.StringVar(&value.staleHead, "stale-head", "", "artifact bound to a stale head")
	flags.StringVar(&value.unauthorizedConsumer, "unauthorized-consumer", "", "raw consumer without attestation")
	flags.StringVar(&value.claimProposition, "claim-proposition-tamper", "", "coherently resealed proposition tamper")
	flags.StringVar(&value.claimDependency, "claim-dependency-tamper", "", "coherently resealed dependency tamper")
	flags.StringVar(&value.claimProofChoice, "claim-proof-choice-tamper", "", "coherently resealed proof-choice tamper")
	flags.StringVar(&value.claimTarget, "claim-target-tamper", "", "coherently resealed target tamper")
	flags.StringVar(&value.unauthorizedBundle, "unauthorized-bundle", "", "bundle supplied to unauthorized consumer")
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
	flags.StringVar(&value.bundle, "bundle", "", "portable proof bundle input")
	flags.StringVar(&value.packBundle, "pack-bundle", "", "portable proof bundle output")
	flags.StringVar(&value.bundleInputs, "bundle-inputs", "", "bundle input manifest")
	flags.StringVar(&value.checkout, "checkout", "", "checkout evidence")
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
	if value.packBundle != "" {
		if value.head == "" || value.bundleInputs == "" || value.checkout == "" {
			return 2
		}
		inputsRaw, err := os.ReadFile(value.bundleInputs)
		if err != nil {
			return 2
		}
		var inputs []verifier.BundleInput
		if err := json.Unmarshal(inputsRaw, &inputs); err != nil {
			return 2
		}
		checkoutRaw, err := os.ReadFile(value.checkout)
		if err != nil {
			return 2
		}
		var checkout verifier.CheckoutEvidence
		if err := json.Unmarshal(checkoutRaw, &checkout); err != nil {
			return 2
		}
		bundle, err := verifier.PackBundle(value.head, checkout, inputs)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		raw, err := json.MarshalIndent(bundle, "", "  ")
		if err != nil || os.WriteFile(value.packBundle, append(raw, '\n'), 0o644) != nil {
			return 1
		}
		fmt.Printf("proof bundle: %s files=%d digest=%s\n", bundle.Digest, len(bundle.Files), bundle.Digest)
		return 0
	}
	if value.bundle != "" {
		if value.output == "" {
			return 2
		}
		raw, err := os.ReadFile(value.bundle)
		if err != nil {
			return 2
		}
		bundle, err := verifier.DecodeBundle(raw)
		if err != nil {
			return 1
		}
		input, err := verifier.InputFromBundle(bundle)
		if err != nil {
			return 1
		}
		report := verifier.Evaluate(input)
		if err := verifier.WriteReport(value.output, report); err != nil {
			return 1
		}
		if report.ConformanceDecision != "PASS" {
			return 1
		}
		return 0
	}
	if value.head == "" || value.contract == "" || value.valid == "" || value.tampered == "" || value.coherentTampered == "" ||
		value.missing == "" || value.byteOnly == "" || value.wrongRecipe == "" || value.recipeOnly == "" || value.missingAttachment == "" || value.wrongAttachmentDigest == "" || value.unrelatedTampered == "" || value.staleHead == "" || value.unauthorizedConsumer == "" || value.claimProposition == "" || value.claimDependency == "" || value.claimProofChoice == "" || value.claimTarget == "" || value.source == "" ||
		value.operation == "" || value.recipe == "" || value.independence == "" || value.writeSet == "" || value.coherentOperation == "" || value.checkout == "" || value.output == "" {
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
	recipeOnly, recipeOnlyOK := read(value.recipeOnly)
	missingAttachment, missingAttachmentOK := read(value.missingAttachment)
	wrongAttachmentDigest, wrongAttachmentDigestOK := read(value.wrongAttachmentDigest)
	unrelatedTampered, unrelatedTamperedOK := read(value.unrelatedTampered)
	staleHead, staleHeadOK := read(value.staleHead)
	unauthorizedConsumer, unauthorizedConsumerOK := read(value.unauthorizedConsumer)
	claimProposition, claimPropositionOK := read(value.claimProposition)
	claimDependency, claimDependencyOK := read(value.claimDependency)
	claimProofChoice, claimProofChoiceOK := read(value.claimProofChoice)
	claimTarget, claimTargetOK := read(value.claimTarget)
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
	checkoutRaw, checkoutOK := read(value.checkout)
	if !validOK || !tamperedOK || !coherentTamperedOK || !missingOK || !byteOnlyOK || !wrongRecipeOK || !sourceOK || !operationOK || !recipeOK || !independenceOK ||
		!recipeOnlyOK || !missingAttachmentOK || !wrongAttachmentDigestOK || !unrelatedTamperedOK || !staleHeadOK || !unauthorizedConsumerOK || !claimPropositionOK || !claimDependencyOK || !claimProofChoiceOK || !claimTargetOK || !checkoutOK ||
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
	var checkout verifier.CheckoutEvidence
	if err := json.Unmarshal(checkoutRaw, &checkout); err != nil {
		return 2
	}
	interventions := []verifier.InterventionInput{
		{ID: "semantic-source-intervention", Kind: "SEMANTIC", Before: verifier.SubjectInput{Artifact: valid, Source: source, Operation: operation, Recipe: recipe}, After: verifier.SubjectInput{Artifact: semanticArtifact, Source: semanticSource, Operation: semanticOperation, Recipe: recipe}},
		{ID: "comment-only-intervention", Kind: "NONSEMANTIC", Before: verifier.SubjectInput{Artifact: valid, Source: source, Operation: operation, Recipe: recipe}, After: verifier.SubjectInput{Artifact: commentArtifact, Source: commentSource, Operation: commentOperation, Recipe: recipe}},
	}
	report := verifier.Evaluate(verifier.Input{Contract: contract, ContractBytes: contractRaw, HeadSHA: value.head, ValidArtifact: valid,
		TamperedArtifact: tampered, CoherentTamperedArtifact: coherentTampered, MissingArtifact: missing, ByteOnlyArtifact: byteOnly, WrongRecipe: wrongRecipe,
		RecipeOnlyArtifact: recipeOnly, MissingAttachment: missingAttachment, WrongAttachmentDigest: wrongAttachmentDigest, UnrelatedTamperedArtifact: unrelatedTampered, StaleHeadArtifact: staleHead, ClaimPropositionArtifact: claimProposition, ClaimDependencyArtifact: claimDependency, ClaimProofChoiceArtifact: claimProofChoice, ClaimTargetArtifact: claimTarget, UnauthorizedConsumer: unauthorizedConsumer,
		Source: source, Operation: operation, Recipe: recipe, Independence: independence, WriteSet: writeSet,
		CoherentOperation: coherentOperation, Interventions: interventions, Checkout: checkout, UnauthorizedBundle: readBundle(value.unauthorizedBundle)})
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
