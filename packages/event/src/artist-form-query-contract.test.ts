import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { ListFormSubmissionsRequestSchema } from "@echovisionlab/geul-proto/secure/form_pb.ts";
import { ArtistWorkSchema } from "@echovisionlab/geul-proto/public/artist_pb.ts";
import { WorkType } from "@echovisionlab/geul-proto/public/work_pb.ts";
import { TimestampSchema } from "@bufbuild/protobuf/wkt";

describe("artist and form query contracts", () => {
  it("round-trips every artist work type", () => {
    const types = [
      WorkType.MUSIC_PROJECT,
      WorkType.PORTFOLIO,
      WorkType.ARTICLE,
      WorkType.CONTRIBUTION,
    ];

    for (const type of types) {
      const work = create(ArtistWorkSchema, { id: "work-1", type });
      const roundTripped = fromBinary(
        ArtistWorkSchema,
        toBinary(ArtistWorkSchema, work),
      );

      expect(roundTripped.type).toBe(type);
    }
  });

  it("round-trips all form submission filters with timestamp presence", () => {
    const request = create(ListFormSubmissionsRequestSchema, {
      formId: "form-1",
      search: "a%_\\b",
      countryCode: "US",
      createdAtFrom: create(TimestampSchema, { seconds: 1_800_000_000n }),
      createdAtBefore: create(TimestampSchema, { seconds: 1_800_003_600n }),
    });
    const roundTripped = fromBinary(
      ListFormSubmissionsRequestSchema,
      toBinary(ListFormSubmissionsRequestSchema, request),
    );

    expect(roundTripped.search).toBe("a%_\\b");
    expect(roundTripped.countryCode).toBe("US");
    expect(roundTripped.createdAtFrom?.seconds).toBe(1_800_000_000n);
    expect(roundTripped.createdAtBefore?.seconds).toBe(1_800_003_600n);
  });
});
