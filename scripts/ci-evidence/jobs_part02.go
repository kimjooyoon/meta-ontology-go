package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readMetadata() (metadata, error) {
	values := map[string]string{
		"repository":   os.Getenv("CI_REPOSITORY"),
		"event":        os.Getenv("CI_EVENT"),
		"event_ref":    os.Getenv("CI_EVENT_REF"),
		"checkout_ref": os.Getenv("CI_CHECKOUT_REF"),
		"base_ref":     os.Getenv("CI_BASE_REF"),
		"base_sha":     os.Getenv("CI_BASE_SHA"),
		"head_sha":     os.Getenv("CI_HEAD_SHA"),
		"workflow_sha": os.Getenv("CI_WORKFLOW_SHA"),
	}
	for name, value := range values {
		if strings.TrimSpace(value) == "" {
			return metadata{}, fmt.Errorf("missing CI evidence field %s", name)
		}
	}
	runID, err := positiveInt("CI_RUN_ID")
	if err != nil {
		return metadata{}, err
	}
	attempt, err := positiveInt("CI_RUN_ATTEMPT")
	if err != nil {
		return metadata{}, err
	}
	if !validSHA(values["base_sha"]) || !validSHA(values["head_sha"]) || !validSHA(values["workflow_sha"]) {
		return metadata{}, fmt.Errorf("CI evidence revisions must be 40-character SHA-1 values")
	}
	if !validEventRef(values["event"], values["event_ref"]) || values["checkout_ref"] != values["head_sha"] {
		return metadata{}, fmt.Errorf("CI evidence refs are missing or mismatched")
	}
	toolchain, err := toolchainIdentity()
	if err != nil {
		return metadata{}, err
	}
	return metadata{Repository: values["repository"], Event: values["event"], EventRef: values["event_ref"], CheckoutRef: values["checkout_ref"], BaseRef: values["base_ref"], BaseSHA: values["base_sha"], HeadSHA: values["head_sha"], RunID: runID, RunAttempt: attempt, WorkflowSHA: values["workflow_sha"], Toolchain: toolchain}, nil
}
func positiveInt(name string) (int64, error) {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("missing or invalid CI evidence field %s", name)
	}
	return value, nil
}
func validSHA(value string) bool {
	if len(value) != 40 || strings.Trim(value, "0") == "" {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
