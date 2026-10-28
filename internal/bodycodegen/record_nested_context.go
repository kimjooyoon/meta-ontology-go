package bodycodegen

// Existing model features do not describe nominal child layouts. Keep their
// versioned encodings intact and retain deterministic source-owned selection.
func recordPlanHasNominalFields(plan recordAssemblyPlan) bool {
	for _, record := range plan.body.records {
		for _, field := range record.Fields {
			switch field.TypeID {
			case "urn:gooo:type:string", "urn:gooo:type:boolean", "urn:gooo:type:integer":
			default:
				return true
			}
		}
	}
	return false
}

func declinedNominalRecordContext(plan recordAssemblyPlan, version string) *RecordOrdinalContext {
	r := recordOrdinalContext(plan.choices)
	r.Status, r.Reason = "DECLINED_TO_DETERMINISTIC", "NOMINAL_RECORD_LAYOUT_FEATURE_UNAVAILABLE"
	r.FeatureVersion, r.Text, r.SHA256 = version, "", ""
	r.Scope = "source-owned nominal record layouts and finite selection; existing model feature contracts do not encode this layout"
	return r
}
