def indicator($id; $class; $proof; $meta_operation; $observed; $expected; $audiences):
  {
    id: $id,
    class: $class,
    proof_choice: $proof,
    meta_operation: $meta_operation,
    observed: $observed,
    expected: $expected,
    satisfied: ($observed == $expected),
    audiences: $audiences
  };

def audience_view($indicators; $audience; $resolution):
  [$indicators[] | select(.audiences | index($audience))] as $visible
  | ($visible | length) as $total
  | ([$visible[] | select(.satisfied)] | length) as $satisfied
  | {
      audience: $audience,
      resolution: $resolution,
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    };

. as $input
| (if ($input.source.input_digest | test("^sha256:[0-9a-f]{64}$")) then 1 else 0 end) as $user_inputs
| (if $input.source.structural_decision == "ACCEPT" then 1 else 0 end) as $structural_accepts
| (if $input.envelope.source.contract_digest == $input.source.contract_digest then 1 else 0 end) as $contract_bindings
| (if $input.envelope.source.rule_id == "default" then 1 else 0 end) as $default_selections
| (if $input.envelope.decision == "FAIL_CLOSED" and $input.envelope.resolution == "LOWER_RESOLUTION" then 1 else 0 end) as $failed_envelopes
| (if $input.observation.decision == "OBSERVED_FAIL_CLOSED" then 1 else 0 end) as $blocked_observations
| ([$input.replay_matches[] | select(. == true)] | length) as $replay_matches
| (if $input.envelope.decision == "READY" then 1 else 0 end) as $false_ready
| ($input.envelope.effects.executed_invocations + $input.observation.effects.executed_invocations) as $executed_invocations
| ($input.effects.repository_writes + $input.envelope.effects.repository_writes + $input.observation.effects.repository_writes) as $repository_writes
| ([$input.effects.mutation_authority, $input.envelope.effects.mutation_authority, $input.observation.effects.mutation_authority]
    | map(select(. == true)) | length) as $mutation_authorities
| [
    indicator("user.external-inputs"; "DRIVER"; "FOUNDATION"; "count-external-user-value-inputs"; $user_inputs; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("external.structural-accepts"; "DRIVER"; "COHERENCE"; "count-pinned-schema-accepts"; $structural_accepts; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("source.compiler-contract-bindings"; "DRIVER"; "FOUNDATION"; "bind-compiler-default-policy-contract"; $contract_bindings; 1; ["TOOL_AUTHOR", "GOVERNOR"]),
    indicator("compiler.default-policy-selections"; "OUTCOME"; "REGRESSION"; "select-unmatched-value-default-policy"; $default_selections; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("user.fail-closed-envelopes"; "OUTCOME"; "REGRESSION"; "count-default-fail-closed-envelopes"; $failed_envelopes; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("user.blocked-observations"; "OUTCOME"; "COHERENCE"; "reduce-default-failure-for-user"; $blocked_observations; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("tool.deterministic-replay-matches"; "DRIVER"; "REGRESSION"; "compare-default-projection-and-consumer-replays"; $replay_matches; 2; ["TOOL_AUTHOR", "GOVERNOR"]),
    indicator("guardrail.false-ready-envelopes"; "GUARDRAIL"; "REGRESSION"; "count-default-values-classified-ready"; $false_ready; 0; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("guardrail.executed-invocations"; "GUARDRAIL"; "FOUNDATION"; "sum-default-policy-executions"; $executed_invocations; 0; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("guardrail.repository-writes"; "GUARDRAIL"; "FOUNDATION"; "sum-default-coverage-repository-writes"; $repository_writes; 0; ["GOVERNOR"]),
    indicator("guardrail.mutation-authorities"; "GUARDRAIL"; "COHERENCE"; "join-default-coverage-mutation-authority"; $mutation_authorities; 0; ["GOVERNOR"])
  ] as $indicators
| ($indicators | length) as $total
| ([$indicators[] | select(.satisfied)] | length) as $satisfied
| (["OUTCOME", "DRIVER", "GUARDRAIL"] | map(
    . as $class
    | [$indicators[] | select(.class == $class)] as $selected
    | {class: $class, satisfied: ([$selected[] | select(.satisfied)] | length), total: ($selected | length)}
  )) as $classes
| (["FOUNDATION", "COHERENCE", "REGRESSION"] | map(
    . as $proof
    | [$indicators[] | select(.proof_choice == $proof)] as $selected
    | {proof_choice: $proof, satisfied: ([$selected[] | select(.satisfied)] | length), total: ($selected | length)}
  )) as $proofs
| {
    schema: "gooo/external-default-coverage-report/v1",
    subject_sha: $input.subject_sha,
    metric_id: "gooo.metric.user.external-default-coverage.v1",
    decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
    resolution: (if $satisfied == $total then "DEFAULT_POLICY_COVERAGE_ONLY" else "INVARIANT_ONLY" end),
    reason: (if $satisfied == $total
      then "STRUCTURAL_ACCEPT_VALUE_FAIL_CLOSED_OBSERVED"
      else "EXTERNAL_DEFAULT_POLICY_COVERAGE_INCOMPLETE"
      end),
    source: $input.source,
    contrast: {
      structural_decision: $input.source.structural_decision,
      value_decision: $input.envelope.decision,
      value_resolution: $input.envelope.resolution,
      selected_rule: $input.envelope.source.rule_id,
      user_decision: $input.observation.decision
    },
    coordinates: {
      satisfied: $satisfied,
      total: $total,
      basis_points: (if $total == 0 then 0 else (($satisfied * 10000 / $total) | floor) end)
    },
    classes: $classes,
    indicators: $indicators,
    views: [
      audience_view($indicators; "USER"; "USER_VISIBLE"),
      audience_view($indicators; "TOOL_AUTHOR"; "TOOL_CONTRACT"),
      audience_view($indicators; "GOVERNOR"; "FULL_RECEIPT")
    ],
    proofs: $proofs,
    envelope: {
      decision: $input.envelope.decision,
      resolution: $input.envelope.resolution,
      reason: $input.envelope.reason,
      diagnostics: $input.envelope.diagnostics,
      digest: $input.envelope.digest
    },
    observation: {
      decision: $input.observation.decision,
      resolution: $input.observation.resolution,
      reason: $input.observation.reason,
      value: $input.observation.value,
      digest: $input.observation.digest
    },
    effects: {
      executed_invocations: $executed_invocations,
      repository_writes: $repository_writes,
      mutation_authorities: $mutation_authorities
    },
    promotion_credit_bps: 0,
    repository_writes: $repository_writes,
    mutation_authority: ($mutation_authorities > 0),
    not_claimed: [
      "generalized arbitrary user input",
      "coverage of every default-policy value shape",
      "effect execution",
      "complete interpreter semantics",
      "domain correctness",
      "production readiness"
    ]
  }
