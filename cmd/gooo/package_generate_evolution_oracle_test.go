package main

type libraryEvolutionObservation struct {
	Origin string   `json:"origin_ref"`
	Paths  []string `json:"import_paths"`
}

type libraryEvolutionAssessment struct {
	Status     string   `json:"status"`
	Missing    int      `json:"missing_count"`
	Undeclared int      `json:"undeclared_count"`
	Origin     string   `json:"observation_origin"`
	Observed   []string `json:"observed_paths"`
}

// Expected classifications are independent of both source revisions and the
// generator. The input population is fixed before either program is generated.
var libraryEvolutionExpectations = [2][3]libraryEvolutionAssessment{
	{{Status: "OBSERVED_EXACT"}, {Status: "UNDECLARED_PACKAGES", Undeclared: 1}, {Status: "ORIGIN_UNOBSERVED"}},
	{{Status: "MISSING_PACKAGES", Missing: 1}, {Status: "OBSERVED_EXACT"}, {Status: "ORIGIN_UNOBSERVED", Missing: 1}},
}

func libraryEvolutionInputs() []libraryEvolutionObservation {
	const module = "github.com/kimjooyoon/gooo-go-ports/examples/recorded-catalog-comparison"
	paths := []string{module, module + "/internal/kernel", module + "/internal/definitionkernel",
		module + "/internal/observationkernel", module + "/internal/availabilitykernel",
		module + "/internal/availabilityobservationkernel", module + "/internal/snapshotkernel", module + "/cmd/compare"}
	extended := append(append([]string(nil), paths...), module+"/internal/recoverykernel")
	return []libraryEvolutionObservation{{"retained:before", paths}, {"retained:after", extended}, {"", paths}}
}

// This exact consumer is compiled before and after regenerating rules. Only
// authored Gooo and its generated Go file change; no expectation is emitted.
const libraryEvolutionConsumer = `package main
import (
 "encoding/json"
 "os"
 "policyprobe/rules"
)
func main() {
 var inputs []rules.R_Observation
 if err := json.NewDecoder(os.Stdin).Decode(&inputs); err != nil { panic(err) }
 outputs := make([]rules.R_Assessment, len(inputs))
 for i, input := range inputs { outputs[i] = rules.Evaluate(input) }
 if err := json.NewEncoder(os.Stdout).Encode(outputs); err != nil { panic(err) }
}
`
