package ciplanusecase

type Indicator struct {
	ID            string `json:"id"`
	Reader        string `json:"reader"`
	Observed      int64  `json:"observed"`
	Comparator    string `json:"comparator"`
	Target        int64  `json:"target"`
	Status        string `json:"status"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
}

type ReaderView struct {
	Reader       string   `json:"reader"`
	Resolution   string   `json:"resolution"`
	IndicatorIDs []string `json:"indicator_ids"`
}

type Proof struct {
	Choice   string   `json:"choice"`
	Status   string   `json:"status"`
	Evidence []string `json:"evidence"`
}

type Report struct {
	Schema         string       `json:"schema"`
	Decision       string       `json:"decision"`
	Resolution     string       `json:"resolution"`
	Interpretation string       `json:"interpretation"`
	ContractDigest string       `json:"contract_digest"`
	Cases          []CaseResult `json:"cases"`
	Summary        Summary      `json:"summary"`
	Indicators     []Indicator  `json:"indicators"`
	ReaderViews    []ReaderView `json:"reader_views"`
	Proofs         []Proof      `json:"proofs"`
	NotClaimed     []string     `json:"not_claimed"`
	ReportDigest   string       `json:"report_digest"`
}
