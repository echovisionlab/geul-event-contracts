import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { CreateCampaignRequestSchema } from "@echovisionlab/geul-proto/secure/campaign_pb.ts";

describe("CreateCampaignRequest target wire contract", () => {
  it.each([
    {
      name: "all recipients",
      target: { case: "all" as const, value: {} },
    },
    {
      name: "one exact audience segment",
      target: { case: "segmentId" as const, value: "segment-1" },
    },
  ])("round-trips the explicit $name target", ({ target }) => {
    const request = create(CreateCampaignRequestSchema, {
      subject: "Subject",
      name: "Campaign",
      target,
    });
    const decoded = fromBinary(
      CreateCampaignRequestSchema,
      toBinary(CreateCampaignRequestSchema, request),
    );

    expect(decoded.target.case).toBe(target.case);
    if (target.case === "all") {
      expect(decoded.target).toMatchObject({
        case: "all",
        value: { $typeName: "google.protobuf.Empty" },
      });
    } else {
      expect(decoded.target).toEqual(target);
    }
  });
});
