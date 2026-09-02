import { describe, expect, it } from "vitest";
import {
  createPgmqEnvelope,
  pgmqEnvelopePayload,
  PGMQ_ENVELOPE_SCHEMA_VERSION,
  PgmqClient,
  type PgmqExecutor,
} from "./pgmq.ts";

describe("PGMQ transport contract", () => {
  it("preserves protobuf bytes without ProtoJSON conversion", () => {
    const payload = Uint8Array.from([0, 255, 1, 128, 127]);
    const envelope = createPgmqEnvelope({
      messageId: "command-1",
      messageType: "api.manage.v1.Command",
      payload,
      createdAt: new Date("2026-08-14T00:00:00.000Z"),
    });
    expect(envelope.schema_version).toBe(PGMQ_ENVELOPE_SCHEMA_VERSION);
    expect(pgmqEnvelopePayload(envelope)).toEqual(payload);
  });

  it("uses caller-owned executors for atomic enqueue", async () => {
    const calls: Array<{ text: string; values?: readonly unknown[] }> = [];
    const transaction: PgmqExecutor = {
      query: (async (text: string, values?: readonly unknown[]) => {
        calls.push({ text, values });
        return { rows: [{ msg_id: "42" }] };
      }) as PgmqExecutor["query"],
    };
    const envelope = createPgmqEnvelope({
      messageId: "command-1",
      messageType: "type",
      payload: new Uint8Array(),
    });
    await expect(
      new PgmqClient().enqueue(transaction, "work.queue", envelope),
    ).resolves.toBe(42n);
    expect(calls).toHaveLength(1);
    expect(calls[0]?.values?.[0]).toBe("work.queue");
  });

  it("returns malformed rows with transport identity so consumers can archive them", async () => {
    const executor: PgmqExecutor = {
      query: (async () => ({
        rows: [
          {
            msg_id: "41",
            read_ct: 1,
            enqueued_at: new Date("2026-08-14T00:00:00.000Z"),
            vt: new Date("2026-08-14T00:00:30.000Z"),
            message: { schema_version: 999 },
            headers: {},
          },
          {
            msg_id: "42",
            read_ct: 1,
            enqueued_at: new Date("2026-08-14T00:00:00.000Z"),
            vt: new Date("2026-08-14T00:00:30.000Z"),
            message: createPgmqEnvelope({
              messageId: "command-2",
              messageType: "type",
              payload: new Uint8Array(),
            }),
            headers: {},
          },
        ],
      })) as PgmqExecutor["query"],
    };

    const messages = await new PgmqClient().read(executor, "work.queue", {
      visibilityTimeoutSeconds: 30,
      batch: 2,
    });
    expect(messages).toHaveLength(2);
    expect(messages[0]).toMatchObject({ transportId: 41n });
    expect(messages[0]?.contractError).toContain("message_id");
    expect(messages[1]?.envelope?.message_id).toBe("command-2");
  });
});
