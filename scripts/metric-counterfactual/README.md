# Metric counterfactual

This command proves that repository metric transitions can represent a non-zero
change without modifying the checked-out repository.

The generator interprets a sealed fixture manifest and mutation plan inside a
disposable temporary root. It records file-level Go and Gooo line counts,
directory-level direct and recursive counts, transformation receipts, deltas,
and six indicators.

The verifier is a separate package. It re-materializes the fixture, replays the
plan, recomputes every state and indicator, and rejects any unbound field.

The project root is intentionally exceptional: counts remain observed while
topology and a root README requirement remain not applicable.
