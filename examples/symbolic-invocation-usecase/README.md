# Symbolic invocation user use case

This experiment keeps compiler production separate from observed user value. A producer job emits a symbolic invocation schema. A consumer job downloads that exact-head artifact and independently asks the pinned external validator to accept one instance and reject one counterexample.

The fixed metric denominator is 6 indicators:

- 1 outcome: 2 externally observed user decisions
- 2 drivers: 1 accepted instance and 1 rejected instance
- 3 guardrails: 1 deterministic producer replay, 0 repository writes, and 0 mutation authorities

The Munchausen proof-choice distribution is fixed at 3 `FOUNDATION`, 2 `COHERENCE`, and 1 `REGRESSION`. Reader views expose 3 indicators to a user, 4 to a tool author, and all 6 to a governor.

Runner wall time and RSS are copied as observations without targets. They are labeled `RUNNER_SCOPED_NONDETERMINISTIC`, have no replay authority, and do not support a general performance claim. The report gives no readiness or promotion credit.

Unknown top-level decisions fail closed at `LOWER_RESOLUTION`. Known identity, coordinate, effect, or digest-link mismatches fail closed at `INVARIANT_ONLY`.
