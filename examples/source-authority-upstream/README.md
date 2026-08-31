# Upstream source authority use cases

This directory documents the fixed v1 conformance denominator for the
read-only upstream source producer.

The producer reads line 1 of the gomacro README at immutable commit
`cf0d4bf32da393dbda97e3572f216731013ffa55`. The selected 77 bytes must
match the pinned SHA-256 digest before an exact snapshot is emitted.

The denominator contains exactly three cases:

1. Exact authority, line span, byte count, and digest produce
   `SATISFIED / EXACT / ALLOW`.
2. A digest mismatch produces `UNKNOWN / INVARIANT_ONLY / BLOCK`.
3. An authority scope mismatch produces `UNKNOWN / INVARIANT_ONLY / BLOCK`
   without crossing the fetch boundary.

Every case keeps repository writes and promotion credit at zero. The suite
proves this producer route only; it does not promote Gooo language readiness.

Live discovery pages such as `https://news.hada.io/` remain candidate inputs.
They cannot become accepted authority without an immutable source snapshot and
an independently pinned authority policy.
