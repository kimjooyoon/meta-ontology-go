package changedsurfacereceipt

type Input struct {
	Schema          string    `json:"schema"`
	SubjectSHA      string    `json:"subject_sha"`
	ChangedSurfaces []string  `json:"changed_surfaces"`
	Receipts        []Receipt `json:"receipts"`
}

type Receipt struct {
	SurfaceID  string `json:"surface_id"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
}

type Summary struct {
	ChangedSurfaces   int `json:"changed_surfaces"`
	ReceiptsObserved  int `json:"receipts_observed"`
	BoundReceipts     int `json:"bound_receipts"`
	MissingReceipts   int `json:"missing_receipts"`
	OrphanReceipts    int `json:"orphan_receipts"`
	ChangedDuplicates int `json:"changed_duplicates"`
	ReceiptDuplicates int `json:"receipt_duplicates"`
	UnknownReceipts   int `json:"unknown_receipts"`
	MalformedPaths    int `json:"malformed_paths"`
	TotalityBPS       int `json:"totality_bps"`
	ChangedSetBPS     int `json:"changed_set_bps"`
	UniqueBindingBPS  int `json:"unique_binding_bps"`
	UnknownPaths      int `json:"unknown_paths"`
	BlockedPaths      int `json:"blocked_paths"`
}

type Report struct {
	Schema            string                    `json:"schema"`
	SubjectSHA        string                    `json:"subject_sha"`
	Decision          string                    `json:"decision"`
	Resolution        string                    `json:"resolution"`
	EnforcementEffect string                    `json:"enforcement_effect"`
	Reason            string                    `json:"reason"`
	DenominatorID     string                    `json:"denominator_id"`
	DenominatorDigest string                    `json:"denominator_digest"`
	Summary           Summary                   `json:"summary"`
	Indicators        []Indicator               `json:"indicators"`
	MetaOperations    []MetaOperationDefinition `json:"meta_operations"`
	RepositoryWrites  int                       `json:"repository_writes"`
	ReportDigest      string                    `json:"report_digest"`
}
