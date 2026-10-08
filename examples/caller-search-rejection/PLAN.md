# Continue after a rejected source expression

Frozen before implementation and candidate execution, 2026-10-09 KST.
The clean de5aba5b compiler stops on the second declared expression (`0`) with
`invalid operation: division by zero`. Its first expression returns 1, while
the unchanged caller expects -1. The third expression (`-input`) is still
inside the source's five-candidate prefix.

1. Retain deterministic source-search evaluation failures as unscored candidate
   rejections. Preserve candidate identity, source/plan digests and the original
   error. Never report an unexecuted caller as a scored zero or success.
2. Charge each rejection to the existing whole-program attempt budget. Preserve
   order and choose only materialized, natively executed programs. Cancellation,
   invalid requests, source reconstruction failures and native execution/toolchain
   failures remain terminal. This change covers local IR search evaluation only.
3. Version histories containing a rejection as joint-construction/v3; preserve
   v1 record-only and v2 search histories without rejections. Saved replay must
   rederive each rejected expression and original failure before accepting it.
4. Require three attempts: caller 1 (mismatch), rejected zero divisor (unscored),
   caller -1 (match). Evaluate the four fixed rows, retaining the local zero
   example and counting its overlap separately. Replay without new inference.
5. Check two-attempt partial results, changed rejection/source/selector histories,
   local evaluation errors and cancellation. Keep existing record/mixed regressions.
6. Teach the separate Gooo workbench reader about rejected attempts without
   treating them as scored caller results. Use the current unchanged own model
   on a mixed record/search program and retain actual outcomes and source identity.

This is one compiler behavior regression, not a training study. Candidate-native
panics, all-invalid initial local construction, multi-hole joint construction and
unseen-program model accuracy remain separate questions.
