package coupling

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

func normalizeSemantic(ir SemanticIR) (normalizedSemantic, error) {
	facts := make([]string, 0, len(ir.Nodes)+len(ir.Relations))
	nodes := make(map[string]struct{}, len(ir.Nodes))
	for _, node := range ir.Nodes {
		if !validID(node.ID) || !validToken(node.Kind) || !validToken(node.Namespace) {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic node")
		}
		if node.Kind != semantic.Entity.String() && node.Kind != semantic.Activity.String() && node.Kind != semantic.Agent.String() {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic kind")
		}
		if _, duplicate := nodes[node.ID]; duplicate {
			return normalizedSemantic{}, fmt.Errorf("duplicate semantic node")
		}
		nodes[node.ID] = struct{}{}
		facts = append(facts, "node\t"+node.ID+"\t"+node.Kind+"\t"+node.Namespace)
	}
	for _, relation := range ir.Relations {
		if !validID(relation.Subject) || !validID(relation.Object) || !validToken(relation.Predicate) {
			return normalizedSemantic{}, fmt.Errorf("invalid semantic relation")
		}
		if _, ok := nodes[relation.Subject]; !ok {
			return normalizedSemantic{}, fmt.Errorf("relation subject is not registered")
		}
		if _, ok := nodes[relation.Object]; !ok {
			return normalizedSemantic{}, fmt.Errorf("relation object is not registered")
		}
		facts = append(facts, "relation\t"+relation.Subject+"\t"+relation.Predicate+"\t"+relation.Object)
	}
	sort.Strings(facts)
	seen := make(map[string]struct{}, len(facts))
	for _, fact := range facts {
		if _, duplicate := seen[fact]; duplicate {
			return normalizedSemantic{}, fmt.Errorf("duplicate semantic fact")
		}
		seen[fact] = struct{}{}
	}
	canonical := strings.Join(append([]string{"semantic-ir-v1"}, facts...), "\n") + "\n"
	return normalizedSemantic{digest: digestBytes([]byte(canonical)), facts: facts}, nil
}
