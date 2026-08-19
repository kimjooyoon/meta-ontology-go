package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
)

type generationResult struct {
	ir     semantic.IR
	result generator.Result
	err    error
}
type generateInput struct {
	source      []byte
	file        *syntax.File
	diagnostics syntax.Diagnostics
	previousGo  []byte
}
type generateArtifacts struct {
	ir           semantic.IR
	result       generator.Result
	output       string
	manifestPath string
	manifest     projectionManifest
	cleanupDirs  []string
}

func reportGenerateUsage(jsonMode bool, stdout, stderr io.Writer, err error) int {
	if jsonMode {
		return reportUsage(true, stdout, stderr, "generate", err.Error())
	}
	fmt.Fprintln(stderr, err)
	return exitUsage
}
