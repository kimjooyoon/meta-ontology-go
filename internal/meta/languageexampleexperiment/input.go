package languageexampleexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

type Input struct {
	ExpectedHead   string
	Contract       Contract
	Golden         Golden
	Artifact       artifactemit.Artifact
	Replay         artifactemit.Artifact
	UnknownEmitter artifactemit.Artifact
	Profile        Profile
}
