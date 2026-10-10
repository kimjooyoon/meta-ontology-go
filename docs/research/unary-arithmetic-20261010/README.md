# Unary arithmetic: original native observations

Clean compiler producer `34dc675b1b24102014c15e1b22635c481b6858d6`,
Go1.27.2, decision runtime0.2.26. Its displayed development version is0.6.23;
the immutable public0.6.23 producer remains2b17c487.

| Example | Program attempts | Local cases | Consumed caller cases | Separate final inputs | Initial model calls | New saved-replay calls |
| --- | --- | --- | --- | --- | --- | --- |
| Unary branch, deterministic | 2/2 | 1/1 | 1/1 | 3/3 | 0 | 0 |
| Unary branch + own record model | 15/16 | 2/2 | 1/1 | 3/3 | 1 | 0 |

The three final inputs are7, -5 and exact9007199254740993. They differ from the
construction input3; model-training exposure is unknown. The record model's
metadata/weights identities and 2096-byte resident tensors are in the original
mixed report. The one-choice typed helper declines its three-choice model ABI
and continues deterministically; arithmetic source normalization is named in
its binding receipt. Model-free subsequent candidates and replay are explicit.

Whole construction/evaluation CLI wall time was1.32s and8.10s; max RSS was
88,506,368 and88,195,072bytes. These are single local command measurements,
not host CPU utilization or inference-only timing.

Focused final source race checks passed2.409/4.569/5.232s; full bodycodegen race
passed31.922s; all vet and fix passed. The extractor initially failed to
decompose a receiver method. Separating arena storage and else handling produced
four successful logical-file projections (15 files), and generated Go race
passed1.996s. The earlier failed extraction and missing-overlay follow-on log
are preserved. The preparation probe's initial test-helper typo and corrected
feature regression are preserved separately.

Installed workbench06445a4 assembled and replayed the mixed program, but its
old Gooo advice stopped at `expand-joint-profile`. Its final separate inputs
therefore remained0/3. This is retained as a usability defect; it is not relabeled
as native constructor failure or success. The native compiler directly completed
the15-attempt construction above. The workbench adapter needs a separate update.

Every gzip was byte-compared with its original. `FILES.sha256` binds the raw
observations and summary; it does not relabel an earlier producer.
