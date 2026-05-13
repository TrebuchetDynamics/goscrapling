# goscrapling-tdd-slice

Use this before implementing any runtime behavior.

## Loop

1. Pick one builder-ready `progress.json` row.
2. Write the smallest failing test that captures the row contract.
3. Run the focused test and confirm it fails for the expected reason.
4. Implement the minimum code to pass.
5. Run the focused test, then `go test ./... -count=1`.
6. Refactor only after tests are green.
7. Update `progress.json`, feature map, and coverage ledger if behavior status
   changed.

## Test Preferences

- Parser behavior: string fixtures.
- Fetcher behavior: `httptest.Server`.
- Store behavior: `t.TempDir()`.
- Browser behavior: fake browser engine first, real browser smoke tests later.
- Spider behavior: fake fetcher and deterministic scheduler fixtures.

## Stop Conditions

Stop and return to planner if the row requires multiple subsystems, live
credentials, live web access, or unresolved API design.
