package artifactresolutionexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"

type Input struct {
	SubjectSHA      string                `json:"subject_sha"`
	Contract        Contract              `json:"contract"`
	Manifest        artifactemit.Artifact `json:"manifest"`
	ManifestReplay  artifactemit.Artifact `json:"manifest_replay"`
	ManifestGolden  artifactemit.Artifact `json:"manifest_golden"`
	Interface       artifactemit.Artifact `json:"interface"`
	InterfaceReplay artifactemit.Artifact `json:"interface_replay"`
	InterfaceGolden artifactemit.Artifact `json:"interface_golden"`
	UnknownEmitter  artifactemit.Artifact `json:"unknown_emitter"`
}
