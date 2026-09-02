import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { parse } from "yaml";
import { describe, expect, it } from "vitest";
import { Queues } from "./index.ts";

type AsyncAPIChannel = {
  address?: string;
  messages?: Record<string, { $ref?: string }>;
  "x-channel-kind"?: "durable_queue" | "signal";
  "x-signal-types"?: string[];
};

const spec = parse(
  readFileSync(resolve(process.cwd(), "asyncapi/asyncapi.yaml"), "utf8"),
) as {
  channels: Record<string, AsyncAPIChannel>;
  components: { messages: Record<string, unknown> };
};

describe("durable asynchronous event model", () => {
  it("exports the authoritative result and projection queues", () => {
    expect(Queues).toMatchObject({
      transcodeResult: "transcode.result",
      waveformResult: "waveform.result",
      meshOptimizationResult: "mesh.optimization.result",
      releaseTrackOriginalAudioProjection:
        "release.track_original_audio.projection",
      emailCampaign: "email.campaign",
      userDeleteIdentity: "user.delete.identity",
      userDeleteAvatar: "user.delete.avatar",
    });

    for (const [channelName, expected] of Object.entries({
      transcodeResult: {
        queueName: "transcode.result",
        messageName: "transcodeCompleteEvent",
        messageRef: "#/components/messages/TranscodeCompleteEvent",
      },
      waveformResult: {
        queueName: "waveform.result",
        messageName: "waveformResultEvent",
        messageRef: "#/components/messages/WaveformResultEvent",
      },
      meshOptimizationResult: {
        queueName: "mesh.optimization.result",
        messageName: "meshOptimizationResultEvent",
        messageRef: "#/components/messages/MeshOptimizationResultEvent",
      },
      releaseTrackOriginalAudioProjection: {
        queueName: "release.track_original_audio.projection",
        messageName: "fileIngestAttachedEvent",
        messageRef: "#/components/messages/FileIngestAttachedEvent",
      },
      emailCampaign: {
        queueName: "email.campaign",
        messageName: "sendBulkEmailBatchEvent",
        messageRef: "#/components/messages/SendBulkEmailBatchEvent",
      },
      userDeleteIdentity: {
        queueName: "user.delete.identity",
        messageName: "userDeleteIdentityCommand",
        messageRef: "#/components/messages/UserDeleteIdentityCommand",
      },
      userDeleteAvatar: {
        queueName: "user.delete.avatar",
        messageName: "userDeleteAvatarCommand",
        messageRef: "#/components/messages/UserDeleteAvatarCommand",
      },
    })) {
      const channel = spec.channels[channelName];
      expect(channel?.address).toBe(expected.queueName);
      expect(channel?.messages).toEqual({
        [expected.messageName]: { $ref: expected.messageRef },
      });
      expect(channel?.["x-channel-kind"]).toBe("durable_queue");
    }
  });

  it("keeps file-ingest signals reconstructable and projection work durable", () => {
    expect(spec.channels.releaseTrackOriginalAudioProjection).toMatchObject({
      "x-channel-kind": "durable_queue",
    });
    expect(spec.channels.fileIngest).toMatchObject({
      "x-channel-kind": "signal",
      "x-signal-types": [
        "upload",
        "download",
        "finalized",
        "attached",
        "failed",
      ],
    });
  });

  it("keeps the retired generic render queues and messages absent", () => {
    for (const queue of [
      "post.render",
      "work.render",
      "program_event.render",
      "privacy.render",
      "terms.render",
    ]) {
      expect(Object.values(Queues)).not.toContain(queue);
    }
    for (const channel of [
      "postRender",
      "workRender",
      "programEventRender",
      "privacyRender",
      "termsRender",
    ]) {
      expect(Object.keys(spec.channels)).not.toContain(channel);
    }
    for (const message of [
      "PostRenderRequest",
      "WorkRenderRequest",
      "ProgramEventRenderRequest",
      "PrivacyRenderRequest",
      "TermsRenderRequest",
    ]) {
      expect(Object.keys(spec.components.messages)).not.toContain(message);
    }
  });
});
