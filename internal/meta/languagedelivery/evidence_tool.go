package languagedelivery

type ConformanceReceipt struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	Summary struct {
		SurfacesSatisfied int `json:"surfaces_satisfied"`
		RepositoryWrites int `json:"repository_writes"`
		MutationAuthorities int `json:"mutation_authorities"`
	} `json:"summary"`
	Surfaces []ConformanceSurface `json:"surfaces"`
	RepositoryWrites int `json:"repository_writes"`
	MutationAuthorized bool `json:"mutation_authorized"`
}

type ConformanceSurface struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	HeadSHA string `json:"head_sha"`
}

type LSPReceipt struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	HeadSHA    string `json:"head_sha"`
	Summary struct {
		DiagnosticPaths int `json:"diagnostic_paths"`
		NavigationPaths int `json:"navigation_paths"`
		RepositoryWrites int `json:"repository_writes"`
		MutationAuthorities int `json:"mutation_authorities"`
	} `json:"summary"`
}

type ReleaseReceipt struct {
	Schema     string `json:"schema"`
	Decision   string `json:"decision"`
	Resolution string `json:"resolution"`
	HeadSHA    string `json:"head_sha"`
	Summary struct {
		PlatformReceipts int `json:"platform_receipts"`
		NativeSmokes int `json:"native_smokes"`
		RepositoryWrites int `json:"repository_writes"`
		MutationAuthorities int `json:"mutation_authorities"`
	} `json:"summary"`
}
