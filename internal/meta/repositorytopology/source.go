package repositorytopology

type SourceReport struct {
	Repository         string            `json:"repository"`
	CommitSHA          string            `json:"commit_sha"`
	Root               string            `json:"root"`
	StorageRoot        string            `json:"storage_root"`
	Files              []FileMetric      `json:"files"`
	Directories        []DirectoryMetric `json:"directories"`
	StorageDirectories []DirectoryMetric `json:"storage_directories"`
	Meta               SourceMeta        `json:"meta"`
}

type SourceMeta struct {
	Schema     string            `json:"schema"`
	Policy     SourcePolicy      `json:"policy"`
	Indicators []SourceIndicator `json:"indicators"`
}

type SourcePolicy struct {
	Schema                        string `json:"schema"`
	MaxFileLines                  int    `json:"max_file_lines"`
	MaxFunctionLines              int    `json:"max_function_lines"`
	MaxDirectDirectoryEntries     int    `json:"max_direct_directory_entries"`
	RequireHomogeneousDirectories bool   `json:"require_homogeneous_directories"`
	ExemptProjectRootTopology     bool   `json:"exempt_project_root_topology"`
	ExemptProjectRootREADME       bool   `json:"exempt_project_root_readme"`
}

type FileMetric struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Lines    int    `json:"lines"`
}

type DirectoryMetric struct {
	Path             string `json:"path"`
	SubjectKind      string `json:"subject_kind"`
	DirectFolders    int    `json:"direct_folders"`
	DirectFiles      int    `json:"direct_files"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
	GoFiles          int    `json:"go_files"`
	GoooFiles        int    `json:"gooo_files"`
	GoLines          int    `json:"go_lines"`
	GoooLines        int    `json:"gooo_lines"`
}
