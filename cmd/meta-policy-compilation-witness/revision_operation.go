package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

var revisionOperationPath = flag.String("revision-operation", "", "bind revision observation to an exact Gooo native operation contract")

func observeBoundRevision(ctx context.Context, policyPath string, source, requestBytes []byte, profilePackage, profileNamespace string, output io.Writer) error {
	operationSource, err := os.ReadFile(*revisionOperationPath)
	if err != nil {
		return fmt.Errorf("read Gooo revision operation: %w", err)
	}
	report, observationError := policycompilation.ObserveGoooPolicyDecisionRevision(
		ctx, operationSource, policyPath, source, profilePackage, profileNamespace, requestBytes,
	)
	if report.Schema == "" {
		return observationError
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return errors.Join(observationError, fmt.Errorf("write Gooo revision operation: %w", err))
	}
	return observationError
}
