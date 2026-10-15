package workspaceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"

type WorkspaceRenameEdit struct {
	Start  int    `json:"start_byte"`
	End    int    `json:"end_byte"`
	Kind   string `json:"kind"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type WorkspaceRenameFile struct {
	PackagePath  string                `json:"package_path"`
	Filename     string                `json:"filename"`
	InputSHA256  string                `json:"input_sha256"`
	OutputSHA256 string                `json:"output_sha256"`
	Edits        []WorkspaceRenameEdit `json:"edits"`
	Source       string                `json:"source"`
}

type WorkspaceActivityRename struct {
	Schema             string                  `json:"schema"`
	ActivityID         string                  `json:"activity_id"`
	PackagePath        string                  `json:"package_path"`
	From               string                  `json:"from"`
	To                 string                  `json:"to"`
	InputImageDigest   string                  `json:"input_image_digest"`
	OutputImageDigest  string                  `json:"output_image_digest"`
	ResolvedReferences int                     `json:"resolved_references"`
	UpdatedReferences  int                     `json:"updated_references"`
	IdentityPinned     bool                    `json:"identity_pinned"`
	EntryUpdated       bool                    `json:"entry_updated"`
	Files              []WorkspaceRenameFile   `json:"files"`
	Manifest           packageruntime.Manifest `json:"manifest"`
	ProgramExecutions  int                     `json:"program_executions"`
	ModelPredictions   int                     `json:"model_predictions"`
	Scope              string                  `json:"scope"`
}

type renameReference struct {
	packagePath   string
	filename      string
	callerID      string
	calleeID      string
	calleePackage string
	surface       string
	start         int
	end           int
	qualified     bool
	literal       bool
}
