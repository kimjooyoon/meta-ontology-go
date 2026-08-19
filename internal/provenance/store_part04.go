package provenance

import (
	"bytes"
)

func canonicalJSONL(records []Evidence) ([]byte, error) {
	var output bytes.Buffer
	for _, record := range records {
		line, err := marshalEvidence(record, true)
		if err != nil {
			return nil, err
		}
		output.Write(line)
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

// ReadAll is a convenience alias for adapters that prefer explicit wording.
func (s *Store) ReadAll(options ReadOptions) (Snapshot, error) { return s.Read(options) }
