package sourceexecution

import "testing"

func TestValidateSuccessBindsEventSubjectsToReceiptMetadata(t *testing.T) {
	request := Request{Filename: "billing.gooo", Source: fixtureSource, Entry: "PayOrder"}
	valid := Execute(request)
	if err := Validate(valid); err != nil {
		t.Fatalf("unchanged generated receipt: %v", err)
	}

	const wantError = "SOURCE_EXECUTION_EVENT_SUBJECT_BINDING_INVALID"
	cases := []struct {
		name   string
		mutate func(*Receipt)
	}{
		{
			name: "source event subject",
			mutate: func(receipt *Receipt) {
				receipt.Events[0].Subject = digestBytes([]byte("source-subject-variant"))
			},
		},
		{
			name: "semantic event subject",
			mutate: func(receipt *Receipt) {
				receipt.Events[1].Subject = digestBytes([]byte("semantic-subject-variant"))
			},
		},
		{
			name: "activity event subject",
			mutate: func(receipt *Receipt) {
				receipt.Events[2].Subject = "RefundOrder"
			},
		},
		{
			name: "output event subject",
			mutate: func(receipt *Receipt) {
				receipt.Events[3].Subject = "billing://entity/refund"
			},
		},
		{
			name: "source metadata",
			mutate: func(receipt *Receipt) {
				receipt.SourceDigest = digestBytes([]byte("source-metadata-variant"))
			},
		},
		{
			name: "semantic metadata",
			mutate: func(receipt *Receipt) {
				receipt.SemanticDigest = digestBytes([]byte("semantic-metadata-variant"))
			},
		},
		{
			name: "activity metadata",
			mutate: func(receipt *Receipt) {
				receipt.Entry.Activity = "RefundOrder"
			},
		},
		{
			name: "output metadata",
			mutate: func(receipt *Receipt) {
				receipt.Entry.Output.ID = "billing://entity/refund"
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := valid
			mutated.Events = append([]Event(nil), valid.Events...)
			test.mutate(&mutated)
			mutated = seal(mutated)
			if err := Validate(mutated); err == nil || err.Error() != wantError {
				t.Fatalf("mutated receipt error = %v, want %s", err, wantError)
			}
		})
	}
}
