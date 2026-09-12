package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

// Constructed only after report validation; never used for malformed evidence.
type validatedNegativeReportError struct {
	decision string
	reason   string
}

func (err validatedNegativeReportError) Error() string {
	return fmt.Sprintf("%s: %s", err.decision, err.reason)
}

func persistBuiltReport(path string, report languagesyntax.Report, buildErr error, stdout io.Writer) error {
	if buildErr != nil {
		if _, validated := buildErr.(validatedNegativeReportError); !validated {
			return buildErr
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return err
	}
	if buildErr != nil {
		return buildErr
	}
	printSummary(stdout, report)
	return nil
}
