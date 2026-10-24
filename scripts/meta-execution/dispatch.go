package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"

type nativeExecutor func(string, string, string, generation.Plan, generation.Action, metaExecutionTrace) (
	operationMaterialization, *operationError)

var nativeExecutors = map[generation.ExecutionRoute]nativeExecutor{
	generation.RouteRegisterSyntax:  executeRegistrationRoute,
	generation.RouteCollapseAssign:  executeCollapseRoute,
	generation.RouteSplitGo:         executeSplitRoute,
	generation.RouteExtractFunction: executeExtractRoute,
}

func executeAction(workspace, gitDir, metricsPath string, plan generation.Plan,
	action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	execute, exists := nativeExecutors[generation.NativeExecutionRoute(action.Operation)]
	if !exists {
		return operationMaterialization{}, newOperationError("execute-operation", "select-executor",
			"UNSUPPORTED_SELECTED_OPERATION", "KNOWN_CONTRADICTION", "report-counterexample")
	}
	return execute(workspace, gitDir, metricsPath, plan, action, trace)
}

func executeRegistrationRoute(workspace, _, _ string, plan generation.Plan,
	action generation.Action, _ metaExecutionTrace) (operationMaterialization, *operationError) {
	return executeNativeRegistration(workspace, plan, action)
}

func executeCollapseRoute(workspace, gitDir, metricsPath string, plan generation.Plan,
	action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	if failure := validateCollapseAction(action); failure != nil {
		return operationMaterialization{}, failure
	}
	return executeCollapse(workspace, gitDir, metricsPath, plan, action, trace)
}

func executeSplitRoute(workspace, gitDir, metricsPath string, plan generation.Plan,
	action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	if failure := checkSourcePrerequisite(workspace, metricsPath, plan, action); failure != nil {
		return operationMaterialization{}, failure
	}
	return executeSplit(workspace, gitDir, metricsPath, plan, action, trace)
}

func executeExtractRoute(workspace, gitDir, metricsPath string, plan generation.Plan,
	action generation.Action, trace metaExecutionTrace) (operationMaterialization, *operationError) {
	if failure := checkSourcePrerequisite(workspace, metricsPath, plan, action); failure != nil {
		return operationMaterialization{}, failure
	}
	return executeExtract(workspace, gitDir, metricsPath, plan, action, trace)
}
