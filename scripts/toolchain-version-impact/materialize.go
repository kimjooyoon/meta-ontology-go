package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

const projectionActivityID = "urn:gooo:toolchain-projection:assess"

type projectionDecision struct {
	ModuleSHA256   string `json:"module_sha256"`
	PlanCoreSHA256 string `json:"plan_core_sha256"`
	CurrentVersion string `json:"current_version"`
	NextVersion    string `json:"next_version"`
	CandidateEdits int    `json:"candidate_edits"`
	DerivedOutputs int    `json:"derived_outputs"`
	Status         string `json:"status"`
}

type projectionReceipt struct {
	Schema                  string             `json:"schema"`
	Scope                   string             `json:"scope"`
	OutsideScopeOccurrences int                `json:"outside_scope_occurrences"`
	PlanCoreSHA256          string             `json:"plan_core_sha256"`
	Decision                projectionDecision `json:"gooo_decision"`
	Files                   []projectedFile    `json:"files"`
	Provenance              map[string]any     `json:"provenance"`
}

type projectedFile struct {
	Path               string `json:"path"`
	InputSHA256        string `json:"input_sha256"`
	OutputSHA256       string `json:"output_sha256"`
	Edits              int    `json:"edits"`
	DerivationRule     string `json:"derivation_rule,omitempty"`
	SourcePath         string `json:"source_path,omitempty"`
	SourceOutputSHA256 string `json:"source_output_sha256,omitempty"`
}

type candidateImage struct {
	path string
	data []byte
	mode os.FileMode
}

// executeProjectionDecision runs the source-owned Gooo decision with the exact
// plan inputs. The adapter only consumes its observed output; it has no policy
// fallback when execution fails or the decision is unresolved.
func executeProjectionDecision(inputs map[string]any) (projectionDecision, error) {
	data, err := json.Marshal(inputs)
	if err != nil {
		return projectionDecision{}, err
	}
	input, err := os.CreateTemp("", "gooo-projection-input-*.json")
	if err != nil {
		return projectionDecision{}, err
	}
	defer os.Remove(input.Name())
	if _, err := input.Write(data); err != nil {
		input.Close()
		return projectionDecision{}, err
	}
	if err := input.Close(); err != nil {
		return projectionDecision{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "./cmd/gooo", "package", "execute", "--json", "--inputs", input.Name(),
		"examples/toolchain-version-impact/plan.workspace.json")
	output, err := command.Output()
	if err != nil {
		return projectionDecision{}, fmt.Errorf("execute Gooo projection decision: %w", err)
	}
	var observed struct {
		Decision string `json:"decision"`
		Result   struct {
			Runtime struct {
				ModelCalls         int  `json:"model_calls"`
				ProjectionReplayed bool `json:"projection_replayed"`
				RuntimeReplayed    bool `json:"runtime_replayed"`
				Traces             []struct {
					Deliveries []struct {
						ActivityID string             `json:"activity_id"`
						Actual     projectionDecision `json:"actual"`
					} `json:"deliveries"`
				} `json:"traces"`
			} `json:"runtime"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output, &observed); err != nil {
		return projectionDecision{}, fmt.Errorf("decode Gooo decision: %w", err)
	}
	runtime := observed.Result.Runtime
	if observed.Decision != "OBSERVED" || runtime.ModelCalls != 0 || !runtime.ProjectionReplayed ||
		!runtime.RuntimeReplayed || len(runtime.Traces) != 1 || len(runtime.Traces[0].Deliveries) != 1 ||
		runtime.Traces[0].Deliveries[0].ActivityID != projectionActivityID {
		return projectionDecision{}, fmt.Errorf("Gooo projection decision has no single deterministic replay")
	}
	return runtime.Traces[0].Deliveries[0].Actual, nil
}

// materializeCandidate writes only to a new directory outside the scanned tree.
// A fresh scan and every exact preimage are checked before the first write.
func materializeCandidate(root, destination string, plan projectionPlan, decision projectionDecision) (_ projectionReceipt, err error) {
	if decision.Status != "READY_FOR_DETACHED_PROJECTION" ||
		decision.ModuleSHA256 != plan.ModuleSHA256 || decision.PlanCoreSHA256 != plan.PlanCoreSHA256 ||
		decision.CurrentVersion != plan.CurrentVersion || decision.NextVersion != plan.NextVersion ||
		decision.CandidateEdits != len(plan.CandidateEdits) || decision.DerivedOutputs != len(plan.DerivedOutputs) {
		return projectionReceipt{}, fmt.Errorf("Gooo did not authorize this exact projection: %s", decision.Status)
	}
	if plan.UnresolvedOccurrences != 0 || plan.StaleSelectors != 0 || plan.OverlappingEdits != 0 ||
		plan.HistoricalWrites != 0 || plan.OutsideScopeWrites != 0 || plan.UnresolvedDerivedOutputs != 0 ||
		len(plan.CandidateEdits) != plan.ActiveOccurrences || len(plan.CandidateEdits) == 0 {
		return projectionReceipt{}, fmt.Errorf("projection has an open active scope")
	}
	fresh, err := scan(root)
	if err != nil {
		return projectionReceipt{}, err
	}
	freshPlan, err := buildProjectionPlan(fresh, plan.NextVersion)
	if err != nil || freshPlan.PlanCoreSHA256 != plan.PlanCoreSHA256 {
		return projectionReceipt{}, fmt.Errorf("projection observations changed before materialization")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return projectionReceipt{}, err
	}
	rootAbs, err = filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return projectionReceipt{}, err
	}
	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return projectionReceipt{}, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(destAbs))
	if err != nil {
		return projectionReceipt{}, fmt.Errorf("candidate parent must exist: %w", err)
	}
	destAbs = filepath.Join(parent, filepath.Base(destAbs))
	if relative, err := filepath.Rel(rootAbs, destAbs); err != nil || relative == "." ||
		(relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return projectionReceipt{}, fmt.Errorf("candidate destination must be outside the source tree")
	}
	groups := map[string][]projectionEdit{}
	for _, edit := range plan.CandidateEdits {
		groups[edit.Path] = append(groups[edit.Path], edit)
	}
	paths := make([]string, 0, len(groups))
	for path := range groups {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	images := make([]candidateImage, 0, len(paths)+len(plan.DerivedOutputs))
	receipt := projectionReceipt{Schema: "gooo/toolchain-detached-projection/v1",
		Scope: "direct module/workflow/Gooo spans and declared Gooo-derived evaluator/badge outputs", OutsideScopeOccurrences: plan.OutsideScope,
		PlanCoreSHA256: plan.PlanCoreSHA256, Decision: decision, Files: make([]projectedFile, 0, len(paths))}
	decisionJSON, err := json.Marshal(decision)
	if err != nil {
		return projectionReceipt{}, err
	}
	planID := "urn:gooo:toolchain-projection:plan:" + plan.PlanCoreSHA256
	decisionID := "urn:gooo:toolchain-projection:decision:" + digest(decisionJSON)
	graph := []any{
		map[string]any{"@id": planID, "@type": "prov:Entity", "gooo:core_sha256": plan.PlanCoreSHA256},
		map[string]any{"@id": decisionID, "@type": "prov:Entity", "gooo:content_sha256": digest(decisionJSON)},
	}
	projected := make(map[string][]byte, len(paths))
	for _, path := range paths {
		if !activePath(path) || filepath.Clean(path) != path || filepath.IsAbs(path) || strings.HasPrefix(path, "..") {
			return projectionReceipt{}, fmt.Errorf("invalid projection path %q", path)
		}
		originalPath := filepath.Join(rootAbs, filepath.FromSlash(path))
		resolved, err := filepath.EvalSymlinks(originalPath)
		if err != nil {
			return projectionReceipt{}, err
		}
		if relative, err := filepath.Rel(rootAbs, resolved); err != nil || relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return projectionReceipt{}, fmt.Errorf("projection source escapes repository: %s", path)
		}
		info, err := os.Lstat(originalPath)
		if err != nil || !info.Mode().IsRegular() {
			return projectionReceipt{}, fmt.Errorf("projection source is not a regular file: %s", path)
		}
		original, err := os.ReadFile(originalPath)
		if err != nil {
			return projectionReceipt{}, err
		}
		edits := groups[path]
		candidate := bytes.Clone(original)
		for _, edit := range slices.Backward(edits) {
			if edit.PreimageSHA256 != digest(original) || edit.StartByte < 0 || edit.EndByte > len(original) ||
				edit.StartByte >= edit.EndByte || string(original[edit.StartByte:edit.EndByte]) != edit.Before ||
				edit.After != plan.NextVersion {
				return projectionReceipt{}, fmt.Errorf("projection preimage changed at %s:%d", path, edit.Line)
			}
			candidate = append(append(bytes.Clone(candidate[:edit.StartByte]), edit.After...), candidate[edit.EndByte:]...)
		}
		images = append(images, candidateImage{path, candidate, info.Mode().Perm()})
		projected[path] = candidate
		file := projectedFile{Path: path, InputSHA256: digest(original), OutputSHA256: digest(candidate), Edits: len(edits)}
		receipt.Files = append(receipt.Files, file)
		pathID := strings.TrimPrefix(digest([]byte(path)), "sha256:")
		inputID := "urn:gooo:toolchain-projection:input:" + pathID + ":" + strings.TrimPrefix(file.InputSHA256, "sha256:")
		outputID := "urn:gooo:toolchain-projection:output:" + pathID + ":" + strings.TrimPrefix(file.OutputSHA256, "sha256:")
		activityID := "urn:gooo:toolchain-projection:write:" + pathID + ":" + strings.TrimPrefix(plan.PlanCoreSHA256, "sha256:")
		graph = append(graph,
			map[string]any{"@id": inputID, "@type": "prov:Entity", "gooo:path": path, "gooo:content_sha256": file.InputSHA256},
			map[string]any{"@id": activityID, "@type": "prov:Activity", "prov:used": []any{
				map[string]string{"@id": inputID}, map[string]string{"@id": planID}, map[string]string{"@id": decisionID}}},
			map[string]any{"@id": outputID, "@type": "prov:Entity", "gooo:path": path, "gooo:content_sha256": file.OutputSHA256,
				"prov:wasGeneratedBy": map[string]string{"@id": activityID}, "prov:wasDerivedFrom": map[string]string{"@id": inputID}})
	}
	derivedImages, err := appendDerivedCandidates(rootAbs, plan, projected, planID, decisionID, &receipt, &graph)
	if err != nil {
		return projectionReceipt{}, err
	}
	images = append(images, derivedImages...)
	receipt.Provenance = map[string]any{"@context": map[string]string{"prov": "http://www.w3.org/ns/prov#",
		"gooo": "urn:gooo:toolchain-projection:vocab:"}, "@graph": graph}
	if err := os.Mkdir(destAbs, 0o755); err != nil {
		return projectionReceipt{}, fmt.Errorf("create new candidate directory: %w", err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(destAbs)
		}
	}()
	for _, file := range images {
		output := filepath.Join(destAbs, filepath.FromSlash(file.path))
		if err = os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			return projectionReceipt{}, err
		}
		if err = os.WriteFile(output, file.data, file.mode); err != nil {
			return projectionReceipt{}, err
		}
	}
	receiptData, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return projectionReceipt{}, err
	}
	if err = os.WriteFile(filepath.Join(destAbs, "projection-receipt.json"), append(receiptData, '\n'), 0o644); err != nil {
		return projectionReceipt{}, err
	}
	return receipt, nil
}

func projectionOutputID(path, sha string) string {
	return "urn:gooo:toolchain-projection:output:" + strings.TrimPrefix(digest([]byte(path)), "sha256:") +
		":" + strings.TrimPrefix(sha, "sha256:")
}
