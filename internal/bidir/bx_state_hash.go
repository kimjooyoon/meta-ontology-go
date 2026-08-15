package bidir

import (
	"encoding/json"
	"fmt"
	"strings"
)

func lstatDigest(stat BXLStat) string {
	return digest(fmt.Sprintf("%s|%d|%d|%t", stat.Path, stat.Size, stat.Mode, stat.Exists))
}

func stateEvidence(model Model, document Document, region Locality, snapshot BXFileSnapshot) BXStateEvidence {
	return BXStateEvidence{
		Semantic: SemanticFingerprint(model),
		Source:   documentDigest(document),
		Region:   localityDigest(region),
		Slot:     slotDigest(document),
		Bytes:    digest(string(snapshot.Bytes)),
		LStat:    lstatDigest(snapshot.LStat),
	}
}

func deferredStateEvidence(model Model, document Document, region Locality) BXStateEvidence {
	return BXStateEvidence{
		Semantic: SemanticFingerprint(model),
		Source:   documentDigest(document),
		Region:   localityDigest(region),
		Slot:     slotDigest(document),
	}
}

func localityDigest(locality Locality) string {
	return digest(localityCanonical(locality))
}

func localityCanonical(locality Locality) string {
	var builder strings.Builder
	for _, id := range locality.Touched {
		writePart(&builder, string(id))
	}
	builder.WriteByte('|')
	for _, id := range locality.Affected {
		writePart(&builder, string(id))
	}
	return builder.String()
}

func slotDigest(document Document) string {
	var builder strings.Builder
	for _, declaration := range document.Declarations {
		writeSlots(&builder, declaration.ID, "input", declaration.Inputs)
		writeSlots(&builder, declaration.ID, "output", declaration.Outputs)
	}
	return digest(builder.String())
}

func writeSlots(builder *strings.Builder, declaration ID, direction string, references []Reference) {
	for index, reference := range references {
		fmt.Fprintf(builder, "%d|", index)
		writePart(builder, string(declaration))
		writePart(builder, direction)
		writePart(builder, string(reference.ID))
		writePart(builder, reference.Name)
		writePart(builder, reference.Namespace)
		writeSpan(builder, reference.Span)
	}
}

func observationMatches(observation BXWriteObservation, before, after Document) error {
	if !observation.Observed {
		return fmt.Errorf("write observer did not report an observation")
	}
	if err := snapshotMatches(observation.Before, before); err != nil {
		return fmt.Errorf("before snapshot: %w", err)
	}
	if err := snapshotMatches(observation.After, after); err != nil {
		return fmt.Errorf("after snapshot: %w", err)
	}
	return nil
}

func snapshotMatches(snapshot BXFileSnapshot, document Document) error {
	want := documentSourceBytes(document)
	if string(snapshot.Bytes) != string(want) {
		return fmt.Errorf("observed bytes do not match source document")
	}
	if !snapshot.LStat.Exists || snapshot.LStat.Path == "" || snapshot.LStat.Mode == 0 || snapshot.LStat.Size != int64(len(snapshot.Bytes)) {
		return fmt.Errorf("observed lstat is incomplete or inconsistent")
	}
	return nil
}

func documentSourceBytes(document Document) []byte {
	return []byte(documentCanonical(document))
}

func canonicalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
