package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	producer "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageproofartifact"
)

func run(args []string) int {
	flags := flag.NewFlagSet("language-proof-carrying-artifact", flag.ContinueOnError)
	head := flags.String("head", "", "exact subject commit")
	sourcePath := flags.String("source-path", "", "Gooo source path")
	source := flags.String("source", "", "Gooo source bytes")
	operation := flags.String("operation", "", "operation receipt")
	recipe := flags.String("recipe", "", "canonical verification recipe")
	writeSet := flags.String("write-set", "", "repository write-set observation")
	out := flags.String("out", "", "artifact output")
	if flags.Parse(args) != nil || *head == "" || *sourcePath == "" || *source == "" || *operation == "" || *recipe == "" || *writeSet == "" || *out == "" {
		return 2
	}
	sourceBytes, err := os.ReadFile(*source)
	if err != nil {
		return 2
	}
	operationBytes, err := os.ReadFile(*operation)
	if err != nil {
		return 2
	}
	recipeBytes, err := os.ReadFile(*recipe)
	if err != nil {
		return 2
	}
	writeSetBytes, err := os.ReadFile(*writeSet)
	if err != nil {
		return 2
	}
	var recipeValue producer.Recipe
	if err := json.Unmarshal(recipeBytes, &recipeValue); err != nil {
		return 2
	}
	var writeSetValue producer.WriteSetObservation
	if err := json.Unmarshal(writeSetBytes, &writeSetValue); err != nil {
		return 2
	}
	artifact, err := producer.Generate(producer.Input{HeadSHA: *head, SourcePath: *sourcePath,
		Source: sourceBytes, Operation: operationBytes, Recipe: recipeValue, WriteSet: writeSetValue})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := producer.Write(*out, artifact); err != nil {
		return 1
	}
	fmt.Printf("proof-carrying artifact: %s evidence=%d authority=%s digest=%s\n",
		artifact.Decision, len(artifact.Evidence), artifact.Authority.ArtifactUseAuthority, artifact.Digest)
	return 0
}
