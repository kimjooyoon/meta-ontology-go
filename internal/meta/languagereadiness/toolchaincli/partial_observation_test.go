package toolchaincli

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

type partialObservationExecutor struct {
	invoke func([]string) (cliruntime.Observation, error)
}

func (fake partialObservationExecutor) BinaryDigest() (string, error) {
	return testDigest, nil
}

func (fake partialObservationExecutor) Invoke(arguments []string) (cliruntime.Observation, error) {
	return fake.invoke(arguments)
}

func TestPartialObservationsRemainUnresolvedAndPreserveEffects(t *testing.T) {
	for _, failingCall := range []int{1, 2} {
		for _, failureMode := range []string{"error", "marker"} {
			t.Run(fmt.Sprintf("call_%d_%s", failingCall, failureMode), func(t *testing.T) {
				var returned []cliruntime.Observation
				invoke := func(arguments []string) (cliruntime.Observation, error) {
					observation := fakeObservation(arguments)
					var err error
					if failingCall == 2 && failureMode == "error" && len(returned) == 0 {
						observation.RepositoryWrites = 2
						observation.TreeAfterDigest = digestJSON("first changed tree")
					}
					if len(returned)+1 >= failingCall {
						if failureMode == "error" {
							observation.RepositoryWrites = 3
							observation.Stderr = "retained partial stderr"
							err = errors.New("fixture observation incomplete")
						} else {
							observation.Failure = "INVOCATION_RESOURCE_UNKNOWN"
						}
					}
					returned = append(returned, observation)
					return observation, err
				}
				executor := partialObservationExecutor{invoke: invoke}
				result := executeCase(executor, expectedRegistry().Cases[0])
				if len(returned) != failingCall ||
					result.Status != "UNRESOLVED" || result.ReplayMatched {
					t.Fatal("partial observation was repeated or accepted", len(returned), result)
				}
				if !reflect.DeepEqual(result.First, returned[0]) {
					t.Fatal("first observation was discarded", result)
				}
				if failingCall == 2 && !reflect.DeepEqual(result.Replay, returned[1]) {
					t.Fatal("replay observation was discarded", result)
				}
				if failingCall == 1 && !reflect.DeepEqual(result.Replay, cliruntime.Observation{}) {
					t.Fatal("unrequested replay acquired observations", result)
				}
				wantInvocations, wantWrites := failingCall, 0
				if failureMode == "error" {
					wantInvocations, wantWrites = failingCall-1, 3+2*(failingCall-1)
				}
				summary := summarize(Source{}, []CaseResult{result}, 0)
				if result.Invocations != wantInvocations ||
					result.RepositoryWrites != wantWrites ||
					summary.Invocations != wantInvocations ||
					summary.RepositoryWrites != wantWrites ||
					summary.Unresolved != 1 || summary.Satisfied != 0 ||
					result.EvidenceDigest != caseDigest(result) {
					t.Fatal("partial facts changed counters or gained acceptance", result, summary)
				}
			})
		}
	}
}
