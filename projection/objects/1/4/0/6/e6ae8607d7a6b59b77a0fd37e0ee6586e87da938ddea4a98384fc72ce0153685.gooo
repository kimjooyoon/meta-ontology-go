package query

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestInferenceProjectionReplayIsRaceSafeAndReadOnly(t *testing.T) {
	path, _ := inferenceQueryFixture(t)
	before := path
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	request := inferenceQueryRequest()
	request.IncludeClaims = true
	request.IncludeEvidence = true
	want, err := projection.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	wantDigest := want.CanonicalDigestValue()
	const workers = 8
	const replays = 20
	var wait sync.WaitGroup
	errorsCh := make(chan error, workers*replays)
	for range workers {
		wait.Go(func() {
			for range replays {
				got, replayErr := projection.Execute(request)
				if replayErr != nil || got.CanonicalDigestValue() != wantDigest {
					errorsCh <- errors.New("inference replay digest changed")
				}
			}
		})
	}
	wait.Wait()
	close(errorsCh)
	for replayErr := range errorsCh {
		t.Fatal(replayErr)
	}
	if !reflect.DeepEqual(path, before) || projection.StableHash() != projection.Path().StableHash() {
		t.Fatal("inference projection mutated input or normalized snapshot")
	}
	candidateRequest := inferenceQueryRequest()
	candidateRequest.Kind = semantic.InferenceObservationCandidate
	candidate, err := projection.Execute(candidateRequest)
	if err != nil || len(candidate.Edges) != 1 || candidate.Edges[0].AuthorityLayer != semantic.AuthorityCandidate ||
		candidate.Edges[0].AuthorityEffect != semantic.AuthorityObserve {
		t.Fatalf("candidate isolation = %#v err=%v", candidate, err)
	}
	graph := New()
	if err := graph.Add(NewFact(ID("inference-query://graph/activity"), Used, ID("inference-query://graph/entity"))); err != nil {
		t.Fatal(err)
	}
	graphCanonical, graphHash := graph.Canonical(), graph.StableHash()
	if _, err := projection.Execute(request); err != nil {
		t.Fatal(err)
	}
	if graph.Canonical() != graphCanonical || graph.StableHash() != graphHash {
		t.Fatal("inference projection mutated authority graph")
	}
}
