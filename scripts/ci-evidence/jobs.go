package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var canonicalJobs = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type apiJob struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

type jobEvidence struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

type digests struct {
	SourceSHA256           string `json:"source_sha256"`
	IRSHA256               string `json:"ir_sha256"`
	GeneratorFixtureSHA256 string `json:"generator_fixture_sha256"`
	GeneratedOutputSHA256  string `json:"generated_output_sha256"`
	SourceMapSHA256        string `json:"source_map_sha256"`
	PolicySHA256           string `json:"policy_sha256"`
	ToolchainSHA256        string `json:"toolchain_sha256"`
	BundleSHA256           string `json:"bundle_sha256"`
}

type evidence struct {
	Schema                  string        `json:"schema"`
	Repository              string        `json:"repository"`
	Event                   string        `json:"event"`
	BaseRef                 string        `json:"base_ref"`
	BaseSHA                 string        `json:"base_sha"`
	HeadSHA                 string        `json:"head_sha"`
	RunID                   int64         `json:"run_id"`
	RunAttempt              int64         `json:"run_attempt"`
	WorkflowSHA             string        `json:"workflow_sha"`
	Toolchain               string        `json:"toolchain"`
	SlotPreservation        bool          `json:"slot_preservation"`
	NoWriteOutsideGenerated bool          `json:"no_write_outside_generated"`
	Jobs                    []jobEvidence `json:"jobs"`
	Digests                 digests       `json:"digests"`
}

type metadata struct {
	Repository  string
	Event       string
	BaseRef     string
	BaseSHA     string
	HeadSHA     string
	RunID       int64
	RunAttempt  int64
	WorkflowSHA string
	Toolchain   string
}

func readMetadata() (metadata, error) {
	values := map[string]string{
		"repository":   os.Getenv("CI_REPOSITORY"),
		"event":        os.Getenv("CI_EVENT"),
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
	toolchain, err := toolchainIdentity()
	if err != nil {
		return metadata{}, err
	}
	return metadata{Repository: values["repository"], Event: values["event"], BaseRef: values["base_ref"], BaseSHA: values["base_sha"], HeadSHA: values["head_sha"], RunID: runID, RunAttempt: attempt, WorkflowSHA: values["workflow_sha"], Toolchain: toolchain}, nil
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

func normalizeJobs(apiJobs []apiJob, headSHA string, selfPolicySuccess bool) ([]jobEvidence, error) {
	byName := make(map[string]apiJob)
	for _, job := range apiJobs {
		if !jobSet()[job.Name] {
			continue
		}
		if _, duplicate := byName[job.Name]; duplicate {
			return nil, fmt.Errorf("duplicate canonical CI job %q", job.Name)
		}
		if job.Name == "CI policy" && job.Conclusion == "" && selfPolicySuccess {
			job.Conclusion = "success"
		}
		byName[job.Name] = job
	}
	result := make([]jobEvidence, 0, len(canonicalJobs))
	for _, name := range canonicalJobs {
		job, ok := byName[name]
		if !ok || job.ID <= 0 || job.Conclusion != "success" || !validSHA(job.HeadSHA) || job.HeadSHA != headSHA {
			return nil, fmt.Errorf("canonical CI job %q is missing or mismatched", name)
		}
		result = append(result, jobEvidence{ID: job.ID, Name: job.Name, Conclusion: job.Conclusion, HeadSHA: job.HeadSHA})
	}
	return result, nil
}

func jobSet() map[string]bool {
	result := make(map[string]bool, len(canonicalJobs))
	for _, name := range canonicalJobs {
		result[name] = true
	}
	return result
}
