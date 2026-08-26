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

def required_property($schema; $name):
  (($schema.required // []) | index($name)) != null;

def schema_guarantees_non_empty_activity($schema):
  ($schema.properties.activity.const? // null) as $activity
  | required_property($schema; "activity")
    and (($activity | type) == "string")
    and (($activity | length) > 0);

def schema_guarantees_non_empty_inputs($schema):
  ($schema.properties.inputs // {}) as $inputs
  | ($inputs.prefixItems // []) as $prefix_items
  | ($inputs.minItems // 0) as $min_items
  | required_property($schema; "inputs")
    and $inputs.type == "array"
    and (($min_items | type) == "number")
    and $min_items >= 1
    and $inputs.items == false
    and (($prefix_items | length) >= $min_items)
    and ([$prefix_items[]
      | (.const? // null) as $value
      | (($value | type) == "string" and ($value | length) > 0)
    ] | all);

def schema_entails_complete_ready($schema; $contract):
  schema_guarantees_non_empty_activity($schema)
  and schema_guarantees_non_empty_inputs($schema)
  and any($contract.rules[]?;
    .id == "complete-symbolic-invocation"
    and .match.activity == "NON_EMPTY"
    and .match.inputs == "NON_EMPTY"
    and .decision == "READY"
    and .resolution == "VALUE_EXACT"
  );

. as $input
| (if ($input.source.input_digest | test("^sha256:[0-9a-f]{64}$")) then 1 else 0 end) as $user_inputs
| (if $input.source.structural_decision == "REJECT" and $input.source.validator_exit_code == 1 then 1 else 0 end) as $structural_rejects
| schema_entails_complete_ready($input.structural_schema; $input.value_contract) as $schema_entails_ready
| (if $schema_entails_ready then 1 else 0 end) as $schema_ready_entailments
| (if $schema_entails_ready then 0 else 1 end) as $default_post_schema_reachability
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
    indicator("external.structural-rejects"; "DRIVER"; "COHERENCE"; "count-pinned-schema-rejects"; $structural_rejects; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("compiler.schema-ready-entailments"; "DRIVER"; "FOUNDATION"; "compare-generated-schema-to-ready-rule"; $schema_ready_entailments; 1; ["TOOL_AUTHOR", "GOVERNOR"]),
    indicator("compiler.default-post-schema-reachability"; "GUARDRAIL"; "COHERENCE"; "count-default-paths-reachable-after-structural-gate"; $default_post_schema_reachability; 0; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
    indicator("source.compiler-contract-bindings"; "DRIVER"; "FOUNDATION"; "bind-compiler-default-policy-contract"; $contract_bindings; 1; ["TOOL_AUTHOR", "GOVERNOR"]),
    indicator("compiler.counterfactual-default-selections"; "OUTCOME"; "REGRESSION"; "select-unmatched-value-default-counterfactually"; $default_selections; 1; ["USER", "TOOL_AUTHOR", "GOVERNOR"]),
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
    metric_id: "gooo.metric.user.external-default-defense.v1",
    decision: (if $satisfied == $total then "PASS" else "FAIL_CLOSED" end),
    resolution: (if $satisfied == $total then "COUNTERFACTUAL_DEFENSE_ONLY" else "INVARIANT_ONLY" end),
    reason: (if $satisfied == $total
      then "STRUCTURAL_REJECT_COUNTERFACTUAL_DEFAULT_FAIL_CLOSED"
      else "EXTERNAL_DEFAULT_DEFENSE_INCOMPLETE"
      end),
    source: $input.source,
    contrast: {
      structural_decision: $input.source.structural_decision,
      value_decision: $input.envelope.decision,
      value_resolution: $input.envelope.resolution,
      selected_rule: $input.envelope.source.rule_id,
      user_decision: $input.observation.decision,
      projection_mode: $input.source.projection_mode,
      schema_entails_ready_rule: $schema_entails_ready,
      default_reachable_after_structural_gate: ($default_post_schema_reachability > 0)
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
      "reachable default selection after structural conformance",
      "generic JSON Schema-to-value-contract entailment",
      "effect execution",
      "complete interpreter semantics",
      "domain correctness",
      "production readiness"
    ]
  }
