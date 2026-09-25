package closure

type Input struct {
	Repository       string
	SubjectSHA       string
	RunID            int64
	RunAttempt       int
	Artifact         ArtifactIdentity
	ProgramJSON      []byte
	Source           []byte
	VerificationJSON []byte
}

type ArtifactIdentity struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Digest string `json:"digest"`
	URL    string `json:"url"`
}
