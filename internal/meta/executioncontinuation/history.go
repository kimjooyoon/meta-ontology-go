package executioncontinuation

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/candidatedelivery"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
)

func (c Client) validReservation(ctx context.Context, run Run, a candidatedelivery.Artifact, e Envelope) (bool, error) {
	if run.Repository.Name != c.Repository || run.Path != policy.WorkflowPath("consumer") || !validSHA(run.Head) || a.Expired {
		return false, unknown("RESERVATION_ORIGIN_UNKNOWN")
	}
	files, err := c.readArtifact(ctx, run, a.Name, a.ID, a.Digest, []string{"reservation.json"})
	if err != nil {
		return false, err
	}
	var r Reservation
	if decode(files["reservation.json"], &r) != nil {
		return false, unknown("RESERVATION_RECORD_UNKNOWN")
	}
	unsigned := r
	unsigned.Digest = ""
	if r.Schema != ConsumptionSchema+"/reservation" || r.Digest != digestValue(unsigned) || r.ConsumerRunID != run.ID || r.ConsumerAttempt != 1 || r.RequestDigest != e.RequestDigest || r.SubjectSHA != e.SubjectSHA || r.Status != "CONSUMPTION_RESERVED" || r.ConsumedUses != 1 || r.RemainingUses != 0 || !r.OneUseEnforced || r.EnforcementScope != ReservationScope {
		return false, unknown("RESERVATION_RECORD_UNBOUND")
	}
	return true, nil
}
