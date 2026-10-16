import { fromJson, toJson, toBinary, fromBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  isSiteVariableKey,
  siteVariableKeys,
  normalizeContentStorageLocale,
  isRichTextCollaborativeTextPath,
  richTextProfiles,
} from "../../proto/gen/api/content/v1/block_catalog.ts";
import { RichTextBlockLocaleDataSchema } from "../../proto/gen/api/content/v1/block_content_pb.ts";

const normalize = (
  content: unknown[],
  profile: keyof typeof richTextProfiles = "post",
) =>
  normalizeContentStorageLocale(
    profile,
    "paragraph",
    fromJson(RichTextBlockLocaleDataSchema, {
      paragraph: { props: {}, content },
    } as never),
  );

describe("site variable contracts", () => {
  it("exports exactly the case-sensitive keys", () => {
    expect(siteVariableKeys).toEqual([
      "site.title",
      "site.url",
      "site.companyName",
      "contact.legalEmail",
      "contact.email",
      "contact.supportEmail",
      "contact.privacyEmail",
    ]);
    for (const key of siteVariableKeys)
      expect(isSiteVariableKey(key)).toBe(true);
    for (const key of ["", "site.Title", "contact.otherEmail"])
      expect(isSiteVariableKey(key)).toBe(false);
  });
  it("preserves distinct atoms, styles and binary/JSON across every prose profile", () => {
    for (const profile of Object.keys(
      richTextProfiles,
    ) as (keyof typeof richTextProfiles)[]) {
      const content = siteVariableKeys.flatMap((key) => [
        { siteVariable: { key, styles: { bold: true } } },
        { siteVariable: { key } },
      ]);
      const data = normalize(content, profile);
      expect(toJson(RichTextBlockLocaleDataSchema, data as never)).toEqual({
        paragraph: { props: {}, content },
      });
      expect(
        fromBinary(
          RichTextBlockLocaleDataSchema,
          toBinary(RichTextBlockLocaleDataSchema, data as never),
        ),
      ).toEqual(data);
    }
    expect(
      isRichTextCollaborativeTextPath(
        "paragraph",
        "content[0].siteVariable.key",
      ),
    ).toBe(false);
  });
  it("validates keys, linked eligibility and styles on write", () => {
    for (const key of siteVariableKeys) {
      const linked = () => normalize([{ siteVariable: { key, linked: true } }]);
      if (key === "site.title" || key === "site.companyName")
        expect(linked).toThrow(/cannot be linked/);
      else expect(linked).not.toThrow();
    }
    for (const key of ["", "site.Title", "other"])
      expect(() => normalize([{ siteVariable: { key } }])).toThrow(
        /unknown site variable key/,
      );
    expect(() =>
      normalize([
        { siteVariable: { key: "site.title", styles: { textColor: "bad" } } },
      ]),
    ).toThrow(/color/);
  });
  it("accepts root-relative links and rejects protocol-relative and unsafe targets", () => {
    for (const href of [
      "/",
      "/my",
      "/tools/hwp?x=1#a",
      "https://example.com",
      "mailto:a@example.com",
    ])
      expect(() =>
        normalize([{ link: { href, content: [{ text: "link" }] } }]),
      ).not.toThrow();
    for (const href of [
      "//x",
      "/\\x",
      "/ /x",
      "javascript:alert(1)",
      "relative",
    ])
      expect(() => normalize([{ link: { href, content: [] } }])).toThrow();
  });
});
