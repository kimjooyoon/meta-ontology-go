# Language operation catalog

This baseline keeps two activities in one Gooo source. `IncrementOne` has a
registry-bound value program; `IncrementTwo` is an explicit extension slot
without a program.

The baseline CI is expected to pass while reporting source-only extension
progress as `0/1`. A later PR may change only the `IncrementTwo` declaration to
add `computes "int.add:2"`; the same observer must then report `1/1`.

Neither state authorizes repository mutation, adds a general expression
language, or proves runtime memory and performance bounds.
