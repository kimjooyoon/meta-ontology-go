def require($ok; $reason): if $ok then . else error($reason) end;
def paths: .report.body_paths;
($context[0]) as $c |
($deterministic[0] | paths) as $d |
($model[0] | paths) as $m |
($native[0].construction.initial.preparations[0].generation | paths) as $n |
($m.declared_contract_ranking) as $r |
require($c.model_predictions == 0 and $c.candidate_tests == 0 and
  $c.selected_emission == false and $c.repository_writes == 0; "inspection performed work") |
require($c.model_compatibility.status == "READY_FOR_RANKING" and
  $c.model_compatibility.model.artifact_sha256 == "sha256:4b61fe8d84df77c8dfac6a880eedcddfd8f5f903fa538d1e80532dd77f74b1ac";
  "unexpected model or representation") |
require(($c.inputs|length) == 2 and all($c.inputs[]; (.ordered_source_features|length) == 528 and .bytes == 2112 and .flow_features == null);
  "ordered-source shape differs") |
require($c.declared_contract_cases.count == 7 and ($c.declared_contract_cases.features|length) == 7 and
  all($c.declared_contract_cases.features[]; length == 32) and
  $c.declared_condition_cases.count == 3 and ($c.declared_condition_cases.features|length) == 3 and
  all($c.declared_condition_cases.features[]; length == 32); "case or condition stream differs") |
require($r.local_model_predictions == 1 and $r.ranking_applied == true and $r.representation_declined == false;
  "ranking was not applied exactly once") |
require(all(range(0; 2); . as $i | $c.inputs[$i].input_sha256 == ("sha256:" + $r.source_feature_sha256[$i]));
  "inspection/source ranking digest mismatch") |
require($c.declared_contract_cases.finite_cases_sha256 == $r.finite_cases_sha256 and
  $c.declared_condition_cases.input_sha256 == ("sha256:" + $r.condition_feature_sha256);
  "inspection/goal ranking digest mismatch") |
require(all([$d, $m, $n][]; .finite_functional_completeness_percent == 100 and
  .search.selected_training_passed == 7 and .conditions.passed == 3); "source contract incomplete") |
require($d.search.selection.local_model_predictions == 0 and $m.search.selection.local_model_predictions == 1 and
  $n.search.selection.local_model_predictions == 1 and
  all($m.declared_contract_progress[]; .search.new_local_model_predictions == 0); "model-call accounting differs") |
require($n.declared_contract_ranking.source_feature_sha256 == $r.source_feature_sha256 and
  $n.declared_contract_ranking.condition_feature_sha256 == $r.condition_feature_sha256 and
  $n.model_retention.artifact_sha256 == $c.model_compatibility.model.artifact_sha256;
  "native construction used another input or model") |
require($native[0].evaluation.runtime.finite_passed == 11 and $native[0].evaluation.runtime.finite_total == 11 and
  $replay[0].evaluation.runtime.finite_passed == 11 and $replay[0].evaluation.runtime.finite_total == 11 and
  $replay[0].generated_now == false and $replay[0].evaluation.construction_replayed == true and
  $replay[0].evaluation.new_model_calls == 0; "native construction or saved replay failed") |
require($delta[0].counts == {"assessment:STILL_MATCHING":11,"outcome:UNCHANGED":11,"presence:BOTH":11,"requirement:UNCHANGED":11};
  "saved native observations differ") |
{schema:"gooo/interaction-native-observation/v1",source_sha:$source_sha,
 model_artifact_sha256:$c.model_compatibility.model.artifact_sha256,
 scope:"one native integration example; finite authored goals; no new held-out accuracy or causal speedup claim",
 context:{choices:2,source_cells_per_choice:528,output_cases:7,condition_cases:3,predictions:0,candidate_tests:0},
 deterministic:{attempts:($d.search.attempts|length),model_calls:0,codegen_ms:$d.timing.total_ms},
 model:{attempts:($m.search.attempts|length),model_calls:1,predict_ns:$r.predict_ns,
   codegen_ms:$m.timing.total_ms,model_load_ms:$m.timing.model_load_ms},
 native:{model_calls:1,predict_ns:$n.declared_contract_ranking.predict_ns,passed:11,total:11},
 replay:{model_calls:0,passed:11,total:11,unchanged_comparison_groups:11},
 customer_savings:null,external_adoption:null}
