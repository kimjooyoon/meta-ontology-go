package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"

type config struct {
	repository     string
	predecessorSHA string
	branch         string
	workflow       string
	output         string
}

type workflowRunList struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

type workflowRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	RunAttempt int    `json:"run_attempt"`
}

type artifactList struct {
	TotalCount int        `json:"total_count"`
	Artifacts  []artifact `json:"artifacts"`
}

type artifact struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Expired     bool   `json:"expired"`
	SizeInBytes int64  `json:"size_in_bytes"`
	Digest      string `json:"digest"`
}

type archivedReceipt struct {
	ReceiptDigest    string `json:"receipt_digest"`
	RepositoryWrites int    `json:"repository_writes"`
	Schema           string `json:"schema"`
	Decision         string `json:"decision"`
	Report           struct {
		Schema         string `json:"schema"`
		SourceDecision string `json:"source_decision"`
		Decision       string `json:"decision"`
		Reason         string `json:"reason"`
		ReportDigest   string `json:"report_digest"`
		Feedback       struct {
			Schema       string `json:"schema"`
			CommitSHA    string `json:"commit_sha"`
			Repository   string `json:"repository"`
			Decision     string `json:"decision"`
			Reason       string `json:"reason"`
			ReportDigest string `json:"report_digest"`
		} `json:"feedback"`
	} `json:"report"`
}

type collection struct {
	Input feedbackpredecessor.Input
}
