package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateRelativeV2Envelopes(raw []byte, c typedSmokeConstruction, replay bool) error {
	var r struct {
		Construction struct {
			Attempts []struct{ Runtime nativeSmokeRuntime }
		}
		Evaluation struct{ Runtime nativeSmokeRuntime }
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	if len(r.Construction.Attempts) != 1 {
		return fmt.Errorf("V2 caller runtime missing")
	}
	for i, runtime := range []nativeSmokeRuntime{r.Construction.Attempts[0].Runtime, r.Evaluation.Runtime} {
		total, schema := 1, "gooo/body-composition-runtime/v2"
		if i == 1 {
			total = 8
			schema = "gooo/body-composition-runtime/v1"
			if replay {
				schema = "gooo/body-composition-runtime/v2"
			}
		}
		if err := validateRelativeV2Envelope(runtime, total, schema, c.Selected.GeneratedSHA256, c.Selected.SelectedSourceSHA256); err != nil {
			return err
		}
	}
	return nil
}
