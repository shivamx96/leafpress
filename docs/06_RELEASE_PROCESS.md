# Release process

Leafpress is a multi-module Go repository:

- `github.com/shivamx96/leafpress/core`
- `github.com/shivamx96/leafpress/cli`

Go resolves versions for nested modules from scoped Git tags. A product
release therefore has three tags on the same commit. For version `vX.Y.Z`:

1. `core/vX.Y.Z` publishes the core module.
2. `cli/vX.Y.Z` publishes the CLI module.
3. `vX.Y.Z` creates the GitHub release and binary archives used by the curl
   installer and self-updater.

The CLI's `go.mod` must require the same `vX.Y.Z` core version. The only local
`replace` belongs in `go.work`, which selects `./core` without leaking local
filesystem paths into the published CLI module.

## Cutting a release

From a clean release commit, using `v1.0.0-beta.9` as an example:

```bash
git tag core/v1.0.0-beta.9
git tag cli/v1.0.0-beta.9
git push origin core/v1.0.0-beta.9 cli/v1.0.0-beta.9

GOWORK=off go install github.com/shivamx96/leafpress/cli/cmd/leafpress@v1.0.0-beta.9

git tag v1.0.0-beta.9
git push origin v1.0.0-beta.9
```

The release workflow refuses to build the product release unless both scoped
module tags resolve to the exact product-tag commit. It also performs the
public `GOWORK=off go install` before building archives.

Never publish the CLI tag before its required core tag: Go's module proxy must
be able to resolve the core dependency when it indexes the CLI module.
