# Symbolic invocation user use case

This experiment keeps compiler production separate from observed user value. A producer job emits a symbolic invocation schema with one invocation generated from Gooo activity and entity declarations. A consumer job downloads that exact-head artifact, compares the generated invocation with an independent golden, and asks the pinned external validator to accept it and reject one counterexample.

The fixed metric denominator is 8 indicators:

- 2 outcomes: 2 externally observed user decisions and 1 compiler-generated invocation
- 3 drivers: 1 accepted instance, 1 rejected instance, and 1 independent golden match
- 3 guardrails: 1 deterministic producer replay, 0 repository writes, and 0 mutation authorities

The Munchausen proof-choice distribution is fixed at 4 `FOUNDATION`, 3 `COHERENCE`, and 1 `REGRESSION`. Reader views expose 5 indicators to a user, 6 to a tool author, and all 8 to a governor.

Runner wall time and RSS are copied as observations without targets. They are labeled `RUNNER_SCOPED_NONDETERMINISTIC`, have no replay authority, and do not support a general performance claim. The report gives no readiness or promotion credit.

Unknown top-level decisions fail closed at `LOWER_RESOLUTION`. Known identity, coordinate, effect, or digest-link mismatches fail closed at `INVARIANT_ONLY`.
