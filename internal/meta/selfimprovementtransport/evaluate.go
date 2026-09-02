package selfimprovementtransport

import (
	"encoding/json"
	"io/fs"
)

var obligationOrder = []string{
	"source-repository-commit",
	"producer-checkout-binding",
	"producer-run-identity",
	"logical-subject-digest",
	"immutable-artifact-locator",
	"artifact-archive-digest",
	"consumer-digest-replay",
	attestationObligation,
}

func Evaluate(repository fs.FS, contractPath, expectedRepository string, expectedRunID int64,
	observationRaw, producerRaw, metadataRaw []byte, actualArchiveDigest string) Report {
	contract, contractErr := CompileContract(repository, contractPath)
	var source observationHeader
	var producer ProducerReceipt
	var metadata TransportMetadata
	sourceErr := json.Unmarshal(observationRaw, &source)
	producerErr := json.Unmarshal(producerRaw, &producer)
	metadataErr := json.Unmarshal(metadataRaw, &metadata)
	report := Report{
		Schema: ReportSchema, MetricID: MetricID, Contract: contract,
		SubjectSHA: source.SubjectSHA, OrchestrationHeadSHA: metadata.OrchestrationHeadSHA,
		SourceObservationDigest: digestBytes(observationRaw), ActualArchiveDigest: actualArchiveDigest,
		Producer: producer, Transport: metadata,
		NotClaimed: []string{"artifact-name-proves-subject", "archive-digest-authenticates-producer", "whole-language-transport-complete"},

		Obligations: evaluateObligations(evaluationInput{
			contract: contract, contractErr: contractErr, expectedRepository: expectedRepository,
			expectedRunID: expectedRunID, source: source, sourceErr: sourceErr,
			observationRaw: observationRaw, producer: producer, producerErr: producerErr,
			metadata: metadata, metadataErr: metadataErr, actualArchiveDigest: actualArchiveDigest,
		})}
	reduce(&report)
	report.Digest = reportDigest(report)
	return report
}
