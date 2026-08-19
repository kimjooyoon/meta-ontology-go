package adapter

import (
	"fmt"
	"path/filepath"
)

func validateVerifiedWorkflow(workflow WorkflowBinding) error {
	if workflow.Status == WorkflowEvidenceMissing {
		return oracleError(OracleNW001, "observer workflow evidence is missing")
	}
	if workflow.Status != WorkflowEvidenceVerified {
		return oracleError(OracleNW003, "observer workflow evidence is not independently verified")
	}
	if err := workflow.validate(); err != nil {
		return oracleError(OracleNW003, "observer workflow evidence: "+err.Error())
	}
	return nil
}
func requestObservationBinding(request Request) ObservationBinding {
	return ObservationBinding{Fixture: request.Fixture, Operation: request.Operation, RunID: request.RunID}
}
func validateObservation(observation NoWriteObservation) error {
	if observation.Reason != "" && !validRejectionKind(observation.Reason) {
		return oracleError(OracleNW003, "observer rejection kind is invalid")
	}
	if err := observation.Binding.validate(); err != nil {
		return oracleError(OracleNW002, "observation binding is stale or malformed")
	}
	if err := validatePaths(observation); err != nil {
		return oracleError(OracleNW003, err.Error())
	}
	if err := validateState(observation.Before, true); err != nil {
		return oracleError(OracleNW003, "before snapshot: "+err.Error())
	}
	if err := validateState(observation.After, true); err != nil {
		return oracleError(OracleNW003, "after snapshot: "+err.Error())
	}
	return nil
}
func validRejectionKind(reason RejectionKind) bool {
	return reason == RejectionCancelled || reason == RejectionClosed
}
func validatePaths(observation NoWriteObservation) error {
	paths := observation.Paths
	if !filepath.IsAbs(paths.SourcePath) || !filepath.IsAbs(paths.OutputPath) || !filepath.IsAbs(paths.TempRoot) {
		return fmt.Errorf("observer paths must be absolute")
	}
	if observation.Before.Source.Path != paths.SourcePath || observation.After.Source.Path != paths.SourcePath {
		return fmt.Errorf("source path is not bound to observer paths")
	}
	if observation.Before.Output.Path != paths.OutputPath || observation.After.Output.Path != paths.OutputPath {
		return fmt.Errorf("output path is not bound to observer paths")
	}
	return nil
}
func validateState(state FilesystemState, requirePrimary bool) error {
	if err := validateFileObservation(state.Source, requirePrimary); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := validateFileObservation(state.Output, requirePrimary); err != nil {
		return fmt.Errorf("output: %w", err)
	}
	if err := validateTempSnapshot(state.Temp); err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	return nil
}
