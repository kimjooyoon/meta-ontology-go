package metacircularboundary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ExecuteReadOnlyMetaOperation(source SourceObservation, grant ExternalGrant, caseID string) (ExecutionArtifact, error) {
	if caseID != "explicit-read-only-capability" || !source.Graph.Valid || grant.Decision != GrantDecision || grant.Scope != ScopeReadOnly || grant.SubjectDigest != source.SemanticDigest || grant.Operation != MetaOperationID {
		return ExecutionArtifact{}, fmt.Errorf("read-only meta operation is not authorized")
	}
	output := fmt.Sprintf("operation=%s\nsource_semantic_digest=%s\ngraph_digest=%s\ncanonical_path=%s\n", MetaOperationID, source.SemanticDigest, source.Graph.Digest, source.Path)
	artifact := ExecutionArtifact{
		Schema: ExecutionSchema, Producer: ExecutionProducer, CaseID: caseID,
		Path: "execution/" + caseID + ".json", OperationID: MetaOperationID,
		GrantDigest: grant.GrantDigest, InputDigest: source.SemanticDigest,
		OutputCanonical: output, OutputDigest: digestBytes([]byte(output)),
	}
	artifact.ArtifactDigest = digestValue(artifact)
	return artifact, nil
}

func WriteExecutionArtifact(root string, artifact ExecutionArtifact) error {
	encoded, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	path := filepath.Join(root, filepath.FromSlash(artifact.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o644)
}

func validExecutionArtifact(source SourceObservation, grant ExternalGrant, caseID string, artifact ExecutionArtifact) bool {
	if artifact.Schema != ExecutionSchema || artifact.Producer != ExecutionProducer || artifact.CaseID != caseID || artifact.OperationID != MetaOperationID || artifact.Path != "execution/"+caseID+".json" || artifact.GrantDigest != grant.GrantDigest || artifact.InputDigest != source.SemanticDigest || artifact.OutputCanonical == "" || artifact.OutputDigest != digestBytes([]byte(artifact.OutputCanonical)) {
		return false
	}
	copy := artifact
	copy.ArtifactDigest = ""
	return artifact.ArtifactDigest != "" && artifact.ArtifactDigest == digestValue(copy)
}

func readExecutionArtifact(root string, artifact ExecutionArtifact) (ExecutionArtifact, error) {
	path := filepath.Join(root, filepath.FromSlash(artifact.Path))
	raw, err := os.ReadFile(path)
	if err != nil {
		return ExecutionArtifact{}, err
	}
	var loaded ExecutionArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loaded); err != nil {
		return ExecutionArtifact{}, err
	}
	return loaded, nil
}
