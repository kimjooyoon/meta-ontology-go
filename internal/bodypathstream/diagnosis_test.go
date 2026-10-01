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

func TestNativeStreamDiagnosisDoesNotPoisonNextRequest(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/path-diagnosis.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile("../../examples/body-codegen/path-diagnosis-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	generator, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	request := Request{Schema: RequestSchema, CorrelationID: "valid", Source: string(source), Activity: "Probe", Document: document,
		Options: bodycodegen.TypedPathOptions{Diagnosis: &bodycodegen.PathDiagnosisOptions{Inputs: []int64{2, 3}, MaxCandidates: 2}}}
	valid, _ := json.Marshal(request)
	request.CorrelationID = "invalid"
	request.Options.Diagnosis.MaxCandidates = 65
	invalid, _ := json.Marshal(request)
	var output bytes.Buffer
	if err = Run(context.Background(), generator, io.NopCloser(bytes.NewReader(bytes.Join([][]byte{invalid, valid}, []byte{'\n'}))),
		asWriteCloser(&output), 2); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, output.Bytes())
	sort.Slice(results, func(i, j int) bool { return results[i].Sequence < results[j].Sequence })
	if len(results) != 2 || results[0].Status != "rejected" || results[0].Response != nil || results[1].Status != "completed" {
		t.Fatal("diagnostic option failure changed subsequent construction")
	}
	receipt := results[1].Response.Report.BodyPaths
	if receipt == nil || receipt.Diagnosis == nil || receipt.Diagnosis.ProbeDistinguished != 1 ||
		receipt.Search.Selection.ModelCalls != 0 || receipt.Diagnosis.ModelPredictions != 0 || results[1].Response.Report.RepositoryWrites != 0 {
		t.Fatal("worker omitted native deterministic diagnosis")
	}
}
