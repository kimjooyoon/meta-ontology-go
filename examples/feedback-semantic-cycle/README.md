# Executable predecessor semantic cycle

This use case set closes the gap between selecting a predecessor artifact and
using its meaning in the next self-improvement cycle.

The CI collector preserves the exact resolution-receipt bytes as base64 and
binds their SHA-256 digest to the selected artifact identity. The read-only
semantic witness decodes only the unique selected payload, replays
`feedbackstate.Evaluate`, and writes its receipt outside the repository.

The cases cover explicit fixed points, executable improvement, monotone
resolution descent, terminal fail-closed behavior, unknown decisions, false
fixed points, payload mismatch, and write effects. They make no novelty claim;
the useful property is the uncommon integration of causal selection, exact
bytes, semantic descent, and executable CI evidence.
