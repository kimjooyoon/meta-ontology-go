package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/viewlens"
)

type bodyViewProjection struct {
	ViewID           string `json:"view_id"`
	CanonicalPointer string `json:"canonical_pointer"`
	// References into canonical.inputs avoid copying declarations or graphs.
	InputRefs []int `json:"input_refs,omitempty"`
}

type bodyViewNext struct {
	Kind         string   `json:"kind"`
	Argv         []string `json:"argv,omitempty"`
	ValuePointer string   `json:"value_pointer,omitempty"`
}

type bodyViewOutput struct {
	Schema         string                        `json:"schema"`
	Status         string                        `json:"status"`
	Selection      viewlens.Selection            `json:"selection"`
	Projection     bodyViewProjection            `json:"projection"`
	Canonical      bodycodegen.TypedValueOrigins `json:"canonical"`
	NextOperations []bodyViewNext                `json:"next_operations"`
}

func makeBodyViewOutput(report bodycodegen.TypedValueOrigins, selection viewlens.Selection,
	options bodyViewOptions, filename string) bodyViewOutput {
	output := bodyViewOutput{Schema: "gooo/body-view/v1", Status: report.Status,
		Selection: selection, Canonical: report, NextOperations: []bodyViewNext{},
		Projection: bodyViewProjection{ViewID: selection.Choice.ViewID, CanonicalPointer: "#/canonical"}}
	if selection.Choice.ViewID == viewlens.NamedInputs {
		refs, supported := namedBodyViewRefs(report)
		if supported {
			output.Projection.InputRefs = refs
		} else {
			output.Projection.ViewID = viewlens.Full
			output.Selection.Choice.ViewID, output.Selection.Choice.Reason = viewlens.Full, "HINT_INPUT_LAYOUT_BOUND"
			output.Selection.Issues = append(output.Selection.Issues, "HINT_INPUT_LAYOUT_BOUND")
			output.Selection.Record = nil
		}
	}
	if output.Selection.Record != nil {
		output.NextOperations = append(output.NextOperations, bodyViewNext{Kind: "persist_view_event",
			Argv: []string{"view-lens", "record", "--event"}, ValuePointer: "#/selection/record_event"})
	}
	if report.Status != "RESOLVED" {
		output.NextOperations = append(output.NextOperations, bodyViewNext{Kind: "inspect_source_origins",
			Argv: []string{"body-origins", "--activity", options.activity, filename}})
	}
	return output
}

func namedBodyViewRefs(report bodycodegen.TypedValueOrigins) ([]int, bool) {
	// The current body grammar permits 16 parameters. Wider future reports use
	// the complete canonical view until this advisory layout supports them.
	seen := [16]bool{}
	refs := []int{}
	for index, input := range report.Inputs {
		if input.Declaration == nil {
			continue
		}
		port := input.Declaration.InputIndex
		if port < 0 || port >= len(seen) {
			return nil, false
		}
		if !seen[port] {
			refs, seen[port] = append(refs, index), true
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		return report.Inputs[refs[i]].Declaration.InputIndex < report.Inputs[refs[j]].Declaration.InputIndex
	})
	return refs, true
}

func writeBodyViewText(w io.Writer, output bodyViewOutput, source []byte) error {
	selection := output.Selection
	if _, err := fmt.Fprintf(w, "View: %s\nSelection: %s\nHint: %s\n",
		selection.Choice.ViewID, selection.Choice.Reason, selection.HintState); err != nil {
		return err
	}
	if selection.Choice.ViewID == viewlens.Full {
		return bodycodegen.WriteTypedOriginText(w, output.Canonical, source)
	}
	if _, err := fmt.Fprintf(w, "Activity: %s\nSource digest: %s\nStatus: %s\n",
		output.Canonical.ActivityID, output.Canonical.SourceSHA256, output.Status); err != nil {
		return err
	}
	for _, ref := range output.Projection.InputRefs {
		input := output.Canonical.Inputs[ref].Declaration
		if _, err := fmt.Fprintf(w, "%s name=%q type=%q id=%q source=[%d,%d) status=%s\n",
			input.Port, input.Name, input.TypeName, input.TypeID, input.Start, input.End, input.Status); err != nil {
			return err
		}
	}
	return nil
}
