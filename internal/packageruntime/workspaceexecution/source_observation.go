package workspaceexecution

import "fmt"

// The oracle is a source dependency even when it does not occur in computes.
// Resolve it in the declaring package and retain it in the same helper closure.
func (c workspaceCalls) rewriteObservationReference(key string, d workspaceCallDeclaration) ([]WorkspaceCallSite, error) {
	o := d.activity.Assembly.Spec.Observation
	if o == nil || o.OracleActivity == "" {
		return nil, nil
	}
	_, sites, err := c.rewriteCalls(key, d, o.OracleActivity+"()", "observe:oracle", false)
	if err != nil {
		return nil, fmt.Errorf("resolve observe oracle: %w", err)
	}
	if len(sites) != 1 || sites[0].Callee == key {
		return nil, fmt.Errorf("observe oracle requires a separate declared activity")
	}
	o.OracleActivity = c[sites[0].Callee].ref.LoweredName
	return sites, nil
}
