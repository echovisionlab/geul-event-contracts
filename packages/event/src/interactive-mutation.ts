import { AIDocumentDomain } from "@echovisionlab/geul-proto/secure/ai_pb.ts";
import {
  InteractiveAIDocumentMutationOrigin,
  type RelayInteractiveAIDocumentMutationRequest,
} from "@echovisionlab/geul-proto/intra/collaboration_pb.ts";

const activeDocumentDomains = new Set<number>(
  Object.values(AIDocumentDomain).filter(
    (value): value is number =>
      typeof value === "number" &&
      value !== AIDocumentDomain.AI_DOCUMENT_DOMAIN_UNSPECIFIED,
  ),
);

function requireNonBlank(value: string, name: string): void {
  if (value.trim() === "") {
    throw new Error(`interactive AI document mutation ${name} is required`);
  }
}

// Transport-independent validation shared by the post-commit API publisher and
// the Collab relay receiver. It creates no retry, history, or persistence.
export function assertRelayInteractiveAIDocumentMutationRequest(
  event: RelayInteractiveAIDocumentMutationRequest,
): void {
  requireNonBlank(event.mutationId, "id");

  switch (event.origin) {
    case InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_IN_EDITOR_AI:
    case InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP:
      break;
    case InteractiveAIDocumentMutationOrigin.INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_UNSPECIFIED:
    default:
      throw new Error(
        `unsupported interactive AI document mutation origin: ${event.origin}`,
      );
  }

  if (event.document === undefined) {
    throw new Error("interactive AI document mutation document is required");
  }
  if (!activeDocumentDomains.has(event.document.domain)) {
    throw new Error(
      `unsupported interactive AI document mutation domain: ${event.document.domain}`,
    );
  }
  requireNonBlank(event.document.reference, "reference");

  if (event.locale === undefined) {
    throw new Error("interactive AI document mutation locale is required");
  }
  requireNonBlank(event.locale.code, "locale");
  requireNonBlank(event.expectedDocumentRevision, "expected document revision");
  requireNonBlank(event.acceptedDocumentRevision, "accepted document revision");

  if (event.operations.length === 0) {
    throw new Error("interactive AI document mutation operations are required");
  }
  event.operations.forEach((operation, index) => {
    if (operation.operation.case === undefined) {
      throw new Error(
        `interactive AI document mutation operation ${index} is required`,
      );
    }
  });
  if (
    event.operations.length !== 1 &&
    event.operations.some(
      (operation) =>
        operation.operation.case === "createTranslation" ||
        operation.operation.case === "deleteTranslation",
    )
  ) {
    throw new Error(
      "interactive AI document mutation locale lifecycle operation must be exclusive",
    );
  }

  if (event.expectedTargetRevision !== undefined) {
    requireNonBlank(event.expectedTargetRevision, "expected target revision");
  }
  if (event.acceptedTargetRevision !== undefined) {
    requireNonBlank(event.acceptedTargetRevision, "accepted target revision");
  }
  const isTargetMutation =
    event.expectedTargetRevision !== undefined ||
    event.acceptedTargetRevision !== undefined;
  if (isTargetMutation) {
    if (event.acceptedDocumentRevision !== event.expectedDocumentRevision) {
      throw new Error(
        "interactive AI document target mutation must preserve the document revision",
      );
    }
  } else {
    if (
      event.expectedTargetRevision === undefined &&
      event.acceptedTargetRevision === undefined &&
      event.acceptedDocumentRevision === event.expectedDocumentRevision
    ) {
      throw new Error(
        "interactive AI document source mutation must advance the document revision",
      );
    }
  }

  const onlyOperation =
    event.operations.length === 1 ? event.operations[0] : undefined;
  if (onlyOperation?.operation.case === "createTranslation") {
    if (
      event.expectedTargetRevision !== undefined ||
      event.acceptedTargetRevision === undefined
    ) {
      throw new Error(
        "interactive AI document mutation create translation revisions are invalid",
      );
    }
  } else if (onlyOperation?.operation.case === "deleteTranslation") {
    if (
      event.expectedTargetRevision === undefined ||
      event.acceptedTargetRevision !== undefined
    ) {
      throw new Error(
        "interactive AI document mutation delete translation revisions are invalid",
      );
    }
  } else if (
    event.expectedTargetRevision !== undefined &&
    event.acceptedTargetRevision === undefined
  ) {
    throw new Error(
      "interactive AI document mutation target revision cannot disappear",
    );
  }

  requireNonBlank(event.actorMemberId, "actor member id");
}
