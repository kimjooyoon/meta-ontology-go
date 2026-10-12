Saved observation: every row below describes the recorded invocation.
Recorded file SHA-256: 2c223c2afbdcaa2bb47bef36822e54c4f710b7b672b5be9c6e3d7fe0c09a737b
Reading this report made 0 model calls and 0 program executions.
Recorded claims have not been reverified. Exit 0 means this report was read, including recorded failures.

# Gooo construction result

| Observation | Result |
| --- | --- |
| Mode | <code>new construction</code> |
| Entry activity | <code>RoundTrip</code> |
| Construction stage | <code>COMPLETE</code> |
| Construction decision | <code>PARTIAL&#95;FINITE</code> |
| Stop reason | <code>DECLARED&#95;SPACE&#95;EXHAUSTED</code> |
| Attempted programs / budget | <code>4 / 4</code> |
| Source-bounded candidate combinations | <code>4</code> |
| Rejected before caller execution | <code>0</code> |
| Selected attempt | <code>1</code> |
| Selected local checks | <code>not measured (no scored checks)</code> |
| Caller expected outputs used for selection | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| First attempt caller expected outputs | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| Selected Gooo criteria | <code>2/2 satisfied; 0 unsatisfied; 0 unknown (predicate/input pairs)</code> |
| Initial Gooo criteria | <code>Decode&#58; 1/1 satisfied; 0 unsatisfied; 0 unknown</code> |
| Initial Gooo criteria | <code>Encode&#58; 1/1 satisfied; 0 unsatisfied; 0 unknown</code> |
| Evaluation stage | <code>COMPLETE</code> |
| Evaluation expected outputs | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| Evaluation input overlap | <code>1 unique tuples&#58; 0 consumed during caller construction, 1 other; 0 duplicate rows</code> |
| Constructed input history | <code>PROGRESS&#58; DISJOINT&#95;EXPECTATIONS&#95;DIFFER</code> |
| Actual constructed inputs | <code>1 unique root tuples&#58; 0 overlapping, 1 disjoint, 0 unknown; 0 duplicate rows</code> |
| Disjoint cases matching every expectation | <code>0/1 (0.00%)</code> |
| Evaluation observation | <code>case index 0, activity relation-codec&#58;//roundtrip&#58; value mismatch; actual false; expected true</code> |

Shares count the named supplied checks or expected activity outputs, not all possible inputs.
Caller overlap counts root tuples. Constructed-input coverage follows actual arguments and recorded local/caller history; model-training exposure remains unknown.
Source observation rows describe only the declared alternatives and budgets. Behavior on other inputs is not assessed by this report.
Typed-path rows retain initial local search history; these candidates differ from caller program attempts.
Recorded local model calls belong to that history, including when a saved construction is replayed.
Use --format json for all attempts, values and evidence. Saved --out files remain full JSON and source.
