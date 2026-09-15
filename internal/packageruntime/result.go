package packageruntime

type Event struct {
	Sequence    int    `json:"sequence"`
	Kind        string `json:"kind"`
	PackagePath string `json:"package_path"`
	Activity    string `json:"activity,omitempty"`
}

type Result struct {
	Schema             string  `json:"schema"`
	Image              Image   `json:"image"`
	Events             []Event `json:"events"`
	Effects            int     `json:"effects"`
	RepositoryWrites   int     `json:"repository_writes"`
	MutationAuthorized bool    `json:"mutation_authorized"`
	ResultDigest       string  `json:"result_digest"`
}
