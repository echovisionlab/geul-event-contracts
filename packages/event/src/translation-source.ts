import type { TranslationJob } from "@echovisionlab/geul-proto/secure/translation_pb.ts";

/**
 * Validates the one request identity shared by every translation-capable
 * domain. Domain-specific source epochs, hashes, and revision counters are
 * intentionally absent from this boundary.
 */
export function assertTranslationJobRequestArtifact(job: TranslationJob): void {
  if (job.target === undefined) {
    throw new Error("translation job target is required");
  }
  if (job.requestArtifactDigest.trim().length === 0) {
    throw new Error("translation job request artifact digest is required");
  }
}
