import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  UserDeleteIdentityCommandSchema,
  UserDeleteIdentityMode,
} from "@echovisionlab/geul-proto/secure/events_pb.ts";

describe("UserDeleteIdentityCommand mode", () => {
  it("keeps the zero value invalid instead of selecting a cleanup mode", () => {
    const command = create(UserDeleteIdentityCommandSchema, {
      memberId: "member-1",
      identityId: "identity-1",
    });

    expect(command.mode).toBe(UserDeleteIdentityMode.UNSPECIFIED);
    expect(command.mode).not.toBe(UserDeleteIdentityMode.TOMBSTONE);
    expect(command.mode).not.toBe(
      UserDeleteIdentityMode.UNONBOARDED_HARD_DELETE,
    );
  });

  it.each([
    UserDeleteIdentityMode.TOMBSTONE,
    UserDeleteIdentityMode.UNONBOARDED_HARD_DELETE,
  ])("round-trips explicit mode %s", (mode) => {
    const command = create(UserDeleteIdentityCommandSchema, {
      mode,
      memberId: "member-1",
      identityId: "identity-1",
    });

    expect(command.mode).toBe(mode);
  });
});
