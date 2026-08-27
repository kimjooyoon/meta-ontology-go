# Capability-scoped expansion

`main.gooo` is the one source used for every case. The witness compares an
exact allow request with an undeclared network request, then runs six bounded
denials and one UNKNOWN case against that same source digest. CI writes the
receipts outside the checkout and uploads them as an artifact.

The source makes four capability values observable through declaration
activities: file read of `source`, logical time read of `logical-clock`, the
`GOOO_EXPANSION_PROFILE` environment value, and a pinned network target. The
Go witness does not contact the network, read a wall clock, or grant ambient
authority. It only accepts the corresponding deterministic evidence values.

The independent judge re-derives the decision from raw receipt JSON and the
source bytes. It is deliberately separate from the producer evaluator so that
the producer cannot turn its own `ALLOW` into proof.
