package metriccounterfactual

const LedgerSchema = "gooo/metric-counterfactual-ledger/v1"

type Receipt struct {
	Kind         string `json:"kind"`
	Path         string `json:"path"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	BeforeLines  int    `json:"before_lines"`
	AfterLines   int    `json:"after_lines"`
}

type Delta struct {
	DirectFolders      int `json:"direct_folders"`
	DirectFiles        int `json:"direct_files"`
	RecursiveFolders   int `json:"recursive_folders"`
	RecursiveFiles     int `json:"recursive_files"`
	GoFiles            int `json:"go_files"`
	GoooFiles          int `json:"gooo_files"`
	GoLines            int `json:"go_lines"`
	GoooLines          int `json:"gooo_lines"`
	ChangedFiles       int `json:"changed_files"`
	ChangedDirectories int `json:"changed_directories"`
}

type Indicator struct {
	ID             string `json:"id"`
	Family         string `json:"family"`
	Trilemma       string `json:"trilemma"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Evidence struct {
	ManifestDigest   string `json:"manifest_digest"`
	PlanDigest       string `json:"plan_digest"`
	ReceiptSetDigest string `json:"receipt_set_digest"`
	BeforeDigest     string `json:"before_state_digest"`
	AfterDigest      string `json:"after_state_digest"`
}

type Ledger struct {
	Schema                    string      `json:"schema"`
	Repository                string      `json:"repository"`
	SubjectSHA                string      `json:"subject_sha"`
	ExecutionPolicy           string      `json:"execution_policy"`
	RepositoryWorkspaceWrites bool        `json:"repository_workspace_writes"`
	Manifest                  Manifest    `json:"manifest"`
	Plan                      Plan        `json:"plan"`
	Receipts                  []Receipt   `json:"receipts"`
	Before                    State       `json:"before"`
	After                     State       `json:"after"`
	Delta                     Delta       `json:"delta"`
	Evidence                  Evidence    `json:"evidence"`
	Indicators                []Indicator `json:"indicators"`
	PromotionAuthorized       bool        `json:"promotion_authorized"`
	Digest                    string      `json:"digest"`
}
