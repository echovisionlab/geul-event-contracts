# RPC Consumer Audit

## Status

The audit is `complete`. `docs/rpc-inventory.json` is the authoritative
per-procedure status and records direct production consumer evidence for all
492 live RPCs. Oathkeeper's `remote_json` authorizer is the production consumer of the private
`api.intra.v1.InternalGatewayAuthorizationService/AuthorizeGatewayAccess`
procedure. Its configured full procedure path and the exercised authentication
boundary provide the same direct HTTP consumer evidence as other non-generated
service integrations.

The completion pass removed four provider-only procedures together with their
dead request/response graph: AI Document read/validate and the Privacy/Terms
internal render loaders. It also found and repaired two live collaboration
consumers whose Campaign and Email Template locale-metadata paths did not match
their declared procedure names.

The protobuf descriptors provide only the exhaustive list of RPCs. They do not
prove that an RPC is used. Automated source scans must not set
`consumerStatus: verified`.

## Verification rule

An RPC is verified only after its production consumer has been read directly
and the following evidence is recorded for that exact procedure:

- `repository`: repository that owns the caller.
- `file`: production source file containing the call.
- `symbol`: function, method, handler, worker, or hook that makes the call.
- `call`: exact generated-client method or full Connect procedure invocation.
- `entrypoint`: production path that reaches the caller, including scheduler,
  route, worker bootstrap, or UI/server action as applicable.

A generated client, imported service descriptor, same-named method, test-only
call, implementation method, or gateway policy rule is not a consumer.

If direct inspection finds no production caller, the RPC is removed from the
proto together with generated code, API implementation, gateway policy, and
dead request/response messages. A live RPC is never marked `unused` and kept for
compatibility.

## Test rule

Consumer and provider tests are recorded only after consumer verification.
They are separate evidence and cannot justify retaining an otherwise unused
RPC.

## Automation boundary

The contract checker may only:

- enumerate the exact descriptor-defined service and procedure set;
- reject missing, stale, or duplicate inventory entries;
- validate the required fields of manually verified evidence;
- reject `consumerAuditStatus: complete` while any procedure is pending.

It may not search repositories for likely call names, infer consumers, or
promote a procedure from `pending` to `verified`.
