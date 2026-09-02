import { describe, expect, it } from "vitest";
import { AIDocumentDomain } from "@echovisionlab/geul-proto/secure/ai_pb.ts";
import { ContentEntityType } from "@echovisionlab/geul-proto/secure/events_pb.ts";

describe("ContentEntityType interactive document parity", () => {
  it("covers every registered AI document domain with one exact typed value", () => {
    const pairs = [
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_POST, ContentEntityType.POST],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_PAGE, ContentEntityType.PAGE],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_WORK, ContentEntityType.WORK],
      [
        AIDocumentDomain.AI_DOCUMENT_DOMAIN_PROGRAM_EVENT,
        ContentEntityType.PROGRAM_EVENT,
      ],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_RELEASE, ContentEntityType.RELEASE],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_ARTIST, ContentEntityType.ARTIST],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_LABEL, ContentEntityType.LABEL],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_MENU, ContentEntityType.MENU],
      [
        AIDocumentDomain.AI_DOCUMENT_DOMAIN_EMAIL_TEMPLATE,
        ContentEntityType.EMAIL_TEMPLATE,
      ],
      [
        AIDocumentDomain.AI_DOCUMENT_DOMAIN_EMAIL_LAYOUT,
        ContentEntityType.EMAIL_LAYOUT,
      ],
      [
        AIDocumentDomain.AI_DOCUMENT_DOMAIN_CAMPAIGN,
        ContentEntityType.CAMPAIGN,
      ],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_FORM, ContentEntityType.FORM],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_PRIVACY, ContentEntityType.PRIVACY],
      [AIDocumentDomain.AI_DOCUMENT_DOMAIN_TERMS, ContentEntityType.TERMS],
      [
        AIDocumentDomain.AI_DOCUMENT_DOMAIN_POST_SERIES,
        ContentEntityType.POST_SERIES,
      ],
    ] as const;

    expect(new Set(pairs.map(([domain]) => domain)).size).toBe(15);
    expect(new Set(pairs.map(([, content]) => content)).size).toBe(15);
    expect(
      Object.values(AIDocumentDomain).filter(
        (value): value is number =>
          typeof value === "number" &&
          value !== AIDocumentDomain.AI_DOCUMENT_DOMAIN_UNSPECIFIED,
      ),
    ).toHaveLength(15);
    expect(
      Object.values(ContentEntityType).filter(
        (value): value is number =>
          typeof value === "number" && value !== ContentEntityType.UNSPECIFIED,
      ),
    ).toHaveLength(15);
  });

  it("preserves existing numbers and allocates the six additive values", () => {
    expect(ContentEntityType.POST).toBe(1);
    expect(ContentEntityType.PAGE).toBe(2);
    expect(ContentEntityType.WORK).toBe(3);
    expect(ContentEntityType.ARTIST).toBe(4);
    expect(ContentEntityType.LABEL).toBe(5);
    expect(ContentEntityType.RELEASE).toBe(6);
    expect(ContentEntityType.FORM).toBe(7);
    expect(ContentEntityType.PROGRAM_EVENT).toBe(8);
    expect(ContentEntityType.POST_SERIES).toBe(9);
    expect(ContentEntityType.MENU).toBe(10);
    expect(ContentEntityType.EMAIL_TEMPLATE).toBe(11);
    expect(ContentEntityType.EMAIL_LAYOUT).toBe(12);
    expect(ContentEntityType.CAMPAIGN).toBe(13);
    expect(ContentEntityType.PRIVACY).toBe(14);
    expect(ContentEntityType.TERMS).toBe(15);
  });
});
