package ciplanusecase

type GoldenReason struct {
	Operation  string `json:"operation"`
	File       string `json:"file"`
	SourcePath string `json:"source_path"`
	SourceLine int    `json:"source_line"`
}

type GoldenCheck struct {
	ID      string         `json:"id"`
	Command string         `json:"command"`
	Files   []string       `json:"files"`
	Reasons []GoldenReason `json:"reasons"`
}

type GoldenPlan struct {
	Schema string        `json:"schema"`
	CaseID string        `json:"case_id"`
	Checks []GoldenCheck `json:"checks"`
}
