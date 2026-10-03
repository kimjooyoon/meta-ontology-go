package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestNativeStreamMixesRecipesAndFullDocuments(t *testing.T) {
	read := func(name string) []byte {
		b, err := os.ReadFile("../../examples/body-codegen/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	source, raw := read("path-recipe.gooo.fixture"), read("path-recipe.json")
	var input bytes.Buffer
	for _, id := range []string{"recipe-first", "recipe-second"} {
		r, err := json.Marshal(Request{Schema: RequestSchema, CorrelationID: id, Source: string(source), Activity: "Compose", Document: raw})
		if err != nil {
			t.Fatal(err)
		}
		input.Write(r)
		input.WriteByte('\n')
	}
	input.Write(requestLine(t, "full", string(read("typed-path-compound.gooo.fixture"))))
	input.WriteByte('\n')
	g, err := bodycodegen.NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Run(context.Background(), g, io.NopCloser(&input), asWriteCloser(&out), 2); err != nil {
		t.Fatal(err)
	}
	results := decodeResults(t, out.Bytes())
	if len(results) != 3 {
		t.Fatal("lost mixed-schema request")
	}
	for _, result := range results {
		if result.Status != "completed" || result.Response == nil || result.Response.Report.BodyPaths.FunctionalCompleteness != 100 {
			t.Fatal("mixed schema generation failed", result)
		}
	}
}
