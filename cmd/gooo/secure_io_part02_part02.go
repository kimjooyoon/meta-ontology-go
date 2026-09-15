package main

func writeGeneratedOutput(path string, data []byte) error {
	if int64(len(data)) > maxOutputBytes {
		return outputLimitError(maxOutputBytes)
	}
	same, err := digestMatches(path, data)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	if err := validateOutputTarget(path); err != nil {
		return err
	}
	return writeAtomicFile(path, data)
}
