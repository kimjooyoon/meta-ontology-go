package provenance

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func normalizeBinding(id string, binding *RunBinding, protected bool) (*RunBinding, error) {
	if binding == nil {
		if protected {
			return nil, &BindingError{ID: id, Kind: "missing", Detail: ErrBindingRequired.Error()}
		}
		return nil, nil
	}
	normalized := *binding
	var err error
	for _, item := range []struct {
		field string
		value string
	}{
		{field: "repository", value: normalized.Repository},
		{field: "base", value: normalized.Base},
		{field: "head", value: normalized.Head},
		{field: "event_ref", value: normalized.EventRef},
		{field: "checkout_ref", value: normalized.CheckoutRef},
	} {
		if _, err = normalizeBindingText(item.field, item.value); err != nil {
			return nil, bindingError(id, "tuple", err)
		}
	}
	normalized.Repository = strings.TrimSpace(normalized.Repository)
	normalized.Base = strings.TrimSpace(normalized.Base)
	normalized.Head = strings.ToLower(strings.TrimSpace(normalized.Head))
	normalized.EventRef = strings.TrimSpace(normalized.EventRef)
	normalized.CheckoutRef = strings.TrimSpace(normalized.CheckoutRef)
	if err := validateCommitSHA(normalized.Head); err != nil {
		return nil, bindingError(id, "head", err)
	}
	if normalized.RunID <= 0 || normalized.RunAttempt <= 0 {
		return nil, bindingError(id, "run", fmt.Errorf("run_id and run_attempt must be positive"))
	}
	normalized.Workflow, err = normalizeWorkflow(normalized.Workflow)
	if err != nil {
		return nil, bindingError(id, "workflow", err)
	}
	normalized.Jobs, err = normalizeJobs(id, normalized.Jobs, normalized.Head, protected)
	if err != nil {
		return nil, err
	}
	for _, item := range []struct {
		name   string
		digest string
	}{
		{name: "policy_digest", digest: normalized.PolicyDigest},
		{name: "toolchain_digest", digest: normalized.ToolchainDigest},
		{name: "bundle_digest", digest: normalized.BundleDigest},
	} {
		if err := validateDigest(strings.ToLower(strings.TrimSpace(item.digest))); err != nil {
			return nil, bindingError(id, item.name, err)
		}
	}
	normalized.PolicyDigest = strings.ToLower(strings.TrimSpace(normalized.PolicyDigest))
	normalized.ToolchainDigest = strings.ToLower(strings.TrimSpace(normalized.ToolchainDigest))
	normalized.BundleDigest = strings.ToLower(strings.TrimSpace(normalized.BundleDigest))
	normalized.Predecessors, err = normalizeReferences(id, "predecessors", normalized.Predecessors, protected)
	if err != nil {
		return nil, err
	}
	normalized.EvidenceRefs, err = normalizeReferences(id, "evidence_refs", normalized.EvidenceRefs, protected)
	if err != nil {
		return nil, err
	}
	if normalized.WriteEffect != 0 {
		return nil, bindingError(id, "write-effect", fmt.Errorf("write_effect must be zero, got %d", normalized.WriteEffect))
	}
	return &normalized, nil
}

func normalizeBindingText(field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("%s must not contain line breaks", field)
	}
	return value, nil
}

func normalizeWorkflow(workflow WorkflowIdentity) (WorkflowIdentity, error) {
	var err error
	workflow.ID, err = normalizeBindingText("workflow.id", workflow.ID)
	if err != nil {
		return WorkflowIdentity{}, err
	}
	workflow.Name, err = normalizeBindingText("workflow.name", workflow.Name)
	if err != nil {
		return WorkflowIdentity{}, err
	}
	workflow.Path, err = normalizeBindingText("workflow.path", workflow.Path)
	if err != nil {
		return WorkflowIdentity{}, err
	}
	workflow.Ref = strings.TrimSpace(workflow.Ref)
	if strings.ContainsAny(workflow.Ref, "\r\n") {
		return WorkflowIdentity{}, fmt.Errorf("workflow.ref must not contain line breaks")
	}
	return workflow, nil
}

func normalizeJobs(id string, jobs CanonicalJobs, head string, protected bool) (CanonicalJobs, error) {
	if len(jobs) != requiredCanonicalJobs {
		return nil, bindingError(id, "jobs", fmt.Errorf("expected %d canonical jobs, got %d", requiredCanonicalJobs, len(jobs)))
	}
	result := append(CanonicalJobs(nil), jobs...)
	seen := make(map[string]struct{}, len(result))
	for index := range result {
		job := &result[index]
		var err error
		job.ID, err = normalizeBindingText("job.id", job.ID)
		if err != nil {
			return nil, bindingError(id, "jobs", err)
		}
		job.ID = strings.TrimSpace(job.ID)
		if _, exists := seen[job.ID]; exists {
			return nil, bindingError(id, "jobs", fmt.Errorf("duplicate job ID %q", job.ID))
		}
		seen[job.ID] = struct{}{}
		job.Conclusion = strings.ToLower(strings.TrimSpace(job.Conclusion))
		if !validConclusion(job.Conclusion) {
			return nil, bindingError(id, "jobs", fmt.Errorf("unsupported conclusion %q", job.Conclusion))
		}
		job.HeadSHA = strings.ToLower(strings.TrimSpace(job.HeadSHA))
		if err := validateCommitSHA(job.HeadSHA); err != nil {
			return nil, bindingError(id, "jobs", err)
		}
		if job.HeadSHA != head {
			return nil, bindingError(id, "jobs", fmt.Errorf("job %q head_sha does not match head", job.ID))
		}
		if protected && job.Conclusion != "success" {
			return nil, bindingError(id, "jobs", fmt.Errorf("job %q conclusion must be success", job.ID))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func normalizeReferences(id, field string, refs []string, required bool) ([]string, error) {
	if required && len(refs) == 0 {
		return nil, bindingError(id, field, fmt.Errorf("at least one reference is required"))
	}
	result := append([]string(nil), refs...)
	seen := make(map[string]struct{}, len(result))
	for index, ref := range result {
		ref, err := normalizeBindingText(fmt.Sprintf("%s[%d]", field, index), ref)
		if err != nil {
			return nil, bindingError(id, field, err)
		}
		if _, exists := seen[ref]; exists {
			return nil, bindingError(id, field, fmt.Errorf("duplicate reference %q", ref))
		}
		seen[ref] = struct{}{}
		result[index] = ref
	}
	sort.Strings(result)
	return result, nil
}

func validConclusion(conclusion string) bool {
	switch conclusion {
	case "success", "failure", "cancelled", "skipped", "neutral", "timed_out", "action_required", "stale":
		return true
	default:
		return false
	}
}

func validateCommitSHA(value string) error {
	if len(value) != 40 {
		return fmt.Errorf("commit SHA must be 40 hex characters")
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return fmt.Errorf("commit SHA must be lowercase hexadecimal")
		}
	}
	return nil
}

func bindingError(id, kind string, err error) error {
	return &BindingError{ID: id, Kind: kind, Detail: err.Error()}
}

func compareBindings(left, right *RunBinding) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func validateBindingSource(id string, evidence Evidence) error {
	if evidence.Binding == nil {
		return nil
	}
	if evidence.Freshness.SourceHash != evidence.Binding.BundleDigest {
		return bindingError(id, "source", fmt.Errorf("freshness source_hash does not match bundle_digest"))
	}
	return nil
}
