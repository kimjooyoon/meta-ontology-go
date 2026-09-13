package generation

import "path/filepath"

// ObservationBundlePath gives distinct manifest outputs distinct observation
// owners. The canonical manifest retains its established bundle location so
// existing artifact consumers keep reading the first execution, not its replay.
// Paths are storage ownership, not semantic evidence or attempt authorization.
func ObservationBundlePath(planPath, manifestPath string) string {
	canonical := filepath.Join(filepath.Dir(planPath), "self-improvement-execution.json")
	if filepath.Clean(manifestPath) == canonical {
		return filepath.Join(filepath.Dir(planPath), "meta-operation-observations.json")
	}
	return filepath.Clean(manifestPath) + ".observations.json"
}
