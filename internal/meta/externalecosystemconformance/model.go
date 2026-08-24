package externalecosystemconformance

import "strings"

type Document struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

type Capability struct {
	ID            string `json:"id"`
	Relation      string `json:"relation"`
	MetaOperation string `json:"meta_operation"`
	Status        string `json:"status"`
}

type Capsule struct {
	Schema          string       `json:"schema"`
	ReferenceID     string       `json:"reference_id"`
	RepositoryURL   string       `json:"repository_url"`
	CommitSHA       string       `json:"commit_sha"`
	TreeSHA         string       `json:"tree_sha"`
	LicenseSPDX     string       `json:"license_spdx"`
	ModulePath      string       `json:"module_path"`
	ModuleGoVersion string       `json:"module_go_version"`
	Documents       []Document   `json:"documents"`
	Capabilities    []Capability `json:"capabilities"`
}

type Evidence struct {
	Readme             []byte
	GoMod              []byte
	RepositoryWrites   int
	ExternalExecutions int
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

func validModule(raw []byte) bool {
	modulePath, goVersion := "", ""
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			modulePath = fields[1]
		}
		if len(fields) == 2 && fields[0] == "go" {
			goVersion = fields[1]
		}
	}
	return modulePath == ExpectedModule && goVersion == ExpectedGoVersion
}
