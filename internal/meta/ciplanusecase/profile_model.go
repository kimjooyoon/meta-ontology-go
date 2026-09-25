package ciplanusecase

type ProfileSample struct {
	CaseID       string `json:"case_id"`
	WallMS       int64  `json:"wall_ms"`
	PeakRSSKiB   int64  `json:"peak_rss_kib"`
	ReceiptBytes int64  `json:"receipt_bytes"`
}

type Profile struct {
	Schema  string          `json:"schema"`
	Samples []ProfileSample `json:"samples"`
}

type SourceProfile struct {
	GoooFiles int `json:"gooo_files"`
	GoFiles   int `json:"go_files"`
	GoooLines int `json:"gooo_lines"`
	GoLines   int `json:"go_lines"`
}
