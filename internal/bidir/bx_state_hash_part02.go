package bidir

import (
	"encoding/json"
	"fmt"
)

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
