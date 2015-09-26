This test checks the success case for stacked entities.

* `group:stacked` has a minimal definition in `01-first.json` and a complete
  definition in `02-second.json`.
* `user:stacked` is the same, but for auxiliary groups (where we merge both
  sets), we check both entries only existing in one definition, and an entry
  existing in both definitions (which should only appear once in the merge
  result).
