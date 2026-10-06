package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	baselineArchiveLimit = 25_000_000
	baselineReceiptLimit = 1_000_000
)

type baselineRunList struct {
	TotalCount int           `json:"total_count"`
	Runs       []baselineRun `json:"workflow_runs"`
}

type baselineRun struct {
	ID         int64  `json:"id"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	RunAttempt int    `json:"run_attempt"`
}

type baselineArtifactList struct {
	TotalCount int                `json:"total_count"`
	Artifacts  []baselineArtifact `json:"artifacts"`
}

type baselineArtifact struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Expired     bool   `json:"expired"`
	SizeInBytes int64  `json:"size_in_bytes"`
	Digest      string `json:"digest"`
}

func discoverBaseline(ctx context.Context, current Report) (Report, []byte, baselineArtifact, error) {
	repository, currentRunID, currentSubject := current.Snapshot.Repository, current.Snapshot.WorkflowRun, current.SubjectSHA
	if repository == "" || currentRunID <= 0 || !canonicalSubject(currentSubject) || os.Getenv("GITHUB_TOKEN") == "" {
		return Report{}, nil, baselineArtifact{}, fmt.Errorf("exact repository, run, commit, and GITHUB_TOKEN are required")
	}
	baseURL := strings.TrimRight(os.Getenv("GITHUB_API_URL"), "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	token := os.Getenv("GITHUB_TOKEN")
	client := &http.Client{Timeout: 45 * time.Second}
	query := url.Values{"branch": {"dev"}, "event": {"push"}, "status": {"completed"}, "per_page": {"100"}}
	endpoint := fmt.Sprintf("%s/repos/%s/actions/workflows/language-utility-evidence.yml/runs?%s", baseURL, repository, query.Encode())
	var runs baselineRunList
	if err := githubJSON(ctx, client, token, endpoint, &runs); err != nil {
		return Report{}, nil, baselineArtifact{}, err
	}
	checked := 0
	for _, run := range runs.Runs {
		if run.ID <= 0 || run.ID == currentRunID || run.HeadSHA == currentSubject || run.HeadBranch != "dev" || run.Event != "push" ||
			run.Status != "completed" || run.Conclusion != "success" || !canonicalSubject(run.HeadSHA) {
			continue
		}
		if checked == 20 {
			break
		}
		checked++
		expectedName := "language-utility-evidence-" + run.HeadSHA
		artifactEndpoint := fmt.Sprintf("%s/repos/%s/actions/runs/%d/artifacts?per_page=100", baseURL, repository, run.ID)
		var artifacts baselineArtifactList
		if err := githubJSON(ctx, client, token, artifactEndpoint, &artifacts); err != nil {
			return Report{}, nil, baselineArtifact{}, err
		}
		if artifacts.TotalCount > len(artifacts.Artifacts) {
			return Report{}, nil, baselineArtifact{}, fmt.Errorf("artifact page for run %d is incomplete", run.ID)
		}
		artifact, found := exactBaselineArtifact(artifacts.Artifacts, expectedName)
		if !found {
			continue
		}
		if artifact.SizeInBytes > maximumEvidenceBytes-current.Investment.Observed.EvidenceBytes ||
			current.Investment.Observed.EvidenceFiles+1 > maximumEvidenceFiles {
			continue
		}
		archiveEndpoint := fmt.Sprintf("%s/repos/%s/actions/artifacts/%d/zip", baseURL, repository, artifact.ID)
		archive, err := githubBytes(ctx, client, token, archiveEndpoint, baselineArchiveLimit)
		if err != nil {
			continue
		}
		if artifactDigest(archive) != artifact.Digest {
			continue
		}
		if int64(len(archive)) > maximumEvidenceBytes-current.Investment.Observed.EvidenceBytes {
			continue
		}
		artifact.SizeInBytes = int64(len(archive))
		receiptRaw, err := baselineReceipt(archive)
		if err != nil {
			continue
		}
		var receipt Report
		if err := json.Unmarshal(receiptRaw, &receipt); err != nil || receipt.SubjectSHA != run.HeadSHA ||
			receipt.Snapshot.WorkflowRun != run.ID || receipt.Snapshot.RunAttempt != run.RunAttempt ||
			receipt.Snapshot.Repository != repository {
			continue
		}
		comparison := compareReports(current, receipt, true)
		if comparison.Status != "COMPARABLE" && comparison.Status != "PARTIAL" {
			continue
		}
		fmt.Printf("domain-completeness baseline: run=%d attempt=%d head=%s artifact=%d digest=%s\n",
			run.ID, run.RunAttempt, run.HeadSHA, artifact.ID, artifact.Digest)
		return receipt, receiptRaw, artifact, nil
	}
	return Report{}, nil, baselineArtifact{}, fmt.Errorf("no compatible retained receipt found in the newest 20 successful dev pushes")
}

func exactBaselineArtifact(artifacts []baselineArtifact, expected string) (baselineArtifact, bool) {
	var selected baselineArtifact
	matches := 0
	for _, artifact := range artifacts {
		if artifact.Name == expected {
			selected = artifact
			matches++
		}
	}
	if matches != 1 || selected.ID <= 0 || selected.Expired || selected.SizeInBytes <= 0 ||
		selected.SizeInBytes > baselineArchiveLimit || !strings.HasPrefix(selected.Digest, "sha256:") {
		return baselineArtifact{}, false
	}
	return selected, true
}

func baselineReceipt(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	var found []byte
	for _, file := range reader.File {
		if path.Clean(file.Name) != "domain-completeness/receipt.json" || file.UncompressedSize64 > baselineReceiptLimit {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("duplicate receipt entry")
		}
		input, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(io.LimitReader(input, baselineReceiptLimit+1))
		closeErr := input.Close()
		if readErr != nil || closeErr != nil || len(data) > baselineReceiptLimit {
			return nil, fmt.Errorf("receipt exceeds bounded read")
		}
		found = data
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("receipt entry is missing")
	}
	return found, nil
}

func githubJSON(ctx context.Context, client *http.Client, token, endpoint string, output any) error {
	data, err := githubBytes(ctx, client, token, endpoint, 2_000_000)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode GitHub API response: %w", err)
	}
	return nil
}

func githubBytes(ctx context.Context, client *http.Client, token, endpoint string, limit int) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("GitHub response exceeds %d byte limit", limit)
	}
	return data, nil
}

func artifactDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
