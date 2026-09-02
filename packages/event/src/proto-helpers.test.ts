import { create, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  OgGenerationLifecycleEventSchema,
  UserDeleteIdentityCommandSchema,
} from "@echovisionlab/geul-proto/secure/events_pb.ts";
import {
  deserializeProto,
  FileIngestAttachedEventSchema,
  FileIngestIdentitySchema,
  FileIngestMediaKind,
  TranscodeEntityType,
} from "./proto-helpers.ts";

describe("deserializeProto", () => {
  it("decodes a generated protobuf message", () => {
    const source = create(OgGenerationLifecycleEventSchema, {
      generationId: "generation-1",
      runId: "run-1",
    });

    expect(
      deserializeProto(
        OgGenerationLifecycleEventSchema,
        toBinary(OgGenerationLifecycleEventSchema, source),
      ),
    ).toEqual(source);
  });

  it("preserves Track first-attach absence and replacement CAS identity", () => {
    const firstAttach = create(FileIngestIdentitySchema, {
      entityType: TranscodeEntityType.TRACK,
      entityId: "track-1",
      fileId: "file-new",
      mediaKind: FileIngestMediaKind.TRACK_AUDIO,
    });
    const firstDecoded = deserializeProto(
      FileIngestIdentitySchema,
      toBinary(FileIngestIdentitySchema, firstAttach),
    );
    expect(firstDecoded.expectedCurrentFileId).toBeUndefined();

    const replacement = create(FileIngestIdentitySchema, {
      entityType: TranscodeEntityType.TRACK,
      entityId: "track-1",
      fileId: "file-new",
      mediaKind: FileIngestMediaKind.TRACK_AUDIO,
      expectedCurrentFileId: "file-current",
    });
    const replacementDecoded = deserializeProto(
      FileIngestIdentitySchema,
      toBinary(FileIngestIdentitySchema, replacement),
    );
    expect(replacementDecoded.expectedCurrentFileId).toBe("file-current");
  });

  it.each([
    { name: "first attach", expectedCurrentFileId: undefined },
    { name: "replacement", expectedCurrentFileId: "file-current" },
  ])("round-trips a $name attached event", ({ expectedCurrentFileId }) => {
    const source = create(FileIngestAttachedEventSchema, {
      correlationId: `correlation-${expectedCurrentFileId ?? "first"}`,
      identity: {
        entityType: TranscodeEntityType.TRACK,
        entityId: "track-1",
        fileId: "file-new",
        mediaKind: FileIngestMediaKind.TRACK_AUDIO,
        expectedCurrentFileId,
      },
      sequenceNumber: 1n,
      timestampMs: 1782864000000n,
      fileName: "field-recording.wav",
      mimeType: "audio/wav",
      fileSize: 1024n,
    });

    const decoded = deserializeProto(
      FileIngestAttachedEventSchema,
      toBinary(FileIngestAttachedEventSchema, source),
    );

    expect(decoded).toEqual(source);
    expect(decoded.identity?.expectedCurrentFileId).toBe(expectedCurrentFileId);
  });

  it("preserves the immutable account-deletion identity/member pair", () => {
    const source = create(UserDeleteIdentityCommandSchema, {
      memberId: "14600000-0000-0000-0000-000000000001",
      identityId: "14600000-0000-0000-0000-000000000003",
      avatarAssetId: "14600000-0000-0000-0000-000000000002",
      notificationEmail: "member@example.test",
      notificationName: "Member",
      notificationLocale: "ko",
    });

    expect(
      deserializeProto(
        UserDeleteIdentityCommandSchema,
        toBinary(UserDeleteIdentityCommandSchema, source),
      ),
    ).toEqual(source);
  });
});
