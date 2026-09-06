# Native registration through common execution

## Decision

RegisterSyntaxCapability is a fifth native operation in the common planner,
not a free-form executor supplied by a caller. The four legacy operations keep
their exact original input contract. The registration operation separately
binds the existing compiled syntax-registration Gooo contract and its source
and semantic digests. No capability denominator or semantic admission policy
is lowered.

The path is explicit:

typed Request -> observed corpus-presence indicator -> common Plan -> Action
-> ExecutionStep -> real worker -> temp-copy conformance -> OperationReceipt.

The request digest travels with the indicator and receipt. The full typed
request travels with the selected action and execution step. Case IDs, paths,
policies, and toolchain identities are not recovered from prose.

## User path

Use the same compiled meta-execution executable for inspection and execution.
The identity includes that executable, the Go driver, and the compiler. An
inspector built from another command is not interchangeable.

1. Prepare an explicit registration request and source metrics.
2. Pin the request using --registration-mode=inspect, --registration-root and
   --registration-request. Its JSON result is the exact request for later steps.
3. Produce the common plan using --registration-mode=plan with the pinned
   request, --source-metrics and --registration-base.
4. Run the normal --plan/--output path with the same executable and explicit
   --source-metrics. Read the manifest and meta-operation-observations.json.
5. Verify the observation bundle and receipts before treating execution as
   conformant. A successful CLI exit alone is not proof of every operation.

Inspection and planning write JSON to stdout. The common executor uses private
temporary directories. The native registration verifier applies the complete
candidate only to a dereferenced temporary copy, then runs the two existing
syntax/closure conformance packages. No candidate is applied to the input
project. Merge, release, and semantic approval remain separate authorities.

## Fixed boundaries and observations

- Native bindings: four unchanged legacy bindings plus one exact registration
  binding; callers cannot add or replace an executor.
- Independent operation floor: 2, unchanged. A registration-only request is
  UNKNOWN/pressure shortfall, not a successful singleton plan.
- Required registration indicators: 4. Artifact completeness, execution
  identity, real native conformance, and deterministic worker replay each have
  a compiled producer/consumer connection and actual operation evidence.
- Artifact roles: 9, unchanged. Physical file count is source-view dependent.
- The native acceptance case pairs registration with a real CollapseAssignReturn
  execution; it does not synthesize a companion receipt.
- UNKNOWN preserves stage, step, reason, unknown_class, next_operation, and
  blocked_by. Missing typed input and stale input have different causes.
- REFUTED remains distinct from missing evidence. Resealing a substituted
  request or receipt input digest cannot make it valid.
- Build and execution durations are integer wall milliseconds from the same
  native test. They are observations, not an improvement claim.
- Utility and improvement remain UNKNOWN. Internal CI is not external utility
  evidence, and no before/after comparison is inferred.
- The evidence capture candidate is a separately labelled read-only replay of
  the same request, not a claim that a digest authorizes application.

## CI authority

The new native workflow adds a narrow independent check and preserves the
existing CI workflows. It does not weaken required checks, invent approvals,
change branch protection, bypass the Guardian, or fix the open release-cache
security finding. Those remain separately observable integration concerns.

Local validation is intentionally not run. Native Actions must establish
actual conformance before this implementation can be counted as closed.
