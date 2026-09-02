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

## Release

Release Please versions both npm packages together. npm publication uses
GitHub Actions trusted publishing; no npm token is stored in the repository.

## License

PolyForm Noncommercial 1.0.0. Commercial use requires a separate license from
Echo Vision Lab. See [LICENSE.md](LICENSE.md).
