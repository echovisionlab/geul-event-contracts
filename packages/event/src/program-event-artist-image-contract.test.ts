import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import { ProgramEventArtistSchema } from "@echovisionlab/geul-proto/public/program_event_pb.ts";

describe("public event artist image contract", () => {
  it("appends the image asset after stable artist field tags", () => {
    expect(
      ProgramEventArtistSchema.fields.map((field) => [
        field.name,
        field.number,
      ]),
    ).toEqual([
      ["id", 1],
      ["name", 2],
      ["slug", 3],
      ["role", 4],
      ["sort_order", 5],
      ["image_asset", 6],
    ]);
  });

  it.each([
    undefined,
    {
      assetId: "artist-image",
      url: "https://cdn.example.com/artist.webp",
      mimeType: "image/webp",
      extension: "webp",
    },
  ])(
    "preserves image presence and artist identity on the wire",
    (imageAsset) => {
      const artist = create(ProgramEventArtistSchema, {
        id: "artist-1",
        name: "Artist",
        slug: "artist",
        role: "Performer",
        sortOrder: 2,
        imageAsset,
      });
      const decoded = fromBinary(
        ProgramEventArtistSchema,
        toBinary(ProgramEventArtistSchema, artist),
      );
      expect(decoded).toEqual(artist);
      expect(decoded.imageAsset?.assetId).toBe(imageAsset?.assetId);
    },
  );
});
