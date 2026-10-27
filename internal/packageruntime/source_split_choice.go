package packageruntime

import "fmt"

// SourceSplitChoice ranks physical cuts only. Consumer and semantic admission
// still belong to SplitSharedSource and the host's publication checks.
type SourceSplitChoice struct {
	Objective        string               `json:"objective"`
	State            string               `json:"state"`
	LineLimit        int                  `json:"line_limit"`
	InputLines       int                  `json:"input_lines"`
	InputSHA256      string               `json:"input_sha256"`
	InputImageDigest string               `json:"input_image_digest"`
	Candidates       int                  `json:"candidates"`
	Improving        int                  `json:"improving_candidates"`
	WithinLimit      int                  `json:"within_limit_candidates"`
	LimitSatisfied   bool                 `json:"limit_satisfied"`
	Boundary         *SourceSplitBoundary `json:"boundary,omitempty"`
}

// ChooseSourceSplit minimizes the largest output's physical line count, then
// the stable declaration ID. It may select a useful intermediate split while
// reporting that the requested limit is still unmet. It never splits a body.
func ChooseSourceSplit(manifest Manifest, packagePath, filename string, limit int) (SourceSplitChoice, error) {
	if limit <= 0 {
		return SourceSplitChoice{}, fmt.Errorf("split line limit must be positive")
	}
	observed, err := InspectSourceSplit(manifest, packagePath, filename)
	if err != nil {
		return SourceSplitChoice{}, err
	}
	return chooseSourceSplit(observed, limit), nil
}

func chooseSourceSplit(observed SourceSplitBoundaries, limit int) SourceSplitChoice {
	choice := SourceSplitChoice{Objective: "minimize-maximum-physical-lines/v1", State: "NO_IMPROVING_CUT",
		LineLimit: limit, InputLines: observed.Input.Lines, InputSHA256: observed.InputSHA256,
		InputImageDigest: observed.InputImageDigest, Candidates: len(observed.Candidates)}
	for _, candidate := range observed.Candidates {
		peak := max(candidate.Left.Lines, candidate.Right.Lines)
		if peak >= observed.Input.Lines {
			continue
		}
		choice.Improving++
		if peak <= limit {
			choice.WithinLimit++
		}
		best := choice.Boundary
		if best == nil || peak < max(best.Left.Lines, best.Right.Lines) ||
			(peak == max(best.Left.Lines, best.Right.Lines) && candidate.BeforeID < best.BeforeID) {
			selected := candidate
			choice.Boundary = &selected
		}
	}
	if observed.Input.Lines <= limit {
		choice.State, choice.LimitSatisfied, choice.Boundary = "ALREADY_WITHIN_LIMIT", true, nil
	} else if choice.Boundary != nil {
		choice.State = "SELECTED"
		choice.LimitSatisfied = max(choice.Boundary.Left.Lines, choice.Boundary.Right.Lines) <= limit
	}
	return choice
}
