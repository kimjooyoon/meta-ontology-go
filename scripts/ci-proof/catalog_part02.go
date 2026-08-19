package main

import (
	"fmt"
	"strconv"
	"strings"
)

func validateFailureCatalog() error {
	if failureCatalogDigestErr != nil {
		return failureCatalogDigestErr
	}
	data, err := readFailureFile(failureCatalogPath)
	if err != nil {
		return fmt.Errorf("read failure catalog: %w", err)
	}
	if want := "sha256:" + digestBytes(data); want != failureCatalogDigest {
		return fmt.Errorf("failure catalog digest changed while validating")
	}
	entries, err := parseCatalogDocument(string(data))
	if err != nil {
		return err
	}
	if len(entries) != len(failureCatalogRecords) {
		return fmt.Errorf("failure catalog metadata count mismatch")
	}
	for _, record := range failureCatalogRecords {
		entry, ok := entries[record.Code]
		if !ok || entry.Class != record.Entry.Class || entry.Severity != record.Entry.Severity || entry.BlockingScope != record.Entry.BlockingScope || entry.Parallelizable != record.Entry.Parallelizable || entry.HandoffRequired != record.Entry.HandoffRequired || entry.Owner != record.Entry.Owner {
			return fmt.Errorf("failure catalog metadata mismatch for %s", record.Code)
		}
	}
	return nil
}
func parseCatalogDocument(document string) (map[string]catalogDocumentEntry, error) {
	const prefix = "<!-- machine-catalog: "
	const suffix = " -->"
	entries := make(map[string]catalogDocumentEntry)
	for _, line := range strings.Split(document, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		if !strings.HasSuffix(line, suffix) {
			return nil, fmt.Errorf("malformed failure catalog metadata")
		}
		fields := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix), "|")
		if len(fields) != 7 || fields[0] == "" {
			return nil, fmt.Errorf("failure catalog metadata must contain seven fields")
		}
		parallelizable, err := strconv.ParseBool(fields[4])
		if err != nil {
			return nil, fmt.Errorf("invalid catalog parallelization for %s", fields[0])
		}
		handoffRequired, err := strconv.ParseBool(fields[5])
		if err != nil {
			return nil, fmt.Errorf("invalid catalog handoff flag for %s", fields[0])
		}
		if _, exists := entries[fields[0]]; exists {
			return nil, fmt.Errorf("duplicate failure catalog metadata for %s", fields[0])
		}
		entries[fields[0]] = catalogDocumentEntry{Code: fields[0], Class: fields[1], Severity: fields[2], BlockingScope: fields[3], Parallelizable: parallelizable, HandoffRequired: handoffRequired, Owner: fields[6]}
	}
	return entries, nil
}
func validateFailureOwnerRegistry(branch string) error {
	data, err := readFailureFile(failureOwnerRegistryPath)
	if err != nil {
		return fmt.Errorf("read failure owner registry: %w", err)
	}
	return validateFailureOwnerRegistryDocument(data, branch)
}
