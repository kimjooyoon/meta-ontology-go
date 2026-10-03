package bodycodegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	decision "github.com/kimjooyoon/gooo-decision-runtime"
)

// The retained stream and decoded receipts require explicit integer probes too.
// encoding/json would otherwise convert an array's null entry to int64(0).
func (o *PathObservationOptions) UnmarshalJSON(raw []byte) error {
	if len(raw) > 4096 || decision.RejectDuplicateJSONKeys(raw) != nil {
		return fmt.Errorf("path observation requires strict JSON of at most 4 KiB")
	}
	type plain PathObservationOptions
	var decoded plain
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&decoded); err != nil {
		return err
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("trailing path observation data")
	}
	var probes struct {
		Inputs []*int64 `json:"inputs"`
	}
	if err := json.Unmarshal(raw, &probes); err != nil {
		return err
	}
	for _, input := range probes.Inputs {
		if input == nil {
			return fmt.Errorf("path probe input requires an explicit integer")
		}
	}
	owned, err := copyPathObservation((*PathObservationOptions)(&decoded))
	if err != nil {
		return err
	}
	*o = *owned
	return nil
}
