package sourceimpact

import (
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type bodySurface struct {
	label, body string
	enclosing   []string
}

func inspectActivityReferences(s *snapshot, node semantic.Node, names map[string]string, inputCount int, revision string) {
	if node.ValueProgram == "" {
		s.addUnknown(revision, string(node.ID), "computes", "", "source body unavailable; implementation calls were not inspected")
	}
	inspectAssemblyDependencies(s, node, names, revision)
	inputNames := make([]string, len(node.InputNames))
	for i, name := range node.InputNames {
		inputNames[i] = name.Name
	}
	for _, surface := range activityBodySurfaces(node) {
		inspectBodySurface(s, string(node.ID), surface, names, inputCount, revision, inputNames)
	}
}

func inspectBodySurface(s *snapshot, id string, surface bodySurface, names map[string]string, inputCount int, revision string, inputNames []string) {
	if surface.body == "" {
		return
	}
	refs, err := bodycodegen.InspectNamedBodySourceReferences(surface.body, inputCount, inputNames, surface.enclosing...)
	if err != nil {
		s.addUnknown(revision, id, surface.label, "", "body reference syntax unavailable: "+err.Error())
		return
	}
	for _, ref := range refs {
		label := surface.label + ":" + strconv.Itoa(ref.Ordinal)
		if ref.AmbiguousScope {
			s.addUnknown(revision, id, label, ref.Name, "enclosing binding may shadow this fragment call; exact candidate scope is unresolved")
			continue
		}
		if !ref.Direct {
			s.addUnknown(revision, id, label, ref.Name, "indirect or external call target is unresolved")
			continue
		}
		if target, ok := names[ref.Name]; ok {
			s.addEdge(target, id, "body_call", label)
		} else if !bodycodegen.IsBodyPrimitiveName(ref.Name) {
			s.addUnknown(revision, id, label, ref.Name, "undeclared call target")
		}
	}
}

func (s *snapshot) addUnknown(revision, id, surface, reference, reason string) {
	s.unknowns = append(s.unknowns, Unknown{Revision: revision, ActivityID: id, Surface: surface, Reference: reference, Reason: reason})
}
