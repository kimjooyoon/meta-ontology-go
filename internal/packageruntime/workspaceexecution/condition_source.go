package workspaceexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"sort"
)

type BodySourceEdit struct {
	Start, End  int
	Replacement string
}

type WorkspaceConditionOrigin struct {
	ActivityID string                           `json:"lowered_activity_id"`
	BodySHA256 string                           `json:"lowered_body_sha256"`
	Start      int                              `json:"lowered_decoded_start"`
	End        int                              `json:"lowered_decoded_end"`
	Input      bodycodegen.ConditionInputSource `json:"input_source"`
}

func prepareWorkspaceConditionSources(manifest packageruntime.Manifest, calls workspaceCalls, records workspaceRecordNames, needed map[string]bool) ([]WorkspaceConditionOrigin, error) {
	contents := map[string]string{}
	for _, pkg := range manifest.Packages {
		for _, source := range pkg.Sources {
			contents[pkg.Path+"\x00"+source.Filename] = source.Content
		}
	}
	keys := make([]string, 0, len(needed))
	for key := range needed {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var result []WorkspaceConditionOrigin
	for _, key := range keys {
		d := calls[key]
		if d.body == "" {
			continue
		}
		ranges, err := bodycodegen.BodyConditionSourceRanges(d.body, len(d.activity.Inputs), syntax.ActivityDeclaredInputNames(d.activity)...)
		if err != nil {
			return nil, err
		}
		if len(ranges) == 0 {
			continue
		}
		var callEdits, recordEdits []BodySourceEdit
		lowered, _, err := calls.rewriteCalls(key, d, d.body, "computes", true, &callEdits)
		if err != nil {
			return nil, err
		}
		scope := records.constructors[key]
		lowered, err = records.rewriteConstructorsWithEdits(d.ref.PackagePath, lowered, true, []recordConstructorScope{scope}, &recordEdits)
		if err != nil {
			return nil, err
		}
		originals := append([]bodycodegen.BodyConditionOrigin(nil), ranges...)
		applyConditionSourceEdits(ranges, callEdits)
		applyConditionSourceEdits(ranges, recordEdits)
		file := contents[d.ref.PackagePath+"\x00"+d.source]
		id, err := workspaceActivityIdentity("gooo_workspace", d.ref.LoweredName, d.activity.ID)
		if err != nil {
			return nil, err
		}
		originalID, err := workspaceActivityIdentity(d.namespace, d.ref.Activity, d.activity.ID)
		if err != nil {
			return nil, err
		}
		bodySHA, err := bodycodegen.NormalizedConditionBodySHA256(lowered)
		if err != nil {
			return nil, err
		}
		for i, cursor := range ranges {
			input := workspaceInputConditionSource(file, d, originalID, originals[i])
			if lowered != d.activity.ValueProgram || cursor.Relation != "retained_condition_node" {
				input.Status = "UNMAPPED"
				input.Reason = "LOWERING_RANGE_DIFFERS"
			}
			result = append(result, WorkspaceConditionOrigin{id, bodySHA, cursor.Start, cursor.End, input})
		}
	}
	return result, nil
}

func applyConditionSourceEdits(ranges []bodycodegen.BodyConditionOrigin, edits []BodySourceEdit) {
	sort.Slice(edits, func(i, j int) bool { return edits[i].Start > edits[j].Start })
	for _, edit := range edits {
		delta := len(edit.Replacement) - (edit.End - edit.Start)
		for i := range ranges {
			c := &ranges[i]
			if edit.End <= c.Start {
				c.Start += delta
				c.End += delta
			} else if edit.Start >= c.End {
				continue
			} else if edit.Start >= c.Start && edit.End <= c.End {
				c.End += delta
			} else {
				c.Relation = "unmapped_crossing_edit"
			}
		}
	}
}

func workspaceInputConditionSource(file string, d workspaceCallDeclaration, activityID string, origin bodycodegen.BodyConditionOrigin) bodycodegen.ConditionInputSource {
	input := bodycodegen.ConditionInputSource{Status: "UNMAPPED", Reason: "INPUT_LITERAL_UNAVAILABLE", View: "workspace_input_file", PackagePath: d.ref.PackagePath, Filename: d.source, ActivityID: activityID, SourceSHA256: sourceSHA256([]byte(file)), Expression: origin.Expression}
	span := d.activity.ValueProgramSpan
	declaration := &bodycodegen.ConditionDeclarationSource{Status: "UNMAPPED", Reason: "DECLARATION_RANGE_UNAVAILABLE",
		View: "workspace_input_file", Kind: "activity", StableID: activityID, Name: d.ref.Activity, SourceSHA256: input.SourceSHA256}
	declSpan := d.activity.Span
	if declSpan.Start.Offset >= 0 && declSpan.End.Offset > declSpan.Start.Offset && declSpan.End.Offset <= len(file) {
		declaration.Status, declaration.Reason = "MAPPED", ""
		declaration.Start, declaration.End = declSpan.Start.Offset, declSpan.End.Offset
	}
	input.Declaration = declaration
	if file == "" || span.Start.Offset < 0 || span.End.Offset > len(file) || span.End.Offset < span.Start.Offset {
		return input
	}
	start, end, err := bodycodegen.SourceLiteralByteRange(file[span.Start.Offset:span.End.Offset], d.body, origin.Start, origin.End)
	if err != nil {
		return input
	}
	input.Status, input.Reason = "MAPPED", ""
	input.Start, input.End = span.Start.Offset+start, span.Start.Offset+end
	return input
}

func bindWorkspaceConditionSources(program Program, runtime *bodyexecution.CompositionRuntime, fills ...[]BodyFillStep) {
	for _, steps := range fills {
		bindWorkspaceFillConditionSources(runtime, steps)
		bindWorkspaceFillLocalSources(runtime, steps)
	}
	type key struct {
		activity, body string
		start, end     int
	}
	index := map[key]bodycodegen.ConditionInputSource{}
	activities, bodies := map[string]bool{}, map[string]bool{}
	for _, origin := range program.ConditionOrigins {
		index[key{origin.ActivityID, origin.BodySHA256, origin.Start, origin.End}] = origin.Input
		activities[origin.ActivityID], bodies[origin.ActivityID+"\x00"+origin.BodySHA256] = true, true
	}
	bindWorkspaceFillInputSources(program, runtime)
	for i := range runtime.ConditionSites {
		site := &runtime.ConditionSites[i]
		if site.Source == nil || site.Source.Status != "MAPPED" {
			continue
		}
		reason := "INPUT_ORIGIN_UNAVAILABLE"
		if activities[site.ActivityID] {
			reason = "SELECTED_BODY_DIFFERS"
		}
		if bodies[site.ActivityID+"\x00"+site.Source.BodySHA256] {
			reason = "INPUT_RANGE_UNAVAILABLE"
		}
		input := &bodycodegen.ConditionInputSource{Status: "UNMAPPED", Reason: reason, View: "workspace_input_file"}
		if original, found := index[key{site.ActivityID, site.Source.BodySHA256, site.Source.DecodedStart, site.Source.DecodedEnd}]; found {
			input = &original
		}
		site.Source.Input = input
	}
}
