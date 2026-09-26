package generation

import "errors"

func AttachJEVActionObservationPart01(receipt JEVDecisionBoundaryPart01, actionObservationDigest string) (JEVDecisionBoundaryPart01, error) {
    if !receipt.ValidPart01() {
        return JEVDecisionBoundaryPart01{}, errors.New("cannot attach observation to invalid JEV receipt")
    }
    if !validDigestJEVPart01(actionObservationDigest) {
        return JEVDecisionBoundaryPart01{}, errors.New("invalid JEV action observation digest")
    }
    receipt.ActionObservationDigest = actionObservationDigest
    receipt.EvidencePrefixDigest = digestJEVPart01(receipt.evidencePrefixPart01())
    receipt.DecisionDigest = digestJEVPart01(receipt.decisionViewPart01())
    return receipt, nil
}