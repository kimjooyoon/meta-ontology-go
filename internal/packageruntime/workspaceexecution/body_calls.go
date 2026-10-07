package workspaceexecution

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// WorkspacePureCalls maps the source package identities to generated call names.
// Sites describe declared calls, including alternatives that may not be selected.
type WorkspacePureCalls struct {
	Schema     string                  `json:"schema"`
	Activities []WorkspaceCallActivity `json:"activities"`
	Sites      []WorkspaceCallSite     `json:"sites"`
}

type WorkspaceCallActivity struct {
	Activity         ActivityRef `json:"activity"`
	ActivityID       string      `json:"activity_id"`
	LoweredID        string      `json:"lowered_activity_id"`
	Source           string      `json:"source"`
	SourceBodySHA256 string      `json:"source_body_sha256"`
}

type WorkspaceCallSite struct {
	Caller  string `json:"caller"`
	Callee  string `json:"callee"`
	Surface string `json:"surface"`
	Start   int    `json:"start_byte"`
	End     int    `json:"end_byte"`
}

type workspaceCallDeclaration struct {
	ref       ActivityRef
	source    string
	namespace string
	body      string
	imports   []syntax.ImportDecl
	activity  *syntax.ActivityDecl
}

type workspaceCalls map[string]workspaceCallDeclaration

func indexWorkspaceCalls(files map[string][]*syntax.File, names map[string]map[string]string) workspaceCalls {
	result := workspaceCalls{}
	for pkg, sources := range files {
		for _, file := range sources {
			for _, declaration := range file.Declarations {
				if activity, ok := declaration.(*syntax.ActivityDecl); ok {
					ref := ActivityRef{pkg, activity.Name, names[pkg][activity.Name]}
					result[packageActivityKey(pkg, activity.Name)] = workspaceCallDeclaration{
						ref: ref, source: activity.Span.Filename, namespace: file.Namespace.Name,
						body: activity.ValueProgram, imports: file.Imports, activity: activity}
				}
			}
		}
	}
	return result
}

func (c workspaceCalls) expand(roots map[string]bool) (map[string]bool, *WorkspacePureCalls, error) {
	needed := map[string]bool{}
	var queue []string
	for key := range roots {
		needed[key], queue = true, append(queue, key)
	}
	sort.Strings(queue)
	report := &WorkspacePureCalls{Schema: "gooo/workspace-pure-calls/v1"}
	participants := map[string]bool{}
	for i := 0; i < len(queue); i++ {
		key := queue[i]
		declaration := c[key]
		sites, err := c.rewriteActivityCalls(key, declaration)
		if err != nil {
			return nil, nil, fmt.Errorf("activity %s: %w", key, err)
		}
		report.Sites = append(report.Sites, sites...)
		for _, site := range sites {
			participants[site.Caller], participants[site.Callee] = true, true
			if !needed[site.Callee] {
				needed[site.Callee], queue = true, append(queue, site.Callee)
			}
		}
	}
	if len(report.Sites) == 0 {
		return needed, nil, nil
	}
	var err error
	report.Activities, err = c.identities(participants)
	return needed, report, err
}

func (c workspaceCalls) identities(keys map[string]bool) ([]WorkspaceCallActivity, error) {
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	var result []WorkspaceCallActivity
	for _, key := range ordered {
		d := c[key]
		original, err := workspaceActivityIdentity(d.namespace, d.ref.Activity)
		if err != nil {
			return nil, err
		}
		lowered, err := workspaceActivityIdentity("gooo_workspace", d.ref.LoweredName)
		if err != nil {
			return nil, err
		}
		result = append(result, WorkspaceCallActivity{d.ref, original, lowered, d.source, sourceSHA256([]byte(d.body))})
	}
	return result, nil
}

func workspaceActivityIdentity(namespace, name string) (string, error) {
	model, err := bidir.Get(bidir.Document{Namespace: namespace,
		Declarations: []bidir.Declaration{{Kind: bidir.ActivityKind, Name: name}}})
	if err != nil {
		return "", err
	}
	return string(model.Nodes[0].ID), nil
}
