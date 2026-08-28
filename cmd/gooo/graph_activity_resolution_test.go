package main

import (
	"bytes"
	"encoding/json"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"testing"
)

const activityResolutionSource = `package activitycardinality
namespace activitycardinality
entity Input id "gooo://activity-cardinality/entity/input"
entity Output id "gooo://activity-cardinality/entity/output"
activity ResolveOne(Input) -> Output
activity ResolveOther(Input) -> Output
`

func TestRunGraphActivityResolutionExposesZeroOneMany(t *testing.T) {
	cases := []struct {
		name string
		args []string
		code int
		want queryengine.ActivityResolutionDecision
		count int
	}{
		{"one", []string{"fixture.gooo", "--namespace", "activitycardinality", "--name", "ResolveOne"}, exitOK, queryengine.ActivityResolutionClosed, 1},
		{"zero", []string{"fixture.gooo", "--name", "Missing"}, exitFailure, queryengine.ActivityResolutionUnknown, 0},
		{"many", []string{"fixture.gooo", "--namespace", "activitycardinality"}, exitFailure, queryengine.ActivityResolutionRefuted, 2},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, code, stderr := runActivityResolutionFixture(t, test.args)
			if code != test.code || stderr != "" || result.Decision != test.want || result.Occurrences != test.count {
				t.Fatalf("resolution = %#v, code=%d, stderr=%q", result, code, stderr)
			}
			if result.Subject.SourceStatus != "bound" || result.Subject.SourceDigest == "" || result.Claim.Stage == "" || result.Claim.NextOperation == "" {
				t.Fatalf("resolution lost evidence context: %#v", result)
			}
		})
	}
}

func TestRunGraphActivityResolutionIsDeterministic(t *testing.T) {
	args := []string{"fixture.gooo", "--name", "ResolveOne"}
	first, firstCode, _ := runActivityResolutionBytes(args)
	second, secondCode, _ := runActivityResolutionBytes(args)
	if firstCode != exitOK || secondCode != exitOK || !bytes.Equal(first, second) {
		t.Fatalf("resolution replay differs: %s != %s", first, second)
	}
}

func runActivityResolutionFixture(t *testing.T, args []string) (queryengine.ActivityCardinalityResolution, int, string) {
	t.Helper()
	payload, code, stderr := runActivityResolutionBytes(args)
	var result queryengine.ActivityCardinalityResolution
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatal(err)
	}
	return result, code, stderr
}

func runActivityResolutionBytes(args []string) ([]byte, int, string) {
	var stdout, stderr bytes.Buffer
	code := runGraphActivityResolution(args, fixtureReader{source: activityResolutionSource}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}
