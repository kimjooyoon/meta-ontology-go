package main

import (
	"fmt"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/generator"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicdiscovery"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func runGenerate(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseGenerateArguments(args)
	if err != nil {
		if jsonMode {
			return reportUsage(true, stdout, stderr, "generate", err.Error())
		}
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	input, code := readGenerateInput(options, reader, parser, jsonMode, stdout, stderr, deadline)
	if code != exitOK {
		return code
	}
	started := time.Now()
	if options.compatibilityCertificateFilename != "" {
		return runCompatibilityGenerate(options, input, reader, jsonMode, stdout, stderr, deadline)
	}
	if options.continuityCertificateFilename != "" {
		return runPublicContinuityGenerate(options, input, reader, jsonMode, stdout, stderr, deadline)
	}
	if options.publicRetentionRequested() {
		return runPublicGenerate(options, input, reader, parser, jsonMode, stdout, stderr, deadline)
	}
	if options.retentionReport {
		return runBaselineGenerateReport(options, input, jsonMode, stdout, stderr, deadline)
	}
	artifacts, code := buildGenerateArtifacts(options, input, jsonMode, stdout, stderr, deadline)
	if code != exitOK {
		return code
	}
	if code := writeGenerateArtifacts(artifacts, jsonMode, stdout, stderr); code != exitOK {
		return code
	}
	var discovery *publicdiscovery.Result
	if options.observationLedgerDir != "" {
		manifestDigest, err := publicObservationManifestDigest(artifacts.manifest)
		if err != nil {
			return reportGenerateError(jsonMode, stdout, stderr, options.observationLedgerDir, "observation.manifest", "observation manifest", err, input.file)
		}
		observed, err := publicdiscovery.Record(options.observationLedgerDir, publicdiscovery.Input{
			SourceDigest: cache.HashBytes(input.source).String(), InputSemanticDigest: artifacts.ir.StableHash(),
			PreviousGoDigest: cache.HashBytes(input.previousGo).String(), ToolchainDigest: generation.SemanticRetentionToolchainDigest(),
			ContractDigest: publicdiscovery.PolicySourceDigest(), EvaluatorDigest: publicdiscovery.GeneratedEvaluatorDigest(),
			GeneratedSemanticDigest: artifacts.ir.StableHash(), GeneratedOutputDigest: cache.HashBytes(artifacts.result.Source).String(),
			GeneratedManifestDigest: manifestDigest,
		}, int64(time.Since(started)/time.Millisecond), readPeakRSSKib())
		if err != nil {
			return reportGenerateError(jsonMode, stdout, stderr, options.observationLedgerDir, "observation.record", "record observation", err, input.file)
		}
		discovery = &observed
	}
	return reportGenerateSuccess(options, input, artifacts, discovery, jsonMode, stdout)
}

func publicObservationManifestDigest(manifest projectionManifest) (string, error) {
	// The ordinary manifest contains the caller's output path. Normalize that
	// path so an identical generate remains comparable across output folders.
	normalized := manifest
	normalized.GeneratedFile = generatedFileName
	data, err := jsonManifestBytes(normalized)
	if err != nil {
		return "", err
	}
	return cache.HashBytes(data).String(), nil
}
func projectionIR(ir semantic.IR) (generator.SemanticIR, error) {
	if err := ir.Validate(); err != nil {
		return generator.SemanticIR{}, err
	}
	if len(ir.RuntimeBindings) != 0 {
		return generator.SemanticIR{}, errRuntimeBindingsUnsupportedByGenerator
	}
	model := generator.SemanticIR{Package: ir.Package}
	entities := make(map[string]int)
	activities := make(map[string]int)
	for _, node := range ir.Graph.Nodes() {
		item := generatorNode(node)
		switch node.Kind {
		case semantic.Entity:
			fields, err := projectionSemanticFields(node)
			if err != nil {
				return generator.SemanticIR{}, err
			}
			item.entity.Fields = fields
			entities[string(node.ID)] = len(model.Entities)
			model.Entities = append(model.Entities, item.entity)
		case semantic.Activity:
			activities[string(node.ID)] = len(model.Activities)
			model.Activities = append(model.Activities, item.activity)
		}
	}
	for _, fact := range ir.Graph.DeterministicFacts() {
		appendFactPort(&model, entities, activities, fact)
	}
	return model, nil
}

type generatorNodeResult struct {
	entity   generator.Entity
	activity generator.Activity
}

func generatorNode(node semantic.Node) generatorNodeResult {
	span := generatorSpan(node.Span)
	entity := generator.Entity{ID: string(node.ID), Name: node.Name, GoName: node.Name, Source: span}
	activity := generator.Activity{ID: string(node.ID), Name: node.Name, GoName: node.Name, Source: span}
	return generatorNodeResult{entity: entity, activity: activity}
}
