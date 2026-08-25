package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

const ProfileSchema = "gooo/language-example-experiment-profile/v1"

type Profile struct {
	Schema           string               `json:"schema"`
	SubjectSHA       string               `json:"subject_sha"`
	ExecutableDigest string               `json:"executable_digest"`
	GoooFiles        int                  `json:"gooo_files"`
	GoFiles          int                  `json:"go_files"`
	PrimaryArtifacts int                  `json:"primary_artifacts"`
	BinaryBytes      int64                `json:"binary_bytes"`
	Samples          []Sample             `json:"samples"`
	Effects          artifactemit.Effects `json:"effects"`
}

type Sample struct {
	Sequence int   `json:"sequence"`
	WallMS   int64 `json:"wall_ms"`
	RSSKiB   int64 `json:"rss_kib"`
}
