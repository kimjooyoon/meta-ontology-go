package toolchainrelease

import (
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func validateTypedPredictionCounter(raw []byte) error {
	var r struct {
		Construction struct {
			Initial struct {
				Preparations []struct {
					Generation struct{ Report json.RawMessage }
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	for _, p := range r.Construction.Initial.Preparations {
		var report struct {
			Path *struct {
				Search struct {
					Selection struct {
						Calls *int `json:"local_model_predictions"`
					}
				}
			} `json:"body_paths"`
		}
		if err := json.Unmarshal(p.Generation.Report, &report); err != nil {
			return err
		}
		if report.Path != nil && !jointSmokeInt(report.Path.Search.Selection.Calls, 0) {
			return fmt.Errorf("typed initial prediction counter is missing or nonzero")
		}
	}
	return nil
}

func validateTypedModelBinding(initial bodyexecution.Composition, contextSHA string) error {
	if len(initial.Preparations) != 2 || contextSHA == "" {
		return fmt.Errorf("typed model preparation count or preflight digest differs")
	}
	record, typed := false, false
	for _, preparation := range initial.Preparations {
		r := preparation.Generation.Report
		switch r.ActivityID {
		case "callerpaths://activity/choose":
			p := r.BodyPaths
			if typed || p == nil || p.ModelRetention == nil || p.ModelContext == nil ||
				p.ModelContext.Status != "DECLINED_TO_DETERMINISTIC" ||
				p.ModelContext.Reason != "THREE_DECISION_COUNT_UNSUPPORTED" ||
				p.Search.Selection.ModelCalls != 0 {
				return fmt.Errorf("typed representation decline or zero predictions differs")
			}
			if err := validateScalarModelIdentity(*p.ModelRetention); err != nil {
				return err
			}
			typed = true
		case "callerpaths://activity/pick":
			p := r.RecordAssembly
			if record || p == nil || p.Model == nil || p.Context == nil || p.Prediction == nil ||
				p.ModelCalls != 1 || p.Context.SHA256 != contextSHA ||
				p.Context.SHA256 != scalarPreflightDigest([]byte(p.Context.Text)) {
				return fmt.Errorf("record prediction, identity or inspected input digest differs")
			}
			if err := validateScalarModelIdentity(*p.Model); err != nil {
				return err
			}
			record = true
		default:
			return fmt.Errorf("unexpected prepared typed example activity")
		}
	}
	if !typed || !record {
		return fmt.Errorf("typed example lost one preparation")
	}
	return nil
}
