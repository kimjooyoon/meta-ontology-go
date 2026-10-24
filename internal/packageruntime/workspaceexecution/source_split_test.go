package workspaceexecution

import (
	"context"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestSourceSplitPreservesNativeCallsBindingsAndOrigins(t *testing.T) {
	manifest := namedInputWorkspace()
	split, err := packageruntime.SplitSource(manifest, "named/app", "금액.gooo", "urn:named:net", "00-net.gooo")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
 {"inputs":{"named/app:Credit":79,"named/app:Net.input1":31},"expected":{"named/app:Credit":102,"named/app:Net":71}},
 {"inputs":{"named/app:Credit":-41,"named/app:Net.input1":7},"expected":{"named/app:Credit":-18,"named/app:Net":-25}},
 {"inputs":{"named/app:Credit":0,"named/app:Net.input1":23},"expected":{"named/app:Credit":23,"named/app:Net":0}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	for label, candidate := range map[string]packageruntime.Manifest{"before": manifest, "after": split.Manifest} {
		t.Run(label, func(t *testing.T) {
			result, err := ExecuteWorkspace(context.Background(), candidate, suite, "", "")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Runtime
			if r.FinitePassed != 6 || r.FiniteTotal != 6 || !r.RuntimeReplayed || r.ModelCalls != 0 {
				t.Fatal("independent output oracle or replay failed", r)
			}
			flow, err := InspectSelectedValueOrigins(context.Background(), result, OriginQuery{Limit: 128})
			if err != nil || flow.Analysis != "MAPPED" || flow.Truncated {
				t.Fatal(flow, err)
			}
			checkNamedInputPaths(t, flow)
			if label == "after" {
				checkSplitBodyOrigin(t, result, split, manifest.Packages[1].Sources[0].Content)
			}
		})
	}
}

func checkSplitBodyOrigin(t *testing.T, result Result, split packageruntime.SourceSplit, original string) {
	t.Helper()
	for _, origin := range result.Program.ActivityOrigins {
		input := origin.Input
		if input.ActivityID != "urn:named:net" {
			continue
		}
		file := split.Files[1]
		if input.Filename != file.Filename || input.SourceSHA256 != file.SHA256 || input.Status != "MAPPED" {
			t.Fatal("moved body is not attributed to the new source", input)
		}
		for _, span := range file.Origins {
			if input.Start >= span.OutputStart && input.End <= span.OutputEnd {
				start := span.InputStart + input.Start - span.OutputStart
				end := span.InputStart + input.End - span.OutputStart
				if original[start:end] != file.Source[input.Start:input.End] {
					t.Fatal("body cannot be traced through the split to its original bytes")
				}
				return
			}
		}
		t.Fatal("body has no original byte range")
	}
	t.Fatal("missing moved activity origin")
}
