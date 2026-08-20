package query

import (
	"errors"
	"sync"
	"testing"
)

func TestDatalogReplayIsRaceSafe(t *testing.T) {
	ir := workspaceIR(t, "billing", "billing://", false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	request := workspaceDatalogRequest()
	want, err := graph.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	wantDigest, err := want.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}

	const workers = 8
	const replays = 20
	var wait sync.WaitGroup
	errorsCh := make(chan error, workers*replays)
	for range workers {
		wait.Go(func() {
			for range replays {
				got, replayErr := graph.EvaluateDatalog(request)
				if replayErr != nil {
					errorsCh <- replayErr
					continue
				}
				digest, digestErr := got.CanonicalDigest()
				if digestErr != nil || digest != wantDigest {
					errorsCh <- errors.New("Datalog replay digest changed")
				}
			}
		})
	}
	wait.Wait()
	close(errorsCh)
	for replayErr := range errorsCh {
		t.Fatal(replayErr)
	}
}
