package extractor

import "testing"

func TestCompositionExecutorTimingRenderedCapacity(t *testing.T) {
	// Find declarations by package and receiver so the same check also reads
	// the final CI workspace, where complete declarations occupy new files.
	requireNamedCapacities(t, "../../../..", []string{
		"internal/bodyexecution#Executor.ExecuteComposition",
		"internal/bodyexecution#Executor.executeWithJointCalls",
		"internal/bodyexecution#compositionArtifactKey",
		"internal/bodyexecution#compositionExecutableFor",
		"internal/bodyexecution#timedCompositionLocalEvidence",
		"internal/bodyexecution#bindCompositionBuild",
	})
}
