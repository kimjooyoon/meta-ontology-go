package metarecognition

import (
	"encoding/json"
	"fmt"
	"io"
)

func expectDelimiter(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != expected {
		return fmt.Errorf("expected JSON delimiter %q", expected)
	}
	return nil
}
func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("replay JSON has trailing value")
		}
		return err
	}
	return nil
}
