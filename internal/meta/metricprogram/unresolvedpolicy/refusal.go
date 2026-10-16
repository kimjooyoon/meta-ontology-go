package unresolvedpolicy

import "fmt"

type InputDigests struct {
	StrategyPayloadDigest     string `json:"strategy_payload_digest"`
	VerificationPayloadDigest string `json:"verification_payload_digest"`
}

// OperationOrigin records the supplied request and declared source snapshot.
// Presence describes retained context; it does not authenticate that context.
type OperationOrigin struct {
	Repository         string           `json:"repository"`
	SubjectSHA         string           `json:"subject_sha"`
	IndicatorID        string           `json:"indicator_id"`
	RequestedOperation string           `json:"requested_operation"`
	Registry           ProjectionOrigin `json:"registry"`
	Inputs             InputDigests     `json:"inputs"`
}

func (origin OperationOrigin) recorded() bool {
	return origin.Repository != "" && origin.SubjectSHA != "" && origin.IndicatorID != "" &&
		origin.RequestedOperation != "" && origin.Registry.SourcePath != "" &&
		origin.Registry.SourceDigest != "" && origin.Registry.ProjectionPath != "" &&
		origin.Registry.ProjectionDigest != "" && origin.Inputs.StrategyPayloadDigest != "" &&
		origin.Inputs.VerificationPayloadDigest != ""
}

// OperationRefusal is diagnostic data. It never permits the requested execution.
type OperationRefusal struct {
	Schema                    string           `json:"schema"`
	Origin                    OperationOrigin  `json:"origin"`
	Policy                    ProjectionOrigin `json:"policy"`
	OriginRecorded            bool             `json:"origin_recorded"`
	Resolution                string           `json:"resolution"`
	Reason                    string           `json:"reason"`
	NextOperation             string           `json:"next_operation"`
	ReadinessCredit           bool             `json:"readiness_credit"`
	RepositoryWorkspaceWrites bool             `json:"repository_workspace_writes"`
	PromotionAuthorized       bool             `json:"promotion_authorized"`
}

func NewOperationRefusal(origin OperationOrigin) (*OperationRefusal, error) {
	if err := VerifyProjection(); err != nil {
		return nil, fmt.Errorf("operation origin policy projection: %w", err)
	}
	recorded := origin.recorded()
	return &OperationRefusal{
		Schema: "gooo/unresolved-source-operation/v1", Origin: origin, Policy: Origin(), OriginRecorded: recorded,
		Resolution: OperationResolution(false, recorded), Reason: OperationReason(false, recorded),
		NextOperation: OperationNextStep(false, recorded),
	}, nil
}

func (r *OperationRefusal) Error() string {
	return fmt.Sprintf("%s: operation %q from indicator %q", r.Reason, r.Origin.RequestedOperation, r.Origin.IndicatorID)
}
