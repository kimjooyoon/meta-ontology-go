package artifactemit

import "encoding/json"

func symbolicReaderRequestResultDigest(value SymbolicReaderRequestResult) (string, error) {
	value.Digest = ""
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return symbolicReaderBytesDigest(payload), nil
}
