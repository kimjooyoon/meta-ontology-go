package analysisprovenance

import "testing"

func TestSubjectDigestBindsSourceNameAndDigest(t *testing.T) {
	first := SubjectDigest("examples/order.gooo", "source-digest")
	if first == "" || first != SubjectDigest("examples/order.gooo", "source-digest") {
		t.Fatalf("subject digest is not deterministic: %q", first)
	}
	if first == SubjectDigest("examples/invoice.gooo", "source-digest") {
		t.Fatal("subject digest ignored source name")
	}
	if first == SubjectDigest("examples/order.gooo", "changed-source-digest") {
		t.Fatal("subject digest ignored source digest")
	}
}
