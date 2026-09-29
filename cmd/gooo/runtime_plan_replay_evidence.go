package main

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func observeRuntimePlanReplayEvidence(
	filename string,
	entry string,
	runtimePlanPath string,
	document runtimePlanDocument,
	runtimePlanDigest string,
	execution valueexecution.Execution,
	reverseObservation string,
) (valueexecution.GeneratedReplayEvidencePart01, valueexecution.ExecutionOriginReceipt) {
	toolchainDigest, evaluatorDigest := valueexecution.RuntimeEvidenceIdentities()
	evidenceDigest := cache.HashBytes([]byte(strings.Join([]string{
		document.SourceDigest,
		document.SemanticHash,
		document.TypedPlanDigest,
		runtimePlanDigest,
		execution.ExecutionDigest,
		reverseObservation,
		toolchainDigest,
		evaluatorDigest,
	}, "\x00"))).String()
	origin := provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:        filename,
		DeclarationSymbol:     "activity:" + entry,
		IRNode:                "typed-plan:" + document.TypedPlanDigest,
		GeneratedURI:          runtimePlanPath,
		GeneratedSymbol:       "runtime-plan:" + document.TypedPlanDigest,
		ReverseObservationURI: "execution-replay:" + execution.ExecutionDigest,
		ReverseObservation:    reverseObservation,
		MetricName:            "gooo.generated-runtime-plan-replay.v1",
		MetricValue:           "phase=" + string(execution.Phase),
		EvidenceDigest:        "sha256:" + evidenceDigest,
	})
	receipt := valueexecution.ObserveExecutionOrigin(origin, execution)
	evidence := valueexecution.GeneratedReplayEvidencePart01{
		SourceDigest:             document.SourceDigest,
		SemanticDigest:           document.SemanticHash,
		TypedPlanDigest:          document.TypedPlanDigest,
		RuntimePlanDigest:        runtimePlanDigest,
		GeneratedArtifactDigest:  runtimePlanDigest,
		ToolchainDigest:          toolchainDigest,
		EvaluatorDigest:          evaluatorDigest,
		ReverseObservationDigest: receipt.ReceiptDigest,
	}
	return evidence, receipt
}
