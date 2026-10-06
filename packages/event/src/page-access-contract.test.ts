import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  PageAccessMode,
  PageAccessMatch,
  PageAccessReason,
} from "@echovisionlab/geul-proto/common/page_access_pb.ts";
import { AuthorizationRole } from "@echovisionlab/geul-proto/policy/access_pb.ts";
import {
  PageSchema as ManagePageSchema,
  PageSummarySchema,
  CreatePageRequestSchema,
  UpdatePageRequestSchema,
  UpdatePageResponseSchema,
} from "@echovisionlab/geul-proto/secure/page_pb.ts";
import {
  PageSchema,
  GetPageResponseSchema,
} from "@echovisionlab/geul-proto/public/page_pb.ts";

describe("Page access contract", () => {
  it("appends audience fields without changing Page field tags", () => {
    for (const [schema, tag] of [
      [ManagePageSchema, 17],
      [PageSummarySchema, 10],
      [CreatePageRequestSchema, 6],
      [UpdatePageRequestSchema, 4],
      [UpdatePageResponseSchema, 6],
      [PageSchema, 16],
    ] as const) {
      expect(schema.fields.at(-1)?.name).toBe("access_policy");
      expect(schema.fields.map((field) => field.number)).toEqual(
        Array.from({ length: tag }, (_, index) => index + 1),
      );
    }
    expect(
      GetPageResponseSchema.fields.map((field) => [field.name, field.number]),
    ).toEqual([
      ["page", 1],
      ["block_media", 2],
      ["access_reason", 3],
    ]);
  });

  it("preserves policy omission and typed role/tag/newsletter conditions", () => {
    const policies = [
      undefined,
      {
        mode: PageAccessMode.CONDITIONS,
        allowedRoles: [
          AuthorizationRole.USER,
          AuthorizationRole.AUTHOR,
          AuthorizationRole.ADMIN,
        ],
        userTagIds: ["tag-1", "tag-2"],
        newsletterSubscriber: true,
        match: PageAccessMatch.ALL,
      },
    ];
    for (const accessPolicy of policies) {
      const request = create(UpdatePageRequestSchema, {
        id: "page-1",
        accessPolicy,
      });
      const decoded = fromBinary(
        UpdatePageRequestSchema,
        toBinary(UpdatePageRequestSchema, request),
      );
      expect(decoded).toEqual(request);
      expect(decoded.accessPolicy?.mode).toBe(accessPolicy?.mode);
    }
  });

  it("round-trips a denied response with no Page content or media", () => {
    const response = create(GetPageResponseSchema, {
      accessReason: PageAccessReason.AUTHENTICATION_REQUIRED,
    });
    const decoded = fromBinary(
      GetPageResponseSchema,
      toBinary(GetPageResponseSchema, response),
    );
    expect(decoded).toEqual(response);
    expect(decoded.page).toBeUndefined();
    expect(decoded.blockMedia).toEqual([]);
  });
});
