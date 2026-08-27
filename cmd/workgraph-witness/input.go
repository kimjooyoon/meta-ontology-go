package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/workgraph"
)

func loadInput(value options) (workgraph.Contract, workgraph.Observation, error) {
	started := time.Now()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	contractData, err := os.ReadFile(value.contract)
	if err != nil { return workgraph.Contract{}, workgraph.Observation{}, err }
	var contract workgraph.Contract
	if err := json.Unmarshal(contractData, &contract); err != nil { return contract, workgraph.Observation{}, err }
	source, err := os.ReadFile(value.source)
	if err != nil { return contract, workgraph.Observation{}, err }
	check, err := os.ReadFile(value.checkReceipt)
	if err != nil { return contract, workgraph.Observation{}, err }
	generated, err := readOptional(value.generated)
	if err != nil { return contract, workgraph.Observation{}, err }
	replay, err := readOptional(value.replay)
	if err != nil { return contract, workgraph.Observation{}, err }
	prior, priorDigest, err := readPredecessor(value.predecessor)
	if err != nil { return contract, workgraph.Observation{}, err }
	resource, err := resourceSample(value.resource, started, before)
	if err != nil { return contract, workgraph.Observation{}, err }
	observation := workgraph.Observation{
		HeadSHA: value.head, SourcePath: value.source, SourceText: string(source), SourceDigest: workgraph.DigestBytes(source),
		CheckDigest: workgraph.DigestBytes(check), GeneratedDigest: optionalDigest(generated), ReplayDigest: optionalDigest(replay),
		GeneratedBytes: int64(len(generated)), Resource: resource, Predecessor: prior, PredecessorDigest: priorDigest,
	}
	if observation.SourcePath != contract.Source { return contract, observation, fmt.Errorf("source path does not match contract authority") }
	return contract, observation, nil
}
