# Optional own Gooo joint path model

An explicit `--path-model` may supply the separate
`gooo/tiny-joint-path-model/v1` schema from SDK v0.2.12-experimental.
The compiler binds the original Gooo body to the typed fallback before opening
or predicting with the model. It projects two complete source-v3 inputs and
judges the four absolute legal masks in one initial prediction. Tests remain
the finite contract; later observed failures may rank remaining paths.

    gooo body-codegen --json --path-plan plan.json --path-model joint/model.json --path-step-attempts 1 --path-feedback-rounds 3 --activity ChoosePath source.gooo

This head requires exactly two two-option decisions. Unsupported arity or
complete context overflow declines optional inference and retains deterministic
finite construction, including partial results. Seed and feedback are skipped
on that decline. The final sole path makes zero further predictions.

`NewTypedPathGenerator` also loads this immutable model once; each request owns
its source, plan, session, workspace and progress. Existing operation/eight-label
path models preserve their ABI. The classifier cannot edit source intent,
change test expectations, grant CI authority or update model weights online.

Weights, matched comparison and negative results:
https://github.com/kimjooyoon/gooo-neural-decision-experiments/tree/main/models/joint-composition-v1

The own joint model has 12,412 parameters; ternary weights occupy 2,590 disk
bytes, decode to 12,496 tensor bytes plus eight scale bytes, and use a
2,160-byte workspace. This is packed storage, not whole-process 1.58-bit RAM.
The research evaluation retains explicit case/view denominators and calibration
selection; no automatic default model promotion or universal completeness claim.
