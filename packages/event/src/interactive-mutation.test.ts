import { create } from "@bufbuild/protobuf";
import {
  AIDocumentCreateTranslationOperationSchema,
  AIDocumentDeleteTranslationOperationSchema,
  AIDocumentDeleteBlockOperationSchema,
  AIDocumentDomain,
  AIDocumentLocaleSchema,
  AIDocumentOperationSchema,
  AIDocumentReferenceSchema,
} from "@echovisionlab/geul-proto/secure/ai_pb.ts";
import {
  InteractiveAIDocumentMutationOrigin,
  RelayInteractiveAIDocumentMutationRequestSchema,
} from "@echovisionlab/geul-proto/intra/collaboration_pb.ts";
import { describe, expect, it } from "vitest";
import { assertRelayInteractiveAIDocumentMutationRequest } from "./interactive-mutation.ts";

function validEvent() {
  return create(RelayInteractiveAIDocumentMutationRequestSchema, {
    mutationId: "mutation-1",
    origin:
      InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP,
    document: create(AIDocumentReferenceSchema, {
      domain: AIDocumentDomain.AI_DOCUMENT_DOMAIN_POST,
      reference: "post-1",
    }),
    locale: create(AIDocumentLocaleSchema, { code: "ko" }),
    expectedDocumentRevision: "revision-before",
    acceptedDocumentRevision: "revision-after",
    operations: [
      create(AIDocumentOperationSchema, {
        operation: {
          case: "deleteBlock",
          value: create(AIDocumentDeleteBlockOperationSchema, {
            blockHandle: "block-1",
          }),
        },
      }),
    ],
    actorMemberId: "member-1",
  });
}

describe("interactive AI document mutation relay", () => {
  it.each([
    InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_IN_EDITOR_AI,
    InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP,
  ])("accepts active interactive origin %s", (origin) => {
    const event = validEvent();
    event.origin = origin;
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(event),
    ).not.toThrow();
  });

  it.each([
    {
      name: "blank mutation id",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.mutationId = "  ";
      },
      error: "id",
    },
    {
      name: "unspecified origin",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.origin =
          InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_UNSPECIFIED;
      },
      error: "origin",
    },
    {
      name: "missing document",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.document = undefined;
      },
      error: "document",
    },
    {
      name: "unknown origin",
      mutate: (event: ReturnType<typeof validEvent>) => {
        // @ts-expect-error Simulate an unknown numeric enum decoded from the wire.
        event.origin = 99;
      },
      error: "origin",
    },
    {
      name: "unspecified domain",
      mutate: (event: ReturnType<typeof validEvent>) => {
        if (event.document === undefined) {
          throw new Error("test fixture document is required");
        }
        event.document.domain = AIDocumentDomain.AI_DOCUMENT_DOMAIN_UNSPECIFIED;
      },
      error: "domain",
    },
    {
      name: "unknown domain",
      mutate: (event: ReturnType<typeof validEvent>) => {
        if (event.document === undefined) {
          throw new Error("test fixture document is required");
        }
        // @ts-expect-error Simulate an unknown numeric enum decoded from the wire.
        event.document.domain = 99;
      },
      error: "domain",
    },
    {
      name: "blank reference",
      mutate: (event: ReturnType<typeof validEvent>) => {
        if (event.document === undefined) {
          throw new Error("test fixture document is required");
        }
        event.document.reference = "  ";
      },
      error: "reference",
    },
    {
      name: "missing locale",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.locale = undefined;
      },
      error: "locale",
    },
    {
      name: "blank locale",
      mutate: (event: ReturnType<typeof validEvent>) => {
        if (event.locale === undefined) {
          throw new Error("test fixture locale is required");
        }
        event.locale.code = "  ";
      },
      error: "locale",
    },
    {
      name: "blank expected document revision",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.expectedDocumentRevision = "  ";
      },
      error: "expected document revision",
    },
    {
      name: "blank accepted document revision",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.acceptedDocumentRevision = "  ";
      },
      error: "accepted document revision",
    },
    {
      name: "blank expected target revision",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.expectedTargetRevision = "  ";
      },
      error: "expected target revision",
    },
    {
      name: "blank accepted target revision",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.acceptedTargetRevision = "  ";
      },
      error: "accepted target revision",
    },
    {
      name: "target revision disappears outside delete",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.expectedTargetRevision = "target-before";
        event.acceptedDocumentRevision = event.expectedDocumentRevision;
      },
      error: "cannot disappear",
    },
    {
      name: "source mutation preserves document revision",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.acceptedDocumentRevision = event.expectedDocumentRevision;
      },
      error: "must advance",
    },
    {
      name: "missing operations",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.operations = [];
      },
      error: "operations",
    },
    {
      name: "missing operation arm",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.operations = [create(AIDocumentOperationSchema)];
      },
      error: "operation 0",
    },
    {
      name: "mixed locale lifecycle operation",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.operations.push(
          create(AIDocumentOperationSchema, {
            operation: {
              case: "createTranslation",
              value: create(AIDocumentCreateTranslationOperationSchema),
            },
          }),
        );
      },
      error: "must be exclusive",
    },
    {
      name: "blank actor member id",
      mutate: (event: ReturnType<typeof validEvent>) => {
        event.actorMemberId = "  ";
      },
      error: "actor member id",
    },
  ])("rejects $name", ({ mutate, error }) => {
    const event = validEvent();
    mutate(event);
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(event),
    ).toThrow(error);
  });

  it("accepts existing-target and missing-target value transitions", () => {
    const existing = validEvent();
    existing.acceptedDocumentRevision = existing.expectedDocumentRevision;
    existing.expectedTargetRevision = "target-before";
    existing.acceptedTargetRevision = "target-after";
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(existing),
    ).not.toThrow();

    const firstWrite = validEvent();
    firstWrite.acceptedDocumentRevision = firstWrite.expectedDocumentRevision;
    firstWrite.acceptedTargetRevision = "target-created";
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(firstWrite),
    ).not.toThrow();
  });

  it("accepts a source mutation with an advanced document revision", () => {
    const event = validEvent();
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(event),
    ).not.toThrow();
  });

  it("accepts explicit target create and delete transitions", () => {
    const createTarget = validEvent();
    createTarget.acceptedDocumentRevision =
      createTarget.expectedDocumentRevision;
    createTarget.acceptedTargetRevision = "target-created";
    createTarget.operations = [
      create(AIDocumentOperationSchema, {
        operation: {
          case: "createTranslation",
          value: create(AIDocumentCreateTranslationOperationSchema),
        },
      }),
    ];
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(createTarget),
    ).not.toThrow();

    const deleteTarget = validEvent();
    deleteTarget.acceptedDocumentRevision =
      deleteTarget.expectedDocumentRevision;
    deleteTarget.expectedTargetRevision = "target-before";
    deleteTarget.operations = [
      create(AIDocumentOperationSchema, {
        operation: {
          case: "deleteTranslation",
          value: create(AIDocumentDeleteTranslationOperationSchema),
        },
      }),
    ];
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(deleteTarget),
    ).not.toThrow();
  });

  it("rejects target changes that replace the shared document token", () => {
    const event = validEvent();
    event.expectedTargetRevision = "target-before";
    event.acceptedTargetRevision = "target-after";
    expect(() =>
      assertRelayInteractiveAIDocumentMutationRequest(event),
    ).toThrow("preserve the document revision");
  });
});
