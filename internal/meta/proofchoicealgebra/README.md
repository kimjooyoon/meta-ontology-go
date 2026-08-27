# Proof-choice algebra

This experiment treats the choice of proof ground as a meta-value. A source
claim can state its proposition, prior state, subject, and required evidence
capability. It cannot state an observation, route result, expected value,
metric slot, or provenance.

The producer lowers the raw Gooo file and creates evidence from that
canonical IR:

* FOUNDATION hashes the lowered node identity, semantic origin, and subject
  binding.
* COHERENCE computes two independent subject/proposition projections and
  compares them.
* REGRESSION compares a separately generated canonical artifact with the
  current artifact, checking both bytes and semantic digest.

The selector accepts exactly one observed route. Zero evidence yields
LOWER_RESOLUTION with observation state UNKNOWN or INSUFFICIENT; more than
one exact route yields FAIL_CLOSED while the subject resolution remains
EXACT. An append-only transition preserves OPEN, discharges an exact
claim, keeps an unknown claim open, and refutes a contradiction. The
ALL_ROUTES composition operator is a separate meta-operation: it discharges
only when the three distinct route witnesses are all exact.

The design adopts two principles from proof assistants rather than copying
their type systems. Rocq's reference manual makes checked proof terms and the
kernel's conversion/checking boundary authoritative:
https://rocq-prover.org/doc/master/refman/index.html
Lean's reference documents elaboration as producing kernel-checkable terms:
https://lean-lang.org/doc/reference/latest/
The adopted principle is that a conclusion is accepted only after an
independent checker reconstructs its semantic evidence. The rejected
principle is using a source tag or a declared type as the proof result: a
FOUNDATION/COHERENCE/REGRESSION label is a capability request here, never
an observation.

The judge is deliberately a separate consumer. It imports neither producer
package nor receipt-only evidence; it parses and lowers raw Gooo, recomputes
raw and canonical semantic digests, reconstructs evidence, and verifies every
receipt field. CI also records separate case, route, claim, intervention,
source-reconstruction, repository-write, and producer-import denominators.
