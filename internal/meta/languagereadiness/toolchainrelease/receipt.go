package toolchainrelease

type BuildEvidence struct {
	VCSRevision string `json:"vcs_revision"`
	VCSModified bool   `json:"vcs_modified"`
	Trimpath    bool   `json:"trimpath"`
	BuildVCS    bool   `json:"build_vcs"`
	CGOEnabled  bool   `json:"cgo_enabled"`
}

type ReplayEvidence struct {
	Name        string `json:"name"`
	Digest      string `json:"digest"`
	Bytes       int64  `json:"bytes"`
	Builds      int    `json:"builds"`
	ReplayEqual bool   `json:"replay_equal"`
}

type SmokeEvidence struct {
	SchemaVersion string `json:"schema_version"`
	Language      string `json:"language"`
	Version       string `json:"version"`
	Status        string `json:"status"`
}

type ArchiveMemberEvidence struct {
	Role       string `json:"role"`
	Name       string `json:"name"`
	Digest     string `json:"digest"`
	Bytes      int64  `json:"bytes"`
	Mode       string `json:"mode"`
	SourcePath string `json:"source_path"`
	Module     string `json:"module"`
	Version    string `json:"version"`
	Ref        string `json:"ref"`
	SourceURL  string `json:"source_url"`
}

type PayloadEvidence struct {
	SourceSHA256     string                  `json:"source_sha256"`
	ProjectionSHA256 string                  `json:"projection_sha256"`
	Members          []ArchiveMemberEvidence `json:"members"`
}

type PlatformReceipt struct {
	Schema              string          `json:"schema"`
	Decision            string          `json:"decision"`
	Reason              string          `json:"reason"`
	Resolution          string          `json:"resolution"`
	HeadSHA             string          `json:"head_sha"`
	Toolchain           string          `json:"toolchain"`
	Platform            Target          `json:"platform"`
	Build               BuildEvidence   `json:"build"`
	Binary              ReplayEvidence  `json:"binary"`
	Archive             ReplayEvidence  `json:"archive"`
	Payload             PayloadEvidence `json:"payload"`
	ArchiveFormat       string          `json:"archive_format"`
	Smoke               SmokeEvidence   `json:"smoke"`
	RepositoryWrites    int             `json:"repository_writes"`
	MutationAuthorities int             `json:"mutation_authorities"`
	ReceiptDigest       string          `json:"receipt_digest"`
}

type PlatformEvidence struct {
	Receipt         PlatformReceipt
	ArchivePath     string
	ArchiveVerified bool
	ReceiptVerified bool
}
