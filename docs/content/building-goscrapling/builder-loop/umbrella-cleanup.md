# Umbrella Cleanup

Umbrella rows are inventory rows, not implementation slices. Split these before
assigning them to a builder.

<!-- PROGRESS:START kind=umbrella-cleanup -->
| Phase | Umbrella row | Owner | Not ready when |
|---|---|---|---|
| phase-4-spider / spider-core | Robots, cache, checkpoint, and stats as separate crawler slices | `spider` | This umbrella has not been split into one row per production control. |
<!-- PROGRESS:END -->
