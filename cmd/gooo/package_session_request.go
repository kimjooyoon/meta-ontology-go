package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func decodePackageSessionRequest(raw []byte) (packageSessionRequest, error) {
	var request packageSessionRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return request, fmt.Errorf("session request must contain one JSON value")
	}
	if request.Schema != packageSessionRequestSchema || request.ID == "" || len(request.ID) > 128 || len(request.Arguments) > 48 {
		return request, fmt.Errorf("session request schema, id or argument bound differs")
	}
	return request, nil
}

func handlePackageSessionLine(ctx context.Context, raw []byte, reader SourceReader,
	owner *workspaceexecution.Executor, stderr io.Writer) packageSessionResponse {
	request, err := decodePackageSessionRequest(raw)
	response := packageSessionResponse{Schema: packageSessionResponseSchema, ID: request.ID, ExitCode: exitUsage}
	if err != nil {
		response.Error = err.Error()
		return response
	}
	if request.Operation == "close" {
		if len(request.Arguments) != 0 {
			response.Error = "close does not accept arguments"
			return response
		}
		response.Closed, response.ExitCode = true, exitOK
		if err := owner.Close(); err != nil {
			response.ExitCode, response.Error = exitFailure, err.Error()
		}
		return response
	}
	return executePackageSessionRequest(ctx, request, reader, owner, stderr)
}

func executePackageSessionRequest(ctx context.Context, request packageSessionRequest, reader SourceReader,
	owner *workspaceexecution.Executor, stderr io.Writer) packageSessionResponse {
	response := packageSessionResponse{Schema: packageSessionResponseSchema, ID: request.ID, ExitCode: exitUsage}
	ctx, args, timing := packageTimingContext(ctx, request.Arguments)
	args, _ = parseJSONFlag(args)
	flags, manifest, err := parsePackageSessionArguments(request.Operation, args)
	if err != nil {
		response.Error = err.Error()
		return response
	}
	var receipt packageExecutionReceipt
	if request.Operation == "execute" {
		receipt, err = executePackageWithExecutor(ctx, reader, manifest, flags, owner)
	} else {
		receipt, err = executePackageSavedWithExecutor(ctx, reader, manifest, flags, false, owner)
	}
	response.Receipt, response.ExitCode = &receipt, exitOK
	if err != nil {
		markPackageExecutionError(&receipt, err)
		response.ExitCode = exitFailure
	}
	writePackageTiming(stderr, timing, request.Operation, receipt)
	return response
}

func parsePackageSessionArguments(operation string, args []string) (map[string]string, string, error) {
	var flags map[string]string
	var manifest string
	var err error
	switch operation {
	case "execute":
		flags, manifest, err = parsePackageExecuteArgs(args)
	case "replay":
		flags, manifest, err = parsePackageReplayArgs(args)
	default:
		return nil, "", fmt.Errorf("session operation must be execute, replay or close")
	}
	if err == nil && flags["--receipt-compression"] != "" {
		err = fmt.Errorf("session framing returns JSON receipts; omit --receipt-compression")
	}
	return flags, manifest, err
}
