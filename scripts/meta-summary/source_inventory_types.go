package main

import "encoding/json"

type sourceDirectoryEvidence struct {
	Path             string `json:"path"`
	RecursiveFolders int    `json:"recursive_folders"`
	RecursiveFiles   int    `json:"recursive_files"`
	GoFiles          int    `json:"go_files"`
	GoLines          int    `json:"go_lines"`
	GoooFiles        int    `json:"gooo_files"`
	GoooLines        int    `json:"gooo_lines"`
}

type sourcePolicyEvidence struct {
	Schema                  string `json:"schema"`
	MaxFileLines            int    `json:"max_file_lines"`
	MaxFunctionLines        int    `json:"max_function_lines"`
	ExemptProjectRootREADME bool   `json:"exempt_project_root_readme"`
}

type sourceIndicatorEvidence struct {
	MetricID            string `json:"metric_id"`
	Subject             string `json:"subject"`
	Value               int    `json:"value"`
	Limit               int    `json:"limit"`
	Applicability       string `json:"applicability"`
	ApplicabilityReason string `json:"applicability_reason"`
	Blocking            bool   `json:"blocking"`
	Role                string `json:"role"`
}

type sourceMetricsEvidence struct {
	Files       []json.RawMessage         `json:"files"`
	Directories []sourceDirectoryEvidence `json:"directories"`
	Meta        struct {
		Schema     string                    `json:"schema"`
		Policy     sourcePolicyEvidence      `json:"policy"`
		Indicators []sourceIndicatorEvidence `json:"indicators"`
	} `json:"meta"`
}

type planSelectionEvidence struct {
	SchemaVersion string            `json:"schema_version"`
	SelectedCount int               `json:"selected_count"`
	Selected      []selectedSubject `json:"selected"`
}
