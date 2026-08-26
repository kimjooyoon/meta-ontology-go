package linecaps

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"

// FileLanguage indicates which extension is used to classify source files.
type FileLanguage string

const (
	FileLanguageGo       FileLanguage = "go"
	FileLanguageGooo     FileLanguage = "gooo"
	FileLanguageOther    FileLanguage = "other"
	maxSummaryIndicators              = 100
)

// FileMetric is a per-file line-count metric for recognized source files.
type FileMetric struct {
	Path     string       `json:"path"`
	Language FileLanguage `json:"language"`
	Lines    int          `json:"lines"`
}

// DirectoryMetric is a directory-level aggregate.
type DirectoryMetric struct {
	Path             string                   `json:"path"`
	SubjectKind      sourcepolicy.SubjectKind `json:"subject_kind"`
	DirectFolders    int                      `json:"direct_folders"`
	DirectFiles      int                      `json:"direct_files"`
	RecursiveFolders int                      `json:"recursive_folders"`
	RecursiveFiles   int                      `json:"recursive_files"`
	GoFiles          int                      `json:"go_files"`
	GoooFiles        int                      `json:"gooo_files"`
	GoLines          int                      `json:"go_lines"`
	GoooLines        int                      `json:"gooo_lines"`
}

// LineMetricsReport is the repository's line and layout metric output.
type LineMetricsReport struct {
	Repository         string              `json:"repository,omitempty"`
	CommitSHA          string              `json:"commit_sha,omitempty"`
	Root               string              `json:"root"`
	StorageRoot        string              `json:"storage_root,omitempty"`
	Files              []FileMetric        `json:"files"`
	Directories        []DirectoryMetric   `json:"directories"`
	StorageDirectories []DirectoryMetric   `json:"storage_directories,omitempty"`
	Meta               sourcepolicy.Report `json:"meta"`
}
type directoryNode struct {
	directFolders    int
	directFiles      int
	recursiveFolders int
	recursiveFiles   int
	goFiles          int
	goooFiles        int
	goLines          int
	goooLines        int
}
