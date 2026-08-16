package selectiveci

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateChains(input Input) issueState {
	state := issueState{}
	edges := indexEdges(input.InferencePath.Edges)
	used := make(map[semantic.ID]struct{})
	for _, path := range input.Paths {
		pathState, pathUsed := validateChainPath(path, edges)
		state.merge(pathState)
		for recordID := range pathUsed {
			used[recordID] = struct{}{}
		}
	}
	for _, edge := range input.InferencePath.Edges {
		if _, exists := used[edge.RecordID]; !exists {
			state.add(issueFailClosed, CodeDisconnected)
		}
	}
	return state
}

func indexEdges(edges []semantic.InferenceEdge) map[semantic.ID]semantic.InferenceEdge {
	indexed := make(map[semantic.ID]semantic.InferenceEdge, len(edges))
	for _, edge := range edges {
		indexed[edge.RecordID] = edge
	}
	return indexed
}

func (s *issueState) merge(other issueState) {
	if other.class != issueNone {
		s.add(other.class, other.code)
	}
}

func validateChainPath(path Path, edges map[semantic.ID]semantic.InferenceEdge) (issueState, map[semantic.ID]struct{}) {
	state := issueState{}
	used := make(map[semantic.ID]struct{})
	selected := make([]semantic.InferenceEdge, 0, len(path.RecordIDs))
	for index, recordID := range path.RecordIDs {
		edge, exists := edges[recordID]
		if !exists {
			state.add(issueUnknown, CodeMissing)
			continue
		}
		used[recordID] = struct{}{}
		if edge.Kind != path.ExpectedKinds[index] {
			state.add(issueFailClosed, CodeMalformed)
		}
		if edge.Kind == semantic.InferenceObservationCandidate {
			state.add(issueUnknown, CodeCandidate)
		}
		selected = append(selected, edge)
	}
	if state.class != issueNone && len(selected) != len(path.RecordIDs) {
		return state, used
	}
	chain, err := semantic.NewInferencePathChain(selected...)
	if err != nil {
		state.add(issueFailClosed, chainCode(err))
		return state, used
	}
	if !chainEndpointsMatch(chain, path) {
		state.add(issueFailClosed, CodeWrongEndpoint)
	}
	if !chainAuthorityMatch(chain, path) {
		state.add(issueFailClosed, CodeWrongEndpoint)
	}
	return state, used
}

func chainCode(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "cycle"):
		return CodeCycle
	case strings.Contains(message, "ambiguity"):
		return CodeAmbiguous
	case strings.Contains(message, "disconnected"), strings.Contains(message, "orphan"):
		return CodeDisconnected
	default:
		return CodeMalformed
	}
}

func chainEndpointsMatch(chain semantic.InferencePathChain, path Path) bool {
	if len(chain.Edges) == 0 || chain.Edges[0].SubjectID != path.RootID {
		return false
	}
	return chain.Edges[len(chain.Edges)-1].ObjectID == path.ReceiptID
}

func chainAuthorityMatch(chain semantic.InferencePathChain, path Path) bool {
	first, last := chain.Edges[0], chain.Edges[len(chain.Edges)-1]
	if first.Kind != semantic.InferenceAuthoritativeDeclaration || !containsID(first.SourceRoots, path.RootID) {
		return false
	}
	if last.Kind != semantic.InferenceIndependentVerification || last.SubjectID != path.CommandID {
		return false
	}
	obligationIndex, commandIndex := -1, -1
	for index, edge := range chain.Edges {
		if edge.ObjectID == path.ObligationID && obligationIndex < 0 {
			obligationIndex = index
		}
		if edge.ObjectID == path.CommandID && obligationIndex >= 0 && commandIndex < 0 {
			commandIndex = index
		}
	}
	return obligationIndex >= 0 && commandIndex > obligationIndex
}
