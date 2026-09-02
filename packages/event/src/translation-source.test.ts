import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  TranslationEntityType,
  TranslationJobSchema,
} from "@echovisionlab/geul-proto/secure/translation_pb.ts";
import { assertTranslationJobRequestArtifact } from "./translation-source.ts";

describe("translation request artifact", () => {
  it("requires one target and canonical request artifact digest", () => {
    const job = create(TranslationJobSchema);
    expect(() => assertTranslationJobRequestArtifact(job)).toThrow(
      "translation job target is required",
    );

    job.target = {
      $typeName: "api.manage.v1.TranslationTarget",
      entityType: TranslationEntityType.PAGE,
      entityId: "entity-1",
    };
    expect(() => assertTranslationJobRequestArtifact(job)).toThrow(
      "translation job request artifact digest is required",
    );

    job.requestArtifactDigest = "artifact-digest";
    expect(() => assertTranslationJobRequestArtifact(job)).not.toThrow();
  });
});
