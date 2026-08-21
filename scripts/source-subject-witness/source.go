package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type sourceReport struct {
	Repository         string            `json:"repository"`
	CommitSHA          string            `json:"commit_sha"`
	Files              []fileMetric      `json:"files"`
	Directories        []directoryMetric `json:"directories"`
	StorageDirectories []directoryMetric `json:"storage_directories"`
	Meta               reportMeta        `json:"meta"`
}

type fileMetric struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Lines    int    `json:"lines"`
}

type directoryMetric struct {
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

type reportMeta struct {
	Schema     string            `json:"schema"`
	Policy     sourcePolicy      `json:"policy"`
	Indicators []sourceIndicator `json:"indicators"`
}

type sourcePolicy struct {
	Schema                        string `json:"schema"`
	MaxFileLines                  int    `json:"max_file_lines"`
	MaxFunctionLines              int    `json:"max_function_lines"`
	MaxDirectDirectoryEntries     int    `json:"max_direct_directory_entries"`
	RequireHomogeneousDirectories bool   `json:"require_homogeneous_directories"`
	ExemptProjectRootTopology     bool   `json:"exempt_project_root_topology"`
}

func loadSource(path string) (sourceReport, error) {
	var report sourceReport
	data, err := os.ReadFile(path)
	if err != nil {
		return report, fmt.Errorf("read source metrics: %w", err)
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return report, fmt.Errorf("decode source metrics: %w", err)
	}
	return report, nil
}
