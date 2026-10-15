package workspaceexecution

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
	"strings"
)

type valueCallFrame struct {
	caller, callee, call string
	arguments            []string
}
type valueVisit struct {
	activity, id, via string
	frames            []valueCallFrame
	previous          int
}
type valueTransferWalker struct {
	ctx             context.Context
	result          Result
	query           OriginQuery
	catalog         valueTransferCatalog
	out             *ProgramValueOrigins
	issues, emitted map[string]bool
	accesses        map[string]bool
	visits, steps   int
}

func (w *valueTransferWalker) issue(activity, reason string) {
	key := activity + "\x00" + reason
	if !w.issues[key] {
		w.issues[key] = true
		w.out.Issues = append(w.out.Issues, OriginQueryIssue{ActivityID: activity, Reason: reason})
	}
}

func valueVisitKey(v valueVisit) string {
	var key strings.Builder
	key.WriteString(v.activity + "\x00" + v.id)
	for _, frame := range v.frames {
		key.WriteString("\x00" + frame.caller + ":" + frame.call)
	}
	return key.String()
}

func (w *valueTransferWalker) walk(root *valueTransferActivity, rootID string) error {
	first := valueVisit{activity: root.id, id: rootID, via: "root", previous: -1}
	queue := []valueVisit{first}
	seen := map[string]bool{valueVisitKey(first): true}
	appendVisit := func(v valueVisit) {
		key := valueVisitKey(v)
		if seen[key] {
			return
		}
		seen[key] = true
		if len(queue) >= 8192 {
			w.out.Truncated = true
			return
		}
		queue = append(queue, v)
	}
	for i := 0; i < len(queue); i++ {
		if err := w.ctx.Err(); err != nil {
			return err
		}
		w.visits++
		if w.visits > 8192 {
			w.out.Truncated = true
			return nil
		}
		if err := w.walkValueVisit(root, rootID, queue, i, appendVisit); err != nil {
			return err
		}
		if w.out.Truncated {
			return nil
		}
	}
	return nil
}

func (w *valueTransferWalker) walkValueVisit(root *valueTransferActivity, rootID string, queue []valueVisit, i int, appendVisit func(valueVisit)) error {
	v := queue[i]
	a := w.catalog.byID[v.activity]
	node, ok := a.nodes[v.id]
	if !ok {
		w.issue(a.id, "VALUE_NODE_UNAVAILABLE")
		return nil
	}
	w.recordAccess(rootID, a, node)
	for _, reason := range a.graph.Reasons {
		if reason != "CALLEE_BODY_NOT_EXPANDED" {
			w.issue(a.id, reason)
		}
	}
	if err := w.emitValueCandidates(root, rootID, a, node, queue, i); err != nil {
		return err
	}
	crossed, skip := w.followValueCall(a, node, v, i, appendVisit)
	if skip {
		return nil
	}
	if node.Kind == "input_reference" {
		crossed = w.followValueInput(a, node, v, i, appendVisit)
	}
	if !crossed {
		for _, id := range node.Inputs {
			appendVisit(valueVisit{a.id, id, "input", v.frames, i})
		}
		if len(node.Inputs) == 0 && w.query.CandidateID == "" {
			kind := "source_expression"
			if node.Kind == "input_reference" {
				kind = "external_input"
			}
			if node.Kind == "unresolved_reference" {
				kind = "unresolved"
			}
			if err := w.emit(root, rootID, kind, "", "", valueActivityLocations(w.result, a.id), queue, i); err != nil {
				return err
			}
		}
	}
	for _, id := range node.Controls {
		appendVisit(valueVisit{a.id, id, "control", v.frames, i})
	}
	return nil
}

func (w *valueTransferWalker) emitValueCandidates(root *valueTransferActivity, rootID string, a *valueTransferActivity, node bodycodegen.BodyValueOrigin, queue []valueVisit, i int) error {
	for _, origin := range a.origins {
		if w.query.CandidateID != "" && origin.CandidateID != w.query.CandidateID {
			continue
		}
		for _, fragment := range origin.Fragments {
			span := fragment.SelectedDecoded
			if fragment.HoleID == "" || node.Decoded.Start >= span.End || span.Start >= node.Decoded.End {
				continue
			}
			insideArgument := false
			for callID, call := range a.calls {
				callSpan := a.nodes[callID].Decoded
				if w.catalog.byName[call.Callee] != nil && callSpan.Start >= node.Decoded.Start && callSpan.End <= node.Decoded.End && span.Start > callSpan.Start && span.End < callSpan.End {
					insideArgument = true
				}
			}
			if insideArgument {
				continue
			}
			if err := w.emit(root, rootID, "candidate", origin.CandidateID, fragment.HoleID, originUseLocations(w.result, *origin, fragment.HoleID), queue, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *valueTransferWalker) followValueCall(a *valueTransferActivity, node bodycodegen.BodyValueOrigin, v valueVisit, i int, appendVisit func(valueVisit)) (bool, bool) {
	if node.Kind == "call_result" {
		call, found := a.calls[node.ID]
		callee := w.catalog.byName[call.Callee]
		if found && callee != nil && len(call.Arguments) == callee.arity {
			if len(v.frames) >= 16 {
				w.issue(a.id, "CALL_DEPTH_LIMIT")
				return true, true
			}
			recursive := callee.id == a.id
			for _, frame := range v.frames {
				recursive = recursive || frame.caller == callee.id
			}
			if recursive {
				w.issue(a.id, "RECURSIVE_VALUE_PATH")
				return true, true
			}
			frames := append(append([]valueCallFrame(nil), v.frames...), valueCallFrame{a.id, callee.id, call.NodeID, call.Arguments})
			returns := 0
			for _, id := range callee.graph.Roots {
				if callee.nodes[id].Kind == "return" {
					appendVisit(valueVisit{callee.id, id, "call_return", frames, i})
					returns++
				}
			}
			if returns == 0 {
				w.issue(callee.id, "CALLEE_RETURN_UNAVAILABLE")
			}
			return true, false
		} else {
			w.issue(a.id, "CALLEE_BODY_NOT_EXPANDED")
		}
	}
	return false, false
}

func (w *valueTransferWalker) followValueInput(a *valueTransferActivity, node bodycodegen.BodyValueOrigin, v valueVisit, i int, appendVisit func(valueVisit)) bool {
	if node.Kind == "input_reference" {
		port, index := valueInputPort(node.Name, a.arity, a.inputNames...)
		if len(v.frames) > 0 {
			frame := v.frames[len(v.frames)-1]
			if frame.callee != a.id || index < 0 || index >= len(frame.arguments) {
				w.issue(a.id, "CALL_ARGUMENT_UNAVAILABLE")
			} else {
				appendVisit(valueVisit{frame.caller, frame.arguments[index], "call_argument", v.frames[:len(v.frames)-1], i})
				return true
			}
		} else if producer := w.catalog.byID[w.catalog.binds[a.id+"\x00"+port]]; producer != nil {
			for _, id := range producer.graph.Roots {
				if producer.nodes[id].Kind == "return" {
					appendVisit(valueVisit{producer.id, id, "bind", nil, i})
				}
			}
			return true
		}
	}
	return false
}

func valueInputPort(name string, arity int, declared ...string) (string, int) {
	names, err := inputbinding.Resolve(arity, declared)
	if err == nil {
		for i, binding := range names {
			if name == binding {
				return inputbinding.PositionName(arity, i), i
			}
		}
	}
	return name, -1
}

func (w *valueTransferWalker) emit(root *valueTransferActivity, rootID, kind, candidate, hole string, locations []OriginUseLocation, queue []valueVisit, index int) error {
	v := queue[index]
	key := root.id + "\x00" + rootID + "\x00" + v.activity + "\x00" + kind + "\x00" + candidate + "\x00" + hole
	if kind != "candidate" {
		key += "\x00" + v.id
	}
	if w.emitted[key] {
		return nil
	}
	w.emitted[key] = true
	if len(w.out.Rows) >= w.query.Limit {
		w.out.Truncated = true
		return nil
	}
	var path []ProgramOriginStep
	for current := index; current >= 0; current = queue[current].previous {
		v := queue[current]
		a := w.catalog.byID[v.activity]
		n := a.nodes[v.id]
		span := n.Decoded
		if span.Start < 0 || span.End < span.Start || span.End > len(a.body) {
			return fmt.Errorf("value origin is outside selected body")
		}
		snippet, cut := originQuerySnippet(a.body[span.Start:span.End])
		path = append(path, ProgramOriginStep{a.id, a.graph.BodySHA256, n.ID, n.Kind, v.via, snippet, span.Start, span.End, cut})
	}
	if w.steps+len(path) > 2048 {
		w.out.Truncated = true
		return nil
	}
	w.steps += len(path)
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	w.out.Rows = append(w.out.Rows, ProgramValueOrigin{RootID: rootID, RootKind: root.nodes[rootID].Kind, Kind: kind, ActivityID: v.activity, CandidateID: candidate, HoleID: hole, Path: path, Locations: locations})
	return nil
}

func valueActivityLocations(result Result, activity string) []OriginUseLocation {
	for _, origin := range result.Program.ActivityOrigins {
		input := origin.Input
		if origin.ActivityID != activity || input.Declaration == nil || input.Declaration.Status != "MAPPED" {
			continue
		}
		d := input.Declaration
		return []OriginUseLocation{{View: "workspace_input_declaration", PackagePath: input.PackagePath, Filename: input.Filename, SourceSHA256: input.SourceSHA256, ActivityID: input.ActivityID, Start: d.Start, End: d.End}}
	}
	return nil
}
