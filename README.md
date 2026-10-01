# Geul event contracts

Canonical Protobuf, AsyncAPI, SpiceDB, and generated language contracts for Geul.

## Packages

- `@echovisionlab/geul-proto`: generated TypeScript Protobuf bindings
- `@echovisionlab/geul-event`: typed event names and helpers
- `github.com/echovisionlab/geul-event-contracts`: generated Go contracts

## Development

Use the Node and pnpm versions pinned by this repository.

```sh
corepack enable
pnpm install --frozen-lockfile
pnpm proto:lint
pnpm catalogs:check
pnpm spicedb:check
pnpm typecheck:proto
pnpm typecheck:event
pnpm test:ts
```

Generated files are committed. After changing a source contract, run the
matching generator and commit both the source and generated output.

## Post configuration concurrency

Management `Post.configuration_revision` is the version of entity-wide slug,
comments, map location, and document layout settings. It is independent of the
collaborative document revision and locale metadata. `UpdatePost` requires the
resident `expected_configuration_revision` and returns the actual persisted
`configuration_revision`, including unchanged saves. The API compares it under
the Post root lock after checking current edit authority. A missing revision
requires a reload; a malformed UUID is invalid; a stale revision returns
`ABORTED` without changing settings. Clients advance their resident version only
from their own acknowledged save and preserve pending edits on conflict rather
than silently fetching a newer revision to retry them.

The database adds a dedicated UUID and advances it only when the four settings
change, including a change back to a previous value. Body, locale, and lifecycle
updates do not advance it. Deploy the `post-configuration-revisions-v1` migration
before an API that reads this field, then deploy the revision-aware Web editor.
An older editor's unversioned settings writes are rejected safely and require a
reload; they are not accepted through a compatibility bypass.

## Release

Release Please versions both npm packages together. npm publication uses
GitHub Actions trusted publishing; no npm token is stored in the repository.

## License

PolyForm Noncommercial 1.0.0. Commercial use requires a separate license from
Echo Vision Lab. See [LICENSE.md](LICENSE.md).
