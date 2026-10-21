# Geul event contracts

Canonical Protobuf, AsyncAPI, SpiceDB, and generated language contracts for Geul.

## Packages

- `@echovisionlab/geul-proto`: generated TypeScript Protobuf bindings
- `@echovisionlab/geul-event`: typed event names and helpers
- `github.com/echovisionlab/geul-event-contracts`: generated Go contracts

## Private dependency access

The geul repositories and npm packages are private. The npm packages
`@echovisionlab/geul-common`, `@echovisionlab/geul-proto`,
`@echovisionlab/geul-event`, and `@echovisionlab/geul-telemetry` use
[GitHub Packages](https://github.com/orgs/echovisionlab/packages).

For local development and consuming projects, use this `.npmrc` configuration
(already checked in here):

```ini
registry=https://registry.npmjs.org/
@echovisionlab:registry=https://npm.pkg.github.com
//npm.pkg.github.com/:_authToken=${NODE_AUTH_TOKEN}
```

Export `NODE_AUTH_TOKEN` from your secret manager before running pnpm. Use a
GitHub personal access token (classic) with `read:packages` and access to the
packages; authorize it for the organization SSO if required. Never commit a
literal token. See [GitHub's npm registry authentication guide](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-npm-registry#authenticating-to-github-packages).

The registry mapping applies to the entire `@echovisionlab` scope. The public
`@echovisionlab/ionian` and `@echovisionlab/audio-transcoder` packages stay on
npmjs. Consumers that need them must use explicit, version-pinned npmjs tarball
URLs in their dependency manifests, for example:

```json
{
  "dependencies": {
    "@echovisionlab/ionian": "https://registry.npmjs.org/@echovisionlab/ionian/-/ionian-VERSION.tgz",
    "@echovisionlab/audio-transcoder": "https://registry.npmjs.org/@echovisionlab/audio-transcoder/-/audio-transcoder-VERSION.tgz"
  }
}
```

Replace `VERSION` with each package's actual pinned version and regenerate the
consumer lockfile. This repository does not depend on either package, so no
manifest or lockfile exceptions are needed here. Package-specific registry
entries in `.npmrc` cannot override a scope mapping.

For private Go modules, including `github.com/echovisionlab/geul-event-contracts`
and `github.com/echovisionlab/geul-telemetry`, configure:

```sh
export GOPRIVATE='github.com/echovisionlab/*'
```

Authenticate Git HTTPS access with your credential manager and a token with
repository contents read access. `GOPRIVATE` bypasses the public Go proxy and
checksum database for these modules; it does not authenticate Git. CI uses
`PRIVATE_DEPS_TOKEN` through a Git URL rewrite and removes it at job completion.

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

## Public contact variables

Content uses `contact.email`; `contact.legalEmail` remains independent. The
catalog accepts exactly five keys and rejects the removed support/privacy keys.
Management `PublicSettings.contact_email` is field 26; public manifest
`SiteSettings.contact_email` is field 18. The support/privacy fields are removed.
There are no deprecated read aliases or runtime fallbacks.

Deploy the schema's `contact-email-v1` migration before API consumers of the new
column. Before the canonical-only API, Web and Collab start, migrate stored live
content and historical snapshots to the canonical keys and catalog identity
under the API migration's backup, writer shutdown and revision contract.
Updating generated bindings alone does not migrate stored content or settings.
All consumers must use the same catalog fingerprint; older clients must reload.

## Display size

Resizable blocks store one shared integer, `displaySize` (10–100, default 100),
which replaces `previewWidth`. Renderers cap the block's width at that percentage
of the content column. Blocks that keep an aspect ratio (images, videos, maps and
standalone link videos) also cap their height at the same percentage of 80% of
the viewport's small height (`80svh`), so tall images fit the screen. Email
renders blocks at the email width and does not apply display size. Authors pick the XS 25, S 40,
M 60, L 80 or XL 100 preset, or drag to any value in range.

The field keeps its wire number. Stored JSON keeps no `previewWidth` alias, so
live content and version snapshots must be migrated by the API migration tool
before the API, Collab and Web that read `displaySize` start. All consumers must
use the same catalog fingerprint; older clients must reload.

## Wire numbers

Every message and enum is numbered densely from 1 (enums from 0) in declaration
order, with no `reserved` ranges or names. Removing a field or value renumbers
the remaining ones; the change is breaking and every consumer of the affected
message upgrades in lockstep. Web and API exchange settings and manifests as
Connect JSON (field names), so their numbers are not persisted. The API sends
`CollaborationConflictReason` to Collab as a binary error detail, so API and
Collab deploy together whenever that enum changes.

## Site logos

Management `PublicSettings` stores six logo slots exactly as configured:
`logo_full_light_asset` (17), `logo_full_dark_asset` (18),
`logo_email_light_asset` (19), `logo_symbol_light_asset` (27),
`logo_symbol_dark_asset` (28) and `logo_email_dark_asset` (29). Full and symbol
logos are SVG or PNG; email logos are PNG. The public manifest `SiteSettings`
exposes display logos resolved by the API: `logo_full_light_asset` (12),
`logo_full_dark_asset` (13), `logo_symbol_light_asset` (19) and
`logo_symbol_dark_asset` (20). A missing theme uses the other theme of the same
form and a missing form uses the other form. Email logos stay server-side.

## Release

Release Please versions both npm packages together and creates the Go module's
`v*` tag. The release workflow publishes `geul-proto` before `geul-event` to
GitHub Packages using `GITHUB_TOKEN` with `packages: write`. pnpm publication
converts workspace dependency references to package versions. The root package
remains unpublished (`private: true`); the two published manifests must remain
`private: false`, with restricted access and the GitHub Packages registry.

Owner/operator cutover order:

1. Create the organization Actions secret `PRIVATE_DEPS_TOKEN` and allow this
   repository and all consuming geul repositories to use it. Use one classic
   PAT with `repo` and `read:packages`, owned by an account with read access to
   the required private repositories and packages; authorize SSO if required.
   Cross-repository reads use this secret; publishing uses `GITHUB_TOKEN`.
2. Merge the distribution changes across producers and consumers, then switch
   the geul repositories to private before the first GitHub Packages release.
   Allow private package creation in the organization. Keep both packages
   private, link them to this repository, and enable repository permission
   inheritance (or explicitly grant this repository Actions write access).
   Grant the token's account and intended consumer teams package read access.
   Review existing packages separately: changing repository visibility does
   not replace checking package visibility and permissions.
3. Publish the versions required by consumers to GitHub Packages. Merging the
   next Release Please PR triggers publication in dependency order. If consumers
   require historical versions, backfill them separately from matching sources
   with the migrated registry/auth configuration; old tags still target npmjs
   and must not be published unchanged. This workflow does not backfill them.
   Publish other geul package producers before consumers that require them.
   Retain the Go module path and `v*` tags.
4. Update consumer registry/auth settings, the public-package tarball exceptions,
   and lockfiles before deploying consumers. Verify installs and releases on CI
   after cutover. Remove obsolete npmjs trusted-publisher settings in the package
   administration UI. Previously published npmjs versions and public Go proxy
   caches remain public; this change cannot revoke existing copies.

Dependabot currently updates only public GitHub Actions, so it needs no private
npm registry credentials. This repository has no Docker builds or Harbor images.
In consuming repositories, pass private dependency credentials to Docker builds
only through BuildKit `--secret` mounts, never token build arguments or image
environment variables. Harbor images remain at `registry.rawsector.com`; the
operator must provision authenticated pulls for the private project.

## License

PolyForm Noncommercial 1.0.0. Commercial use requires a separate license from
Echo Vision Lab. See [LICENSE.md](LICENSE.md).
