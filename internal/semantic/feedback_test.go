package semantic

import "testing"

func TestFeedbackSchemaChangesSemanticIdentityAndRejectsUnknownPhase(t *testing.T) {
	binding := RuntimeBinding{
		Schema: RuntimeBindingSchema, ProducerActivity: "p://activity/a", ProducerPort: RuntimeOutputPort,
		ConsumerActivity: "p://activity/b", ConsumerPort: RuntimeInputPort, Entity: "p://entity/integer",
	}
	feedback := binding
	feedback.Schema = RuntimeFeedbackSchema
	if _, err := feedback.Normalized(); err != nil {
		t.Fatal(err)
	}
	if binding.SemanticCanonical() == feedback.SemanticCanonical() || binding.Key() == feedback.Key() {
		t.Fatal("execution phase is missing from semantic identity")
	}
	feedback.Schema = "gooo.runtime-feedback/future"
	if _, err := feedback.Normalized(); err == nil {
		t.Fatal("unknown feedback schema accepted")
	}
}
