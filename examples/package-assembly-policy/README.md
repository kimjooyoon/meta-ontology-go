# A Gooo package controls construction of another Gooo workspace

The diagnostic workspace builds a report and an explanation. This separate
policy workspace receives the observed case counts after each scored candidate
and decides whether construction should continue. `policy.gooo.fixture` calls
`counts.CanContinue(input)` from a shared Gooo package. The compiler resolves
and checks that helper through the same package mechanism used by the target.

## Run the connected tool

From this repository revision with Go 1.27.1:

```sh
go build -o /tmp/gooo-package-policy ./cmd/gooo
/tmp/gooo-package-policy package execute --json \
  --cases examples/package-body-calls/cases.json \
  --assembly-policy-workspace examples/package-assembly-policy/gooo.workspace.json \
  examples/package-body-calls/gooo.workspace.json > package-policy.json
```

Add `--assembly-model /path/to/shared-qat/model.json` for the optional
[own compact model](https://huggingface.co/asketeddy/gooo-record-shared-field-tiny-v1/tree/5800946afb35506d66357ee3ea6f956f506795b9/models/qat_ternary).
The model ranks the target's declared choices. The fixed Gooo policy consumes
actual observations before the next candidate is constructed. Without the model
flag, candidate order is deterministic. Use `--go /path/to/go1.27.1` if needed.

Replace the policy manifest with `checkpoint.workspace.json` to stop after the
first scored candidate. The result retains the partial cases, fields and runtime
mismatches. The operation does not turn that candidate into a complete result.
The supported operations and case-count contract are described in the
[assembly-policy guide](../assembly-policy/README.md#meaning-of-the-operations).

## Replay without policy files or inference

```sh
/tmp/gooo-package-policy package replay --json \
  --receipt package-policy.json \
  --cases examples/package-body-calls/cases.json \
  examples/package-body-calls/gooo.workspace.json
```

The target workspace must still match the saved source. The policy's source
manifest, source files, imported helpers, identity mapping and lowering travel
in `result.assembly_policy`. Replay rebuilds that saved policy, compares its
lowering and checks every recorded policy decision through the existing
composition replay. Policy files and model files are not opened for replay.
Changing a policy for a new construction requires a new execute command.

The package snapshot preserves which source produced the policy. It does not
authenticate the author or the origin of a copied receipt. The original source
names and namespace-derived IDs remain visible beside their lowered identities.
Model observations in a saved construction are historical; replay reports zero
new model calls and fresh native outputs.

## Current construction contract

- The policy entry accepts the `gooo://tools/assembly-observation` record and
  returns `gooo://tools/assembly-explanation` with the declared field IDs.
- Policies have fixed typed bodies and can call fixed pure helpers across
  packages. Their source manifest is bounded to 128 KiB of encoded JSON.
- Bound producer activities are not supplied as policy inputs. The counts are
  delivered directly by the record-construction adapter.
- Record-choice activities receive the policy. Other assembly profiles retain
  their existing behavior; at least one record-choice activity is required.
- The source owns candidates and total attempt budget. Policy messages remain
  domain judgments; matched/total reports come from actual finite checks.
- This route executes one bounded construction. Saved continuation is currently
  available through `body-compose`, and has a separate receipt contract.

Tests exercise continued and partial construction, imported helpers, replay
after policy-file removal, changed-source rejection and agreement between policy
interpretation and separately compiled native policy outputs.

## 한국어: 조립 방법도 Gooo 도구로 나누기

진단 도구가 부품을 조립하면, 정책 도구는 그때까지 관측한 결과를 읽고 다음
부품을 시도할지 판단합니다. “예산이 남았는가” 같은 작은 규칙은 별도 패키지의
함수로 가져다 씁니다. 도구를 만드는 규칙도 Gooo로 작성하고 재사용하는 구조입니다.

정책과 모델의 역할도 실행 기록에서 읽을 수 있습니다. 모델은 후보 순서를
제안하고, 정책은 실제 사례 결과를 보고 계속할지 결정합니다. 저장한 결과에는
정책의 원본 패키지도 들어 있어, 나중에 모델 없이 같은 판단과 실행을 재구성할
수 있습니다. 첫 후보에서 멈췄다면 충족하지 못한 사례도 함께 남습니다.
