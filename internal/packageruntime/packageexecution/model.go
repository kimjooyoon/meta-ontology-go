package packageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

const ReceiptSchema = "gooo/package-source-execution-receipt/v1"

type Source struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type Request struct {
	PackagePath string   `json:"package_path"`
	Entry       string   `json:"entry"`
	Sources     []Source `json:"sources"`
}

type SourceEvidence struct {
	Filename         string `json:"filename"`
	Digest           string `json:"digest"`
	DeclarationCount int    `json:"declaration_count"`
}

type Event struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Subject  string `json:"subject"`
}

type Diagnostic struct {
	Stage    string `json:"stage"`
	Code     string `json:"code"`
	Filename string `json:"filename,omitempty"`
	Message  string `json:"message"`
}

type Effects struct {
	RepositoryWrites  int  `json:"repository_writes"`
	MutationAuthority bool `json:"mutation_authority"`
}

type Receipt struct {
	Schema               string                   `json:"schema"`
	Decision             string                   `json:"decision"`
	Reason               string                   `json:"reason"`
	Resolution           string                   `json:"resolution"`
	PackagePath          string                   `json:"package_path"`
	Package              string                   `json:"package,omitempty"`
	Namespace            string                   `json:"namespace,omitempty"`
	Entry                string                   `json:"entry"`
	Sources              []SourceEvidence         `json:"sources"`
	CombinedSourceDigest string                   `json:"combined_source_digest,omitempty"`
	SemanticDigest       string                   `json:"semantic_digest,omitempty"`
	Execution            *sourceexecution.Receipt `json:"execution,omitempty"`
	Events               []Event                  `json:"events"`
	Diagnostics          []Diagnostic             `json:"diagnostics"`
	Effects              Effects                  `json:"effects"`
	Digest               string                   `json:"digest"`
}
