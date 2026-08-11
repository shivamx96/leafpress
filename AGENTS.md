# Repository instructions for coding agents

## Cross-module Go changes

Leafpress contains two independently versioned Go modules, `core` and `cli`.
The root `go.work` makes local development use `./core`, which can hide a stale
Core version in `cli/go.mod`.

Whenever CLI behavior or CLI tests depend on a change made in `core`:

1. Put the Core implementation in a commit that can be fetched from the
   repository.
2. In a following commit, pin `github.com/shivamx96/leafpress/core` in
   `cli/go.mod` to that commit's Go pseudo-version and update `cli/go.sum`. Use
   `GOWORK=off` while updating the dependency so the local workspace cannot
   mask the declared module version.
3. Validate the external-consumer path from `cli/` with:

   ```bash
   GOWORK=off go test -mod=readonly ./...
   ```

   Run the normal workspace tests as well.
4. During release preparation, replace the pseudo-version with the matching
   released Core version. Publish the `core/vX.Y.Z` tag before the
   `cli/vX.Y.Z` tag, following `docs/06_RELEASE_PROCESS.md`.

Do not solve a standalone-module failure by adding a `replace` directive to
`cli/go.mod`, enabling `go.work` in the standalone CI job, or skipping the CLI
tests. Those approaches hide the mismatch that external CLI consumers would
experience.

If the Core commit is not yet fetchable and committing or pushing is outside
the current task's authorization, report the required dependency pin as
remaining work rather than treating workspace-only test success as complete.
