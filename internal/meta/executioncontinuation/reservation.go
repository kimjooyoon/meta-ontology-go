package executioncontinuation

import (
	"io/fs"
	"reflect"
	"strings"

	input "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementvaluewitnessinput"
)

const ReservationScope = "SERIALIZED_CONSUMER_WITH_RETAINED_ARTIFACT_HISTORY"

type Reservation struct {
	Schema             string `json:"schema"`
	GrantID            string `json:"grant_id"`
	RequestDigest      string `json:"request_digest"`
	GrantReceiptDigest string `json:"grant_receipt_digest"`
	MetadataDigest     string `json:"metadata_digest"`
	ConsumerRunID      int64  `json:"consumer_run_id"`
	ConsumerAttempt    int    `json:"consumer_run_attempt"`
	SubjectSHA         string `json:"subject_sha"`
	Status             string `json:"status"`
	ConsumedUses       int    `json:"consumed_uses"`
	RemainingUses      int    `json:"remaining_uses"`
	ExecutionCount     int    `json:"execution_count"`
	OneUseEnforced     bool   `json:"one_use_enforced"`
	EnforcementScope   string `json:"enforcement_scope"`
	Digest             string `json:"digest"`
}

func ReservationName(request string) string {
	return "gooo-consumed-" + strings.ReplaceAll(request, ":", "-")
}

func validatePrepared(root fs.FS, p Prepared) (*input.ExecutionInput, error) {
	in, err := ValidatePair(root, p.Pair, p.Metadata)
	if err != nil {
		return nil, err
	}
	if p.SourceBytes != in.Source.Bytes || Digest([]byte(p.SourceBytes)) != p.Metadata.SubjectFileDigest || in.Source.Digest != p.Metadata.SubjectFileDigest {
		return nil, refuted("SUBJECT_SOURCE_MISMATCH")
	}
	return in, nil
}

func Reserve(root fs.FS, p Prepared) (Reservation, error) {
	if _, err := validatePrepared(root, p); err != nil {
		return Reservation{}, err
	}
	return reservationFor(p), nil
}

func reservationFor(p Prepared) Reservation {
	receipt := p.Pair.Grant.Resolution.Receipt
	r := Reservation{Schema: ConsumptionSchema + "/reservation", GrantID: receipt.GrantID, RequestDigest: p.Pair.Grant.Request.Digest,
		GrantReceiptDigest: receipt.Digest, MetadataDigest: p.Metadata.Digest, ConsumerRunID: p.Metadata.ConsumerRunID,
		ConsumerAttempt: p.Metadata.ConsumerAttempt, SubjectSHA: p.Metadata.Source.SubjectSHA, Status: "CONSUMPTION_RESERVED",
		ConsumedUses: 1, RemainingUses: 0, ExecutionCount: 0, OneUseEnforced: true, EnforcementScope: ReservationScope}
	r.Digest = digestValue(r)
	return r
}

func validateReservation(p Prepared, r Reservation) error {
	if !reflect.DeepEqual(r, reservationFor(p)) {
		return refuted("RESERVATION_BINDING_MISMATCH")
	}
	return nil
}
