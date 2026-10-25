package main

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

type activitySignature struct {
	Name         string
	Inputs       []string
	Output       string
	InputOrigins map[string]semantic.Span
}

type contractModel struct {
	EntityByID   map[string]string
	Activities   map[string]activitySignature
	SemanticHash string
}

type Indicator struct {
	ID       string `json:"id"`
	Route    string `json:"route"`
	Verdict  string `json:"verdict"`
	Relation string `json:"relation"`
	Value    string `json:"value"`
	Limit    string `json:"limit"`
}

type Report struct {
	Schema              string             `json:"schema"`
	Metaprogram         string             `json:"metaprogram"`
	CommitSHA           string             `json:"commit_sha"`
	ContractPath        string             `json:"contract_path"`
	SourceSHA256        string             `json:"source_sha256"`
	SemanticHash        string             `json:"semantic_hash"`
	RegistryDigest      string             `json:"registry_digest"`
	Status              string             `json:"status"`
	Reason              string             `json:"reason"`
	EntityCount         int                `json:"entity_count"`
	ActivityCount       int                `json:"activity_count"`
	Registry            []RegistryBinding  `json:"registry"`
	ExecutorCoverage    []ExecutorCoverage `json:"executor_coverage"`
	Errors              []string           `json:"errors"`
	Indicators          []Indicator        `json:"indicators"`
	PromotionAuthorized bool               `json:"promotion_authorized"`
}

type analysis struct {
	Report        Report
	SemanticOK    bool
	LoopOK        bool
	ExecutorOK    bool
	TrilemmaOK    bool
	ObservationOK bool
}
