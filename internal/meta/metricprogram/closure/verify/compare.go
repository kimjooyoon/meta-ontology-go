package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func compareReceipts(actual, expected receipt) error {
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return err
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return err
	}
	if !bytes.Equal(actualJSON, expectedJSON) {
		return fmt.Errorf("closure receipt does not match independent reconstruction")
	}
	return nil
}
