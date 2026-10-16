# RPC consumer inventory

[RPC inventory](rpc-inventory.json) records the production callers of each
protobuf service procedure. Each verified caller identifies its repository,
source file, function, exact client call and runtime entrypoint. Consumer and
provider tests are recorded separately.

Run `pnpm inventory:sync` after changing RPC declarations, then inspect the
production callers and update their evidence. New entries start as `pending`;
mark them `verified` after reading the caller and its entrypoint.

`pnpm check:contracts` checks the descriptor-defined procedure set, duplicate
entries, evidence fields and completion status. It leaves caller verification
to the maintainer.

The editor first-paint RPCs `GetPostEditorDocument` and
`GetPageEditorDocument` are implemented by geul-api and consumed by geul-web.
The Post edit route and Page edit route call their authorized snapshot query
helpers and pass the returned snapshots to the editors for the initial server
render. Their inventory entries record these production callers and the
consumer/provider tests separately. The audit is complete with no pending
procedures. See
[the coordinated launch](../../geul-api/docs/collaboration-durable-launch.md).
