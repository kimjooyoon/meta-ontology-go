Saved observation: every row below describes the recorded invocation.
Recorded file SHA-256: 3069711b83f236ffb73d08ddc4f79da0437edc8c063c168401a4d6edfde22ec2
Reading this report made 0 model calls and 0 program executions.
Recorded claims have not been reverified. Exit 0 means this report was read, including recorded failures.

# Gooo construction result

| Observation | Result |
| --- | --- |
| Mode | <code>new construction</code> |
| Entry activity | <code>Price</code> |
| Construction stage | <code>COMPLETE</code> |
| Construction decision | <code>PARTIAL&#95;FINITE</code> |
| Stop reason | <code>DECLARED&#95;SPACE&#95;EXHAUSTED</code> |
| Attempted programs / budget | <code>2 / 2</code> |
| Source-bounded candidate combinations | <code>2</code> |
| Rejected before caller execution | <code>0</code> |
| Selected attempt | <code>1</code> |
| Selected local checks | <code>not measured (no scored checks)</code> |
| Caller expected outputs used for selection | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| First attempt caller expected outputs | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| Selected Gooo criteria | <code>1/2 satisfied; 1 unsatisfied; 0 unknown (predicate/input pairs)</code> |
| Initial Gooo criteria | <code>Price&#58; 1/2 satisfied; 1 unsatisfied; 0 unknown</code> |
| Final Gooo criteria on held inputs | <code>Price&#58; 1/2 satisfied; 1 unsatisfied; 0 unknown</code> |
| Criteria holdout caller overlap | <code>0 observed; 1 other; 0 unknown</code> |
| Gooo criteria questions | <code>Price&#58; NO&#95;SURVIVING&#95;CANDIDATE</code> |
| Criteria input obligations | <code>1 initial + 0 acquired = 1 effective</code> |
| Criteria question work | <code>2 candidate evaluations; 4 predicate evaluations; 0 reused outputs; 0 reused predicate results</code> |
| Criteria question strategy | <code>distinguish&#95;then&#95;check</code> |
| Alternatives after criteria questions | <code>0 surviving; 2 rejected; 0 unavailable; 0 unobserved</code> |
| Initial Gooo rejection explanation | <code>Price&#58; ALL&#95;DECLARED&#95;CANDIDATES&#95;REJECTED; 2/2 candidates examined; 2 rejected, 2 explained; 0 unavailable, 0 unobserved</code> |
| Rejection history source | <code>sha256&#58;9e34df7ab245a8a0d81431ace68c962e4803971d38e21b5e5b6d2d57854ccfbb</code> |
| Recorded rejection reason | <code>cost-covered (CostCovered) returned false for input &#91;10&#93;&#58; candidate 0 produced {&#34;price&#34;&#58;10}</code> |
| Recorded rejection reason | <code>within-budget (WithinBudget) returned false for input &#91;10&#93;&#58; candidate 1 produced {&#34;price&#34;&#58;30}</code> |
| Rejection scope | <code>initial recorded question history and declared candidates; other implementations may satisfy the conditions</code> |
| Evaluation stage | <code>COMPLETE</code> |
| Evaluation expected outputs | <code>0/1 (0.00%) matched; 1 mismatched, 0 faulted, 0 blocked, 0 unobserved; 0 additional unscored faults/blocks</code> |
| Evaluation input overlap | <code>1 unique tuples&#58; 0 consumed during caller construction, 1 other; 0 duplicate rows</code> |
| Constructed input history | <code>UNKNOWN&#58; NO&#95;DISJOINT&#95;INPUTS</code> |
| Actual constructed inputs | <code>1 unique root tuples&#58; 1 overlapping, 0 disjoint, 0 unknown; 0 duplicate rows</code> |
| Disjoint cases matching every expectation | <code>not measured (no scored checks)</code> |
| Evaluation observation | <code>case index 0, activity quote&#58;//activity/price&#58; value mismatch; actual {&#34;price&#34;&#58;11}; expected {&#34;price&#34;&#58;21}</code> |

Shares count the named supplied checks or expected activity outputs, not all possible inputs.
Caller overlap counts root tuples. Constructed-input coverage follows actual arguments and recorded local/caller history; model-training exposure remains unknown.
Source observation rows describe only the declared alternatives and budgets. Behavior on other inputs is not assessed by this report.
Typed-path rows retain initial local search history; these candidates differ from caller program attempts.
Recorded local model calls belong to that history, including when a saved construction is replayed.
Use --format json for all attempts, values and evidence. Saved --out files remain full JSON and source.
