package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

// BodyValueOrigins describes bounded static may-dependence. Branch alternatives
// are retained together; it does not claim which assignment ran for an input.
type BodyValueOrigins struct {
	Schema     string            `json:"schema"`
	Status     string            `json:"status"`
	Reasons    []string          `json:"reasons,omitempty"`
	BodySHA256 string            `json:"body_sha256"`
	Nodes      []BodyValueOrigin `json:"nodes"`
	Roots      []string          `json:"roots"`
}

type BodyValueOrigin struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`
	Name     string          `json:"name,omitempty"`
	Decoded  FillOriginRange `json:"decoded"`
	Inputs   []string        `json:"inputs,omitempty"`
	Controls []string        `json:"controls,omitempty"`
}

type valueOriginEnv map[*ast.Object][]string

type valueOriginReader struct {
	set            *token.FileSet
	header, length int
	nodes          map[string]*BodyValueOrigin
	reasons        map[string]bool
	roots          map[string]bool
	calls          map[string]BodyValueCall
	steps          int
	mapping        bodySurfaceMap
}

func InspectBodyValueOrigins(body string, arity int, declared ...string) BodyValueOrigins {
	out, _ := inspectBodyValueOrigins(body, arity, false, declared...)
	return out
}

func inspectBodyValueOrigins(body string, arity int, captureCalls bool, declared ...string) (BodyValueOrigins, []BodyValueCall) {
	out := BodyValueOrigins{Schema: "gooo/body-value-origins/v1", Status: "UNMAPPED"}
	if len(body) > 128<<10 {
		out.Reasons = []string{"BODY_SIZE_LIMIT"}
		return out, nil
	}
	out.BodySHA256, _ = NormalizedConditionBodySHA256(body)
	file, set, header, err := parseRenameBody(body, arity, declared...)
	if err != nil {
		out.Reasons = []string{"BODY_SYNTAX_UNAVAILABLE"}
		return out, nil
	}
	_, mapping, _ := rewriteBodySurface(body)
	r := valueOriginReader{mapping: mapping, set: set, header: header, length: len(body), nodes: map[string]*BodyValueOrigin{}, reasons: map[string]bool{}, roots: map[string]bool{}}
	if captureCalls {
		r.calls = map[string]BodyValueCall{}
	}
	var function *ast.FuncLit
	ast.Inspect(file, func(n ast.Node) bool {
		if f, ok := n.(*ast.FuncLit); ok && function == nil {
			function = f
			return false
		}
		return true
	})
	if function == nil {
		out.Reasons = []string{"BODY_FUNCTION_UNAVAILABLE"}
		return out, nil
	}
	r.block(function.Body, valueOriginEnv{}, nil)
	out.Status = "MAPPED"
	for reason := range r.reasons {
		out.Reasons = append(out.Reasons, reason)
	}
	if len(out.Reasons) > 0 {
		out.Status = "PARTIAL"
		sort.Strings(out.Reasons)
	}
	for _, node := range r.nodes {
		out.Nodes = append(out.Nodes, *node)
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].ID < out.Nodes[j].ID })
	for id := range r.roots {
		out.Roots = append(out.Roots, id)
	}
	sort.Strings(out.Roots)
	var calls []BodyValueCall
	for _, call := range r.calls {
		calls = append(calls, call)
	}
	sort.Slice(calls, func(i, j int) bool { return calls[i].NodeID < calls[j].NodeID })
	return out, calls
}

func (r *valueOriginReader) node(kind, name string, at ast.Node, inputs, controls []string) string {
	span := r.mapping.originalRange(FillOriginRange{r.set.Position(at.Pos()).Offset - r.header, r.set.Position(at.End()).Offset - r.header})
	if span.Start < 0 || span.End > r.length || span.End < span.Start {
		r.reasons["SOURCE_RANGE_UNAVAILABLE"] = true
		return ""
	}
	id := fmt.Sprintf("%s:%d:%d:%s", kind, span.Start, span.End, name)
	node := r.nodes[id]
	if node == nil {
		if len(r.nodes) >= 256 {
			r.reasons["NODE_LIMIT"] = true
			return ""
		}
		node = &BodyValueOrigin{ID: id, Kind: kind, Name: name, Decoded: span}
		r.nodes[id] = node
	}
	node.Inputs = joinValueOriginIDs(node.Inputs, inputs)
	node.Controls = joinValueOriginIDs(node.Controls, controls)
	return id
}

func joinValueOriginIDs(groups ...[]string) []string {
	seen := map[string]bool{}
	for _, group := range groups {
		for _, id := range group {
			if id != "" {
				seen[id] = true
			}
		}
	}
	var result []string
	for id := range seen {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func copyValueOriginEnv(env valueOriginEnv) valueOriginEnv {
	copy := valueOriginEnv{}
	for variable, ids := range env {
		copy[variable] = append([]string(nil), ids...)
	}
	return copy
}

func joinValueOriginEnvs(a, b valueOriginEnv) valueOriginEnv {
	out := copyValueOriginEnv(a)
	for variable, ids := range b {
		out[variable] = joinValueOriginIDs(out[variable], ids)
	}
	return out
}

func (r *valueOriginReader) step() bool {
	if len(r.nodes) >= 256 {
		r.reasons["NODE_LIMIT"] = true
		return false
	}
	r.steps++
	if r.steps > 8192 {
		r.reasons["WORK_LIMIT"] = true
		return false
	}
	return true
}
