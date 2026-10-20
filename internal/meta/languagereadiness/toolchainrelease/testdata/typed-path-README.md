# Original typed preflight and native regression observations

The four `typed-path-*.json.gz` files are byte-identical gzip -n copies of actual
CLI stdout from clean accepted compiler source
`426caecb711e47da26fe659237a117f301d112f4`, captured on 2026-10-10 KST with
Go1.27.2 and the existing own model. Their embedded producers remain unchanged.
No original generation was rerun to create these fixtures.

Unary inspection declined the one-choice typed input; record inspection was
ready and its input digest matches actual construction. Native construction
retained first local2/2 and caller0/1, 11/16 attempts, separate final3/3 and
saved replay3/3 with zero new inference. Record inference was one call; typed
inference was zero. Exact input9007199254740993/output27021597764222979 are
checked through JSON RawMessage / Number values, with no float conversion.

These are portable unit regression fixtures. Release readiness separately
executes each candidate on four native platforms and retains its own raw
outputs. The fixtures establish neither training independence nor general
model accuracy. The original full observation is linked from the release guide.
