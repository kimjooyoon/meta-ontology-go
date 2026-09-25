package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
