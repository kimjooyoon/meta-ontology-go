package workspaceexecution

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type WorkspaceActivityOrigin struct {
	ActivityID string                           `json:"lowered_activity_id"`
	Input      bodycodegen.ConditionInputSource `json:"input_source"`
}

type WorkspaceFillOrigin struct {
	ActivityID   string                             `json:"lowered_activity_id"`
	SourceSHA256 string                             `json:"assembly_source_sha256"`
	BodySHA256   string                             `json:"selected_body_sha256"`
	CandidateID  string                             `json:"candidate_id"`
	Input        *bodycodegen.SourceFillInputOrigin `json:"input_source"`
}

func prepareWorkspaceActivityOrigins(manifest packageruntime.Manifest, calls workspaceCalls, needed map[string]bool) ([]WorkspaceActivityOrigin, error) {
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
	var out []WorkspaceActivityOrigin
	for _, key := range keys {
		d := calls[key]
		id, err := workspaceActivityIdentity("gooo_workspace", d.ref.LoweredName, d.activity.ID)
		if err != nil {
			return nil, err
		}
		original, err := workspaceActivityIdentity(d.namespace, d.ref.Activity, d.activity.ID)
		if err != nil {
			return nil, err
		}
		input := workspaceInputConditionSource(contents[d.ref.PackagePath+"\x00"+d.source], d, original, bodycodegen.BodyConditionOrigin{Start: 0, End: len(d.body)})
		out = append(out, WorkspaceActivityOrigin{id, input})
	}
	return out, nil
}

func workspaceFillOrigins(program Program, result Result) ([]WorkspaceFillOrigin, error) {
	generations, err := originQueryGenerations(result)
	if err != nil {
		return nil, err
	}
	var out []WorkspaceFillOrigin
	seen := map[string]bool{}
	for _, generation := range generations {
		fill := generation.Report.BodyFill
		if fill == nil || fill.SourceOrigin == nil {
			continue
		}
		o := fill.SourceOrigin
		key := o.ActivityID + "\x00" + o.SourceSHA256 + "\x00" + o.SelectedBodySHA256 + "\x00" + o.CandidateID
		if seen[key] {
			continue
		}
		seen[key] = true
		input, ok := program.fillInputs[o.ActivityID]
		if !ok {
			continue
		}
		link := &bodycodegen.SourceFillConditionOrigin{CandidateID: o.CandidateID, Fragments: o.Fragments}
		out = append(out, WorkspaceFillOrigin{o.ActivityID, o.SourceSHA256, o.SelectedBodySHA256, o.CandidateID, workspaceFillInputOrigin(input, link)})
	}
	return out, nil
}

func collectWorkspaceFillOrigins(program Program, result Result) ([]WorkspaceFillOrigin, string) {
	origins, err := workspaceFillOrigins(program, result)
	if err != nil {
		return nil, err.Error()
	}
	return origins, ""
}

func checkWorkspaceFillOrigins(program Program, prior Result) error {
	if prior.FillOrigins == nil {
		return nil
	}
	current, err := workspaceFillOrigins(program, prior)
	if err != nil {
		return err
	}
	a, _ := json.Marshal(current)
	b, _ := json.Marshal(prior.FillOrigins)
	if string(a) != string(b) {
		return fmt.Errorf("saved workspace original fill locations differ from current source declarations")
	}
	return nil
}
