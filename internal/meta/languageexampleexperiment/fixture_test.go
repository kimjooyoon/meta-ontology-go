package languageexampleexperiment

import (
	"encoding/json"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
)

func validInput() Input {
	artifact := fixtureArtifact("PayOrder")
	samples := make([]Sample, 5)
	for index := range samples {
		samples[index] = Sample{Sequence: index + 1, WallMS: int64(index + 1), RSSKiB: int64(index + 1)}
	}
	return Input{
		ExpectedHead: "head", Contract: CanonicalContract(),
		Golden:   Golden{Package: artifact.Package, Operation: artifact.Operation},
		Artifact: artifact, Replay: artifact,
		UnknownEmitter: fixtureEmitter("not-registered", "PayOrder"),
		Profile: Profile{Schema: ProfileSchema, SubjectSHA: "head",
			ExecutableDigest: "sha256:" + strings.Repeat("a", 64), GoooFiles: 2,
			PrimaryArtifacts: 1, BinaryBytes: 100, Samples: samples},
	}
}

func fixtureArtifact(activity string) artifactemit.Artifact {
	return fixtureEmitter(artifactemit.OperationManifestKind, activity)
}

func fixtureEmitter(kind, activity string) artifactemit.Artifact {
	digest := "sha256:" + strings.Repeat("b", 64)
	payload, _ := json.Marshal(map[string]any{
		"schema": artifactemit.PackageReceiptSchema, "decision": "PASS", "resolution": "EXACT",
		"package_path": "billing-package", "package": "billing", "namespace": "billing", "entry": activity,
		"sources": []map[string]any{{"filename": "activity.gooo", "digest": digest, "declaration_count": 1},
			{"filename": "entities.gooo", "digest": digest, "declaration_count": 2}},
		"execution": map[string]any{"entry": map[string]any{"package": "billing", "namespace": "billing",
			"activity": activity, "inputs": []map[string]any{{"name": "Order", "id": "urn:order"}},
			"output": map[string]any{"name": "Receipt", "id": "urn:receipt"}}},
		"effects": map[string]any{"repository_writes": 0, "mutation_authority": false}, "digest": digest,
	})
	return artifactemit.Emit(kind, payload)
}
