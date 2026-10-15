# Imported declaration compilation origins

Fresh source cohort `compile101719` binds declared types to Go compilation and unchanged source inputs. Its functional source revisions change a named type from integer to floating-point. Both source and export identities change, project type objects are fresh, and standard-library identity remains shared. A separate invalid dependency body is rejected before declaration-only reading.

`project108271` is a newly authored consumer of the current generation package. Its source is saved in the observation. It binds 60 compiled dependencies with two metadata queries and approximately 2.49 seconds of measured import/type-read work. It executed zero old tests and has no matched historical timing baseline.

The compiler's complete race suite and end-to-end deadline are separate scopes. Completed local source cohorts were not rerun.
