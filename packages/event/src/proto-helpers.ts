import { fromBinary, type Message } from "@bufbuild/protobuf";
import type { GenMessage } from "@bufbuild/protobuf/codegenv2";

export type {
  AssetRef,
  AssetWriteResult,
  AssetWriteTarget,
} from "@echovisionlab/geul-proto/common/media_pb.ts";

export {
  AssetDisposition,
  MediaProcessingStatus,
} from "@echovisionlab/geul-proto/common/media_pb.ts";

export type {
  FileIngestAttachedEvent,
  FileIngestDownloadEvent,
  FileIngestFailedEvent,
  FileIngestFinalizedEvent,
  FileIngestIdentity,
  FileIngestProgress,
  FileIngestUploadEvent,
  MediaProcessingLifecycleEvent,
  MediaProcessingLifecycleOutputs,
  MeshOptimizationResultEvent,
  OgGenerationJob,
  OgGenerationLifecycleEvent,
  OgGenerationTarget,
  TranscodeCompleteEvent,
  TranslationLifecycleEvent,
  WaveformResultEvent,
} from "@echovisionlab/geul-proto/secure/events_pb.ts";

export {
  FileIngestAttachedEventSchema,
  FileIngestDownloadEventSchema,
  FileIngestFailedEventSchema,
  FileIngestFailureReason,
  FileIngestFinalizedEventSchema,
  FileIngestIdentitySchema,
  FileIngestMediaKind,
  FileIngestSource,
  FileIngestUploadEventSchema,
  MediaProcessingLifecycleEventSchema,
  MeshOptimizationResultEventSchema,
  OgEntityType,
  OgGenerationJobSchema,
  OgGenerationLifecycleEventSchema,
  OgGenerationStatus,
  OgGenerationTargetSchema,
  TranscodeCompleteEventSchema,
  TranscodeEntityType,
  TranslationLifecycleEventSchema,
  WaveformResultEventSchema,
} from "@echovisionlab/geul-proto/secure/events_pb.ts";

export function deserializeProto<T extends Message>(
  schema: GenMessage<T>,
  data: Uint8Array,
): T {
  return fromBinary(schema, data);
}
