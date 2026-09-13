package policycompilation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

func decodeStrictJSON(data []byte, target any) error {
	if err := rejectDuplicateObjectKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("json document contains trailing data")
	}
	return nil
}

// rejectDuplicateObjectKeys performs a token-level preflight so json.Decoder
// cannot silently retain only the last value for a repeated object key. Token
// values are decoded strings, so escaped spellings of the same key share one
// identity, while each object receives its own key set.
func rejectDuplicateObjectKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	return scanJSONValue(decoder, "$")
}

func scanJSONValue(decoder *json.Decoder, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delimiter {
	case '{':
		seen := make(map[string]int64)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("json object key is not a string")
			}
			keyOffset := decoder.InputOffset()
			if previousOffset, exists := seen[key]; exists {
				return duplicateObjectKeyError(path, key, keyOffset, previousOffset)
			}
			seen[key] = keyOffset
			if err := scanJSONValue(decoder, jsonKeyPath(path, key)); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return errors.New("json object did not close with an object delimiter")
		}
	case '[':
		for index := 0; decoder.More(); index++ {
			if err := scanJSONValue(decoder, jsonArrayPath(path, index)); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return errors.New("json array did not close with an array delimiter")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func jsonKeyPath(path, key string) string {
	return path + "[" + strconv.Quote(key) + "]"
}

func jsonArrayPath(path string, index int) string {
	return path + "[" + strconv.Itoa(index) + "]"
}

func duplicateObjectKeyError(path, key string, keyOffset, previousOffset int64) error {
	return errors.New("duplicate JSON object key " + strconv.Quote(key) + " at " + jsonKeyPath(path, key) +
		" (byte offset " + strconv.FormatInt(keyOffset, 10) + "; first occurrence at byte offset " + strconv.FormatInt(previousOffset, 10) + ")")
}
