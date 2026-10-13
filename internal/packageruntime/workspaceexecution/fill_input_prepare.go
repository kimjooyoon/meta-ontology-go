package workspaceexecution

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

// Transient syntax material is rebuilt from the manifest on replay. It is not
// serialized into Program or substituted for the saved source identity.
type workspaceFillInput struct {
	pkg, filename, source, activity, activityID string
	loweredBody                                 string
	callEdits, recordEdits                      []BodySourceEdit
	reason                                      string
}

func prepareWorkspaceFillInputs(manifest packageruntime.Manifest, calls workspaceCalls, records workspaceRecordNames, needed map[string]bool) map[string]workspaceFillInput {
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
	result := map[string]workspaceFillInput{}
	for _, key := range keys {
		d := calls[key]
		if d.activity.Assembly == nil || d.activity.Assembly.Spec.FillPlan == nil {
			continue
		}
		id, err := workspaceActivityIdentity("gooo_workspace", d.ref.LoweredName, d.activity.ID)
		if err != nil {
			continue
		}
		input := workspaceFillInput{pkg: d.ref.PackagePath, filename: d.source, activity: d.ref.Activity,
			source: contents[d.ref.PackagePath+"\x00"+d.source]}
		input.activityID, err = workspaceActivityIdentity(d.namespace, d.ref.Activity, d.activity.ID)
		if err != nil {
			input.reason = "INPUT_ACTIVITY_ID_UNAVAILABLE"
		}
		body, _, callErr := calls.rewriteCalls(key, d, d.body, "computes", true, &input.callEdits)
		if callErr == nil {
			scope := records.constructors[key]
			body, callErr = records.rewriteConstructorsWithEdits(d.ref.PackagePath, body, true, []recordConstructorScope{scope}, &input.recordEdits)
		}
		input.loweredBody = body
		if callErr != nil || body != d.activity.ValueProgram {
			input.reason = "TEMPLATE_LOWERING_UNAVAILABLE"
		}
		result[id] = input
	}
	return result
}
