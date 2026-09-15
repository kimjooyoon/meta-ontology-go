package operationconformance

import "bytes"

func observeHeader(evidence SplitGoEvidence) Decision {
	expected, err := sourceHeader(evidence.Source)
	if err != nil || len(evidence.Candidates) == 0 {
		return DecisionFail
	}
	for _, candidate := range evidence.Candidates {
		actual, headerErr := sourceHeader(candidate)
		if headerErr != nil || !bytes.Equal(actual, expected) {
			return DecisionFail
		}
	}
	return DecisionPass
}

func sourceHeader(file FileEvidence) ([]byte, error) {
	fset, parsed, err := parseEvidence(file)
	if err != nil {
		return nil, err
	}
	offset := fset.Position(parsed.Package).Offset
	return append([]byte(nil), file.Data[:offset]...), nil
}
