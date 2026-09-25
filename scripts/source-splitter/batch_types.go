package main

const (
	logicalSplitPlanSchema = "gooo.logical-split-plan.v1"
	splitBatchSchema       = "gooo.logical-source-split.v1"
	splitBatchOperation    = "split-logical-declarations"
)

type logicalSplitPlan struct {
	Schema    string                `json:"schema"`
	SourceSHA string                `json:"source_sha"`
	Subjects  []logicalSplitSubject `json:"subjects"`
}

type logicalSplitSubject struct {
	Logical   string `json:"logical"`
	Reason    string `json:"reason"`
	Consumer  string `json:"consumer"`
	Operation string `json:"meta_operation"`
}

type plannedSplit struct {
	logical string
	plan    splitPlan
}

type splitBatchSubject struct {
	Logical       string   `json:"logical"`
	Status        string   `json:"status"`
	ChangedFiles  []string `json:"changed_files"`
	CreatedFiles  []string `json:"created_files"`
	ReceiptDigest string   `json:"receipt_digest"`
}
