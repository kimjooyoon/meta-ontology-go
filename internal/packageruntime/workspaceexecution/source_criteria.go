package workspaceexecution

import "fmt"

// Predicates are dependencies even when the assembling computes body does not
// call them. Preserve the package-resolved activity in the source carrier.
func (c workspaceCalls) rewriteCriteriaReferences(key string, d workspaceCallDeclaration) ([]WorkspaceCallSite, error) {
	criteria := d.activity.Assembly.Spec.Criteria
	if criteria == nil {
		return nil, nil
	}
	var references []WorkspaceCallSite
	for i := range criteria.Predicates {
		predicate := &criteria.Predicates[i]
		_, sites, err := c.rewriteCalls(key, d, predicate.Activity+"()", "criteria:"+predicate.ID, false)
		if err != nil {
			return nil, fmt.Errorf("resolve criterion %q: %w", predicate.ID, err)
		}
		if len(sites) != 1 || sites[0].Callee == key {
			return nil, fmt.Errorf("criterion %q requires a separate declared activity", predicate.ID)
		}
		predicate.Activity = c[sites[0].Callee].ref.LoweredName
		references = append(references, sites...)
	}
	return references, nil
}
