package toolchainrelease

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func validateJointFillRejectionSmoke(raw, source, feedback, evaluation []byte, budget int, replay bool, selected string) (string, error) {
	var r jointSmokeOutput
	if err := json.Unmarshal(raw, &r); err != nil {
		return "", err
	}
	construction, err := jointSmokeCases(feedback, 1)
	if err != nil {
		return "", err
	}
	cases, err := jointSmokeCases(evaluation, 4)
	if err != nil {
		return "", err
	}
	if err := validateJointFillRejectionIdentity(r, source, budget, replay, selected); err != nil {
		return "", err
	}
	if err := validateJointFillRejectionInitial(r); err != nil {
		return "", err
	}
	if err := validateJointFillRejectionAttempts(r, construction); err != nil {
		return "", err
	}
	return validateJointFillRejectionFinal(r, cases, budget)
}

func validateJointFillRejectionFinal(r jointSmokeOutput, cases []jointSmokeCase, budget int) (string, error) {
	c, e := r.Construction, r.Evaluation
	index := 4
	if budget == 3 {
		index = 0
	}
	if c.Attempts[index].Runtime.SHA != c.Selected.SHA || e.Runtime.SHA != c.Selected.SHA ||
		c.Attempts[index].FillCandidates[0].SelectedSHA != fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(c.Source))) {
		return "", fmt.Errorf("rejected-fill selected source differs from execution")
	}
	var actual []json.RawMessage
	if budget == 3 {
		actual = []json.RawMessage{[]byte("2"), []byte("-12"), []byte("5"), []byte("9007199254740994")}
	}
	if err := validateJointFillRuntime(e.Runtime, cases, actual); err != nil {
		return "", err
	}
	return c.Selected.SHA, nil
}

func validateJointFillRejectionInitial(r jointSmokeOutput) error {
	return validateJointFillInitialIDs(r, []string{"late_unbounded", "early_wrong_cap", "bounded"})
}

func validateJointFillInitialIDs(r jointSmokeOutput, ids []string) error {
	if len(r.Construction.Initial.Preparations) != 1 {
		return fmt.Errorf("rejected-fill initial count differs")
	}
	i := r.Construction.Initial.Preparations[0].Generation.Report.Fill
	if i == nil || i.Selected != "late_unbounded" || i.PlanSHA == "" || !jointSmokeInt(i.Calls, 0) ||
		len(i.Scores) != len(ids) || len(i.Rejected) != 2 {
		return fmt.Errorf("rejected-fill initial score/rejection partition differs")
	}
	for index, score := range i.Scores {
		if score.ID != ids[index] ||
			!jointSmokeBool(score.Typed, true) || !jointSmokeInt(score.Passed, 1) || !jointSmokeInt(score.Total, 1) {
			return fmt.Errorf("rejected-fill initial score differs")
		}
	}
	for index, rejection := range i.Rejected {
		if err := validateJointFillRejection(rejection, index+1); err != nil {
			return err
		}
		attempt := r.Construction.Attempts[index+1]
		if attempt.Rejection == nil || attempt.Rejection.Reason != rejection.Reason {
			return fmt.Errorf("initial and whole-program rejection reasons differ")
		}
	}
	return nil
}
