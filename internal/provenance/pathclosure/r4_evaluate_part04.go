package pathclosure

import (
	"bytes"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func evaluateR4Path(path R4Path, records map[semantic.ID]R4Record, receipts map[semantic.ID]R4Receipt) (Status, string, string) {
	if len(path.RecordIDs) == 0 || len(path.RecordIDs) != len(path.RecordBytes) {
		return FAIL_CLOSED, CodeInvalidPath, "path record IDs and canonical record bytes are not a non-empty equal-length sequence"
	}
	if err := invalidR4ID(path.StartID, "path start ID"); err != nil {
		return FAIL_CLOSED, CodeInvalidPath, err.Error()
	}
	if err := invalidR4ID(path.EndID, "path end ID"); err != nil {
		return FAIL_CLOSED, CodeInvalidPath, err.Error()
	}
	var previous R4Record
	for index, recordID := range path.RecordIDs {
		record, exists := records[recordID]
		if !exists {
			return UNKNOWN, CodeMissingRecord, "missing record " + recordID.String()
		}
		provided, err := decodeCanonicalR4Record([]byte(path.RecordBytes[index]))
		if err != nil {
			return FAIL_CLOSED, CodeInvalidPath, fmt.Sprintf("record %s bytes: %v", recordID, err)
		}
		if provided.ID != record.ID {
			return FAIL_CLOSED, CodeInvalidPath, "record ID does not match ordered record ID"
		}
		actualBytes, _ := record.CanonicalRecordBytes()
		providedBytes, _ := provided.CanonicalRecordBytes()
		if !bytes.Equal(actualBytes, providedBytes) || !bytes.Equal(actualBytes, []byte(path.RecordBytes[index])) {
			return FAIL_CLOSED, CodeInvalidPath, "canonical record bytes do not match the record identity and fields"
		}
		if index == 0 {
			if record.PredecessorID != "" {
				return FAIL_CLOSED, CodeInvalidPath, "root record has a predecessor"
			}
			if record.SubjectID != path.StartID {
				return FAIL_CLOSED, CodeInvalidPath, "path start does not match first record subject"
			}
		} else {
			if record.PredecessorID != previous.ID {
				return FAIL_CLOSED, CodeInvalidPath, "record predecessor does not match ordered predecessor"
			}
			if previous.ObjectID != record.SubjectID {
				return FAIL_CLOSED, CodeInvalidPath, "ordered edge subject/object endpoints do not join"
			}
		}
		if index == len(path.RecordIDs)-1 && record.ObjectID != path.EndID {
			return FAIL_CLOSED, CodeInvalidPath, "path end does not match last record object"
		}
		if status, code, reason := validateR4Binding(record, receipts); status != PASS {
			return status, code, reason
		}
		previous = record
	}
	return PASS, CodeR4ProofValid, "path covered"
}
