export const PGMQ_ENVELOPE_SCHEMA_VERSION = 1 as const;

export type PgmqEnvelope = {
  message_id: string;
  message_type: string;
  schema_version: typeof PGMQ_ENVELOPE_SCHEMA_VERSION;
  created_at: string;
  correlation_id?: string;
  causation_id?: string;
  payload_base64: string;
};

export type PgmqHeaders = Record<string, string>;

type QueryResult<Row> = { rows: Row[] };

export interface PgmqExecutor {
  query<Row extends Record<string, unknown>>(
    text: string,
    values?: readonly unknown[],
  ): Promise<QueryResult<Row>>;
}

export type PgmqMessage = {
  transportId: bigint;
  readCount: number;
  enqueuedAt: Date;
  visibleAt: Date;
  envelope?: PgmqEnvelope;
  headers: PgmqHeaders;
  /** Present when the row was read but violates the shared transport contract. */
  contractError?: string;
};

function encodeBase64(payload: Uint8Array): string {
  return Buffer.from(payload).toString("base64");
}

function decodeBase64(payload: string): Uint8Array {
  return new Uint8Array(Buffer.from(payload, "base64"));
}

export function createPgmqEnvelope(input: {
  messageId: string;
  messageType: string;
  payload: Uint8Array;
  correlationId?: string;
  causationId?: string;
  createdAt?: Date;
}): PgmqEnvelope {
  const envelope: PgmqEnvelope = {
    message_id: input.messageId,
    message_type: input.messageType,
    schema_version: PGMQ_ENVELOPE_SCHEMA_VERSION,
    created_at: (input.createdAt ?? new Date()).toISOString(),
    payload_base64: encodeBase64(input.payload),
    ...(input.correlationId ? { correlation_id: input.correlationId } : {}),
    ...(input.causationId ? { causation_id: input.causationId } : {}),
  };
  assertPgmqEnvelope(envelope);
  return envelope;
}

export function assertPgmqEnvelope(
  value: unknown,
): asserts value is PgmqEnvelope {
  if (!value || typeof value !== "object") {
    throw new Error("PGMQ envelope must be an object");
  }
  const envelope = value as Partial<PgmqEnvelope>;
  if (!envelope.message_id || typeof envelope.message_id !== "string") {
    throw new Error("PGMQ envelope message_id is required");
  }
  if (!envelope.message_type || typeof envelope.message_type !== "string") {
    throw new Error("PGMQ envelope message_type is required");
  }
  if (envelope.schema_version !== PGMQ_ENVELOPE_SCHEMA_VERSION) {
    throw new Error(
      `Unsupported PGMQ envelope schema ${envelope.schema_version}`,
    );
  }
  if (
    typeof envelope.created_at !== "string" ||
    Number.isNaN(Date.parse(envelope.created_at))
  ) {
    throw new Error("PGMQ envelope created_at must be an ISO timestamp");
  }
  if (typeof envelope.payload_base64 !== "string") {
    throw new Error("PGMQ envelope payload_base64 is required");
  }
  const normalized = Buffer.from(envelope.payload_base64, "base64").toString(
    "base64",
  );
  if (normalized !== envelope.payload_base64) {
    throw new Error("PGMQ envelope payload_base64 is invalid");
  }
}

export function pgmqEnvelopePayload(envelope: PgmqEnvelope): Uint8Array {
  assertPgmqEnvelope(envelope);
  return decodeBase64(envelope.payload_base64);
}

export class PgmqClient {
  async enqueue(
    executor: PgmqExecutor,
    queue: string,
    envelope: PgmqEnvelope,
    options: { headers?: PgmqHeaders; delaySeconds?: number } = {},
  ): Promise<bigint> {
    assertPgmqEnvelope(envelope);
    if (!queue) throw new Error("PGMQ queue is required");
    const delaySeconds = Math.max(0, Math.trunc(options.delaySeconds ?? 0));
    const result = await executor.query<{ msg_id: string | number | bigint }>(
      "SELECT pgmq.send($1, $2::jsonb, $3::jsonb, $4::integer) AS msg_id",
      [queue, envelope, options.headers ?? {}, delaySeconds],
    );
    const row = result.rows[0];
    if (!row) throw new Error(`PGMQ send returned no message for ${queue}`);
    return BigInt(row.msg_id);
  }

  async read(
    executor: PgmqExecutor,
    queue: string,
    options: { visibilityTimeoutSeconds: number; batch?: number },
  ): Promise<PgmqMessage[]> {
    if (!queue) throw new Error("PGMQ queue is required");
    const visibilityTimeoutSeconds = Math.trunc(
      options.visibilityTimeoutSeconds,
    );
    const batch = Math.trunc(options.batch ?? 1);
    if (visibilityTimeoutSeconds <= 0) {
      throw new Error("PGMQ visibility timeout must be positive");
    }
    if (batch <= 0) throw new Error("PGMQ batch must be positive");
    const result = await executor.query<{
      msg_id: string | number | bigint;
      read_ct: number;
      enqueued_at: Date | string;
      vt: Date | string;
      message: unknown;
      headers: PgmqHeaders | null;
    }>(
      "SELECT msg_id, read_ct, enqueued_at, vt, message, headers FROM pgmq.read($1, $2::integer, $3::integer, '{}'::jsonb)",
      [queue, visibilityTimeoutSeconds, batch],
    );
    return result.rows.map((row) => {
      const message = {
        transportId: BigInt(row.msg_id),
        readCount: row.read_ct,
        enqueuedAt: new Date(row.enqueued_at),
        visibleAt: new Date(row.vt),
        headers: row.headers ?? {},
      };
      try {
        assertPgmqEnvelope(row.message);
        return { ...message, envelope: row.message };
      } catch (error) {
        return {
          ...message,
          contractError:
            error instanceof Error ? error.message : "Invalid PGMQ envelope",
        };
      }
    });
  }

  async complete(
    executor: PgmqExecutor,
    queue: string,
    transportId: bigint,
  ): Promise<void> {
    await this.booleanOperation(executor, "delete", queue, transportId);
  }

  async retry(
    executor: PgmqExecutor,
    queue: string,
    transportId: bigint,
    delaySeconds: number,
  ): Promise<void> {
    const result = await executor.query<{ msg_id: string | number | bigint }>(
      "SELECT msg_id FROM pgmq.set_vt($1, $2::bigint, $3::integer)",
      [queue, transportId.toString(), Math.max(0, Math.trunc(delaySeconds))],
    );
    if (BigInt(result.rows[0]?.msg_id ?? -1) !== transportId) {
      throw new Error(`PGMQ retry did not update ${queue}/${transportId}`);
    }
  }

  async deadLetter(
    executor: PgmqExecutor,
    queue: string,
    transportId: bigint,
  ): Promise<void> {
    await this.booleanOperation(executor, "archive", queue, transportId);
  }

  private async booleanOperation(
    executor: PgmqExecutor,
    operation: "delete" | "archive",
    queue: string,
    transportId: bigint,
  ): Promise<void> {
    const result = await executor.query<{ succeeded: boolean }>(
      `SELECT pgmq.${operation}($1, $2::bigint) AS succeeded`,
      [queue, transportId.toString()],
    );
    if (result.rows[0]?.succeeded !== true) {
      throw new Error(
        `PGMQ ${operation} did not update ${queue}/${transportId}`,
      );
    }
  }
}
