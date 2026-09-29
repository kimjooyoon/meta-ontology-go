package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

const decideUsage = "usage: gooo decide [--json] <request.json>"

func runDecide(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	if len(args) != 1 {
		return reportUsage(jsonMode, stdout, stderr, "decide", decideUsage)
	}
	raw, err := reader.ReadFile(args[0])
	if err != nil {
		return decideFailure(jsonMode, stdout, stderr, err)
	}
	var request decisionroute.Request
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return decideFailure(jsonMode, stdout, stderr, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return decideFailure(jsonMode, stdout, stderr, fmt.Errorf("request has trailing JSON content"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	receipt, err := decisionroute.Resolve(ctx, request, os.Getenv("GOOO_LAYA_URL"), os.Getenv("GOOO_LAYA_API_KEY"))
	if err != nil {
		return decideFailure(jsonMode, stdout, stderr, err)
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(receipt); err != nil {
			fmt.Fprintf(stderr, "gooo: decide: write receipt: %v\n", err)
			return exitFailure
		}
	} else {
		fmt.Fprintf(stdout, "selected=%s mode=%s provider=%s", receipt.Selected, receipt.Mode, receipt.Provider)
		if receipt.FallbackReason != "" {
			fmt.Fprintf(stdout, " fallback_reason=%s", receipt.FallbackReason)
		}
		fmt.Fprintln(stdout)
	}
	return exitOK
}

func decideFailure(jsonMode bool, stdout, stderr io.Writer, err error) int {
	if jsonMode {
		_ = json.NewEncoder(stdout).Encode(map[string]string{
			"schema": decisionroute.ReceiptSchema, "mode": "rejected", "reason": "REQUEST_INVALID",
		})
	} else {
		fmt.Fprintf(stderr, "gooo: decide: %v\n", err)
	}
	return exitFailure
}
