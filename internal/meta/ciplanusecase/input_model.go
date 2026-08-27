package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

type Input struct {
	Contract        Contract
	Reports         map[string]metainvocation.Report
	Replays         map[string]metainvocation.Report
	Goldens         map[string]GoldenPlan
	Profile         Profile
	Source          SourceProfile
	GeneratedReplay bool
}
