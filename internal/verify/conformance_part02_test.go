package verify

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

func verificationDocument() bidir.Document {
	return bidir.Document{
		Package: "billing", Namespace: "billing",
		Declarations: []bidir.Declaration{
			{Kind: bidir.EntityKind, ID: "billing://entity/order", Name: "Order"},
			{Kind: bidir.EntityKind, ID: "billing://entity/payment", Name: "Payment"},
			{Kind: bidir.EntityKind, ID: "billing://entity/audit", Name: "Audit"},
			{Kind: bidir.EntityKind, ID: "billing://entity/unrelated", Name: "Unrelated"},
			{Kind: bidir.ActivityKind, Name: "PayOrder", Inputs: []bidir.Reference{{Name: "Order"}}, Outputs: []bidir.Reference{{Name: "Payment"}}},
			{Kind: bidir.ActivityKind, Name: "AuditPayment", Inputs: []bidir.Reference{{Name: "Payment"}}, Outputs: []bidir.Reference{{Name: "Audit"}}},
		},
	}
}
