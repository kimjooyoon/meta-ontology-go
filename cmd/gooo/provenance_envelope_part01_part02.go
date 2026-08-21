package main

import (
	"encoding/json"
	"errors"
	"io"
)

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("evidence input must contain one JSON value")
		}
		return err
	}
	return nil
}
