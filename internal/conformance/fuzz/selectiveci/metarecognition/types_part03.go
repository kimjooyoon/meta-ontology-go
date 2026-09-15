package metarecognition

type Case struct {
	ID       string         `json:"id"`
	Expected Expected       `json:"expected"`
	Baseline BaselineConfig `json:"baseline"`
}
type Work struct {
	// Units is the deterministic command work count; provenance dimensions are
	// reported separately and are never folded into a weighted scalar.
	Units       int `json:"work_units"`
	Selected    int `json:"selected_commands"`
	Full        int `json:"full_commands"`
	ProvRecords int `json:"prov_records"`
	ProvPaths   int `json:"prov_paths"`
}
type Outcome struct {
	State        State    `json:"state"`
	Reason       Reason   `json:"reason"`
	LocalizedIDs []string `json:"localized_ids"`
	Work         Work     `json:"work"`
}
type ComparisonCase struct {
	ID                    string   `json:"id"`
	Expected              Expected `json:"expected"`
	Gooo                  Outcome  `json:"gooo"`
	Baseline              Outcome  `json:"baseline"`
	ExactOutcomeVector    bool     `json:"exact_outcome_vector"`
	ExactReasonVector     bool     `json:"exact_reason_localization_vector"`
	GoooFalsePass         bool     `json:"gooo_false_pass"`
	GoooFalseNegative     bool     `json:"gooo_false_negative"`
	BaselineFalsePass     bool     `json:"baseline_false_pass"`
	BaselineFalseNegative bool     `json:"baseline_false_negative"`
}
type Ratio struct {
	Selected int  `json:"selected"`
	Full     int  `json:"full"`
	Known    bool `json:"known"`
}
type Summary struct {
	CaseCount                     int   `json:"case_count"`
	ExactOutcomeVector            bool  `json:"exact_outcome_vector"`
	ExactReasonLocalizationVector bool  `json:"exact_reason_localization_vector"`
	GoooFalsePasses               int   `json:"gooo_false_passes"`
	GoooFalseNegatives            int   `json:"gooo_false_negatives"`
	BaselineFalsePasses           int   `json:"baseline_false_passes"`
	BaselineFalseNegatives        int   `json:"baseline_false_negatives"`
	GoooWorkUnits                 int   `json:"gooo_work_units"`
	BaselineWorkUnits             int   `json:"baseline_work_units"`
	GoooRatio                     Ratio `json:"gooo_selected_full_ratio"`
	BaselineRatio                 Ratio `json:"baseline_selected_full_ratio"`
	GoooProvRecords               int   `json:"gooo_prov_records"`
	BaselineProvRecords           int   `json:"baseline_prov_records"`
	GoooProvPaths                 int   `json:"gooo_prov_paths"`
	BaselineProvPaths             int   `json:"baseline_prov_paths"`
}
type Manifest struct {
	Schema  string           `json:"schema"`
	Finding Finding          `json:"finding"`
	Cases   []ComparisonCase `json:"cases"`
	Summary Summary          `json:"summary"`
}

func (s State) Valid() bool {
	return s == ClosedSound || s == FailClosedUnsound || s == UnknownFullSuiteRequired
}
