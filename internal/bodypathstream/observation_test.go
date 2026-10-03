package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestNativeStreamOwnsProbeSessionsAndRejectsNullInputs(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile("../../examples/body-codegen/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	request := Request{Schema: RequestSchema, CorrelationID: "probe", Activity: "Probe",
		Source: string(read("path-observation.gooo.fixture")), Document: read("path-observation-plan.json"),
		Options: bodycodegen.TypedPathOptions{Observation: &bodycodegen.PathObservationOptions{
			Inputs: []int64{2, 3, 0}, MaxCandidates: 2, MaxRounds: 2, OracleActivity: "Expected"}}}
	fresh, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	request.Options.Observation.ReuseProbeOutputs = true
	reuse, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	invalid := bytes.Replace(reuse, []byte(`"inputs":[2,3,0]`), []byte(`"inputs":[2,null,0]`), 1)
	var out bytes.Buffer
	raw := bytes.Join([][]byte{invalid, reuse, fresh, reuse}, []byte{'\n'})
	if err := Run(context.Background(), g, io.NopCloser(bytes.NewReader(raw)), asWriteCloser(&out), 3); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, out.Bytes())
	sort.Slice(results, func(i, j int) bool { return results[i].Sequence < results[j].Sequence })
	if len(results) != 4 || results[0].Status != "rejected" {
		t.Fatal("null probe accepted or records lost")
	}
	for i, r := range results[1:] {
		if r.Status != "completed" || r.Response == nil {
			t.Fatal(r)
		}
		p := r.Response.Report.BodyPaths
		if p.Observation.Options.ReuseProbeOutputs != (i != 1) || p.Observation.Status != "ONE_SURVIVING_CANDIDATE" ||
			p.Observation.EffectiveCases != 2 || p.Search.Selection.ModelCalls != 0 {
			t.Fatal("request inherited another session", p)
		}
		if r.Response.Source != results[1].Response.Source {
			t.Fatal("mode changed generated source")
		}
	}
}
