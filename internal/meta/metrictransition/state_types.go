package metrictransition

// Counts is the canonical repository size vector for one metric plane.
type Counts struct {
	DirectFolders    int `json:"direct_folders"`
	DirectFiles      int `json:"direct_files"`
	RecursiveFolders int `json:"recursive_folders"`
	RecursiveFiles   int `json:"recursive_files"`
	GoFiles          int `json:"go_files"`
	GoooFiles        int `json:"gooo_files"`
	GoLines          int `json:"go_lines"`
	GoooLines        int `json:"gooo_lines"`
}

// LanguageFile preserves each Go and Gooo file measurement.
type LanguageFile struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Lines    int    `json:"lines"`
}

// DirectoryState preserves direct and recursive counts for one directory.
type DirectoryState struct {
	Path        string `json:"path"`
	SubjectKind string `json:"subject_kind"`
	Counts      Counts `json:"counts"`
}
