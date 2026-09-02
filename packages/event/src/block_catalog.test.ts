import { create, fromJson, toJson } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import {
  canonicalStorageDocumentBytes,
  contentBlockCatalogFingerprint,
  extractContentStorageFileReferences,
  flattenPageDocumentStorage,
  flattenPageMutationBatchStorage,
  flattenRichTextMutationBatchStorage,
  flattenRichTextDocumentStorage,
  isPageSectionCollaborativeTextPath,
  isRichTextCollaborativeTextPath,
  materializePageDocumentStorage,
  materializeRichTextDocumentStorage,
  normalizeContentStorageLocale,
  normalizeContentStorageBlock,
  normalizeContentStorageShared,
  pageSectionKinds,
  richTextBlockKinds,
  richTextBlockCatalog,
  richTextProfiles,
  validatePageSectionMutationBatch,
  validateRichTextBlockMutationBatch,
} from "../../proto/gen/api/content/v1/block_catalog.ts";
import {
  PageDocumentSchema,
  PageSectionDataSchema,
  PageSectionLocaleDataSchema,
  PageSectionMutationBatchSchema,
  ContentValidationMode,
  MissingAttachmentMediaKind,
  RichTextBlockLocaleDataSchema,
  RichTextBlockDataSchema,
  RichTextBlockMutationBatchSchema,
  RichTextDocumentSchema,
  RichTextProfile,
  ShaderProps_StagesItem_ChannelsItem_Kind,
  ShaderProps_StagesItem_Kind,
} from "../../proto/gen/api/content/v1/block_content_pb.ts";

const BLOCK_ID = "00000000-0000-4000-8000-000000000001";
const SECOND_BLOCK_ID = "00000000-0000-4000-8000-000000000003";
const SECTION_ID = "00000000-0000-4000-8000-000000000002";

describe("generated collaborative text paths", () => {
  it("keeps only authored rich text and code fields collaborative", () => {
    expect(
      isRichTextCollaborativeTextPath("paragraph", "content[0].text.text"),
    ).toBe(true);
    expect(
      isRichTextCollaborativeTextPath(
        "paragraph",
        "content[0].link.content[0].text",
      ),
    ).toBe(true);
    expect(
      isRichTextCollaborativeTextPath("paragraph", "content[0].link.href"),
    ).toBe(false);
    expect(
      isRichTextCollaborativeTextPath(
        "paragraph",
        "content[0].text.styles.textColor",
      ),
    ).toBe(false);
    expect(
      isRichTextCollaborativeTextPath("shader", "props.stages[4].source"),
    ).toBe(true);
    expect(
      isRichTextCollaborativeTextPath(
        "shader",
        "props.stages[4].channels[0].file",
      ),
    ).toBe(false);
    expect(isRichTextCollaborativeTextPath("file", "props.name")).toBe(true);
    expect(
      isRichTextCollaborativeTextPath("file", "props.attachment.activeFileId"),
    ).toBe(false);
  });

  it("covers Page nested authored strings without treating selectors as text", () => {
    expect(
      isPageSectionCollaborativeTextPath("text-marquee", "props.items[2].text"),
    ).toBe(true);
    expect(
      isPageSectionCollaborativeTextPath("text-marquee", "props.items[2].href"),
    ).toBe(false);
    expect(
      isPageSectionCollaborativeTextPath(
        "immersive-scene",
        "units[1].props.name",
      ),
    ).toBe(true);
    expect(
      isPageSectionCollaborativeTextPath(
        "immersive-scene",
        "units[1].props.title",
      ),
    ).toBe(true);
    expect(
      isPageSectionCollaborativeTextPath(
        "immersive-scene",
        "units[1].props.meshFile",
      ),
    ).toBe(false);
  });
});

describe("generated storage materializers", () => {
  it("round-trips Rich Text rows with canonical kebab-case kinds", () => {
    const document = create(RichTextDocumentSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      sourceLocale: "en",
      base: {
        nodes: [
          {
            block: {
              id: BLOCK_ID,
              value: { case: "codeBlock", value: { props: {} } },
            },
            placement: { index: 0 },
          },
        ],
      },
      localeOverlays: [
        {
          locale: "en",
          blocks: [
            {
              blockId: BLOCK_ID,
              value: {
                case: "codeBlock",
                value: { props: {}, content: "const answer = 42" },
              },
            },
          ],
        },
      ],
    });

    const rows = flattenRichTextDocumentStorage(document);
    expect(rows).toHaveLength(1);
    expect(rows[0]?.kind).toBe("code-block");

    const rebuilt = materializeRichTextDocumentStorage(
      RichTextProfile.POST,
      "en",
      rows,
    );
    expect(rebuilt.localeOverlays[0]?.blocks[0]?.value).toMatchObject({
      case: "codeBlock",
      value: { content: "const answer = 42" },
    });
    expect(
      canonicalStorageDocumentBytes(
        "post",
        flattenRichTextDocumentStorage(rebuilt),
      ),
    ).toEqual(canonicalStorageDocumentBytes("post", rows));
  });

  it("round-trips Page sections and their separately stored Rich Text descendants", () => {
    const document = create(PageDocumentSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      sourceLocale: "en",
      base: {
        nodes: [
          {
            section: {
              id: SECTION_ID,
              settings: {},
              value: {
                case: "richText",
                value: {
                  props: {},
                  blocks: {
                    nodes: [
                      {
                        block: {
                          id: BLOCK_ID,
                          value: { case: "codeBlock", value: { props: {} } },
                        },
                        placement: { index: 0 },
                      },
                      {
                        block: {
                          id: SECOND_BLOCK_ID,
                          value: { case: "paragraph", value: { props: {} } },
                        },
                        placement: { index: 1 },
                      },
                    ],
                  },
                },
              },
            },
            placement: { index: 0 },
          },
        ],
      },
      localeOverlays: [
        {
          locale: "en",
          sections: [
            {
              sectionId: SECTION_ID,
              value: {
                case: "richText",
                value: {
                  props: {},
                  blocks: {
                    locale: "en",
                    blocks: [
                      {
                        blockId: BLOCK_ID,
                        value: {
                          case: "codeBlock",
                          value: { props: {}, content: "nested code" },
                        },
                      },
                      {
                        blockId: SECOND_BLOCK_ID,
                        value: {
                          case: "paragraph",
                          value: {
                            props: {},
                            content: [
                              {
                                value: {
                                  case: "text",
                                  value: { text: "second nested block" },
                                },
                              },
                            ],
                          },
                        },
                      },
                    ],
                  },
                },
              },
            },
          ],
        },
      ],
    });

    const rows = flattenPageDocumentStorage(document);
    expect(rows.map((row) => row.kind)).toEqual([
      "rich-text",
      "code-block",
      "paragraph",
    ]);

    const rebuilt = materializePageDocumentStorage("en", rows);
    const nested = rebuilt.base?.nodes[0]?.section?.value;
    expect(nested?.case).toBe("richText");
    if (nested?.case !== "richText")
      throw new Error("expected Rich Text section");
    expect(nested.value.blocks?.nodes[0]?.block?.value.case).toBe("codeBlock");
    expect(nested.value.blocks?.nodes[1]?.block?.value.case).toBe("paragraph");
    expect(
      canonicalStorageDocumentBytes(
        "page",
        flattenPageDocumentStorage(rebuilt),
      ),
    ).toEqual(canonicalStorageDocumentBytes("page", rows));
  });

  it("uses the protobuf JSON name show3dBuildings in catalog and storage", () => {
    expect(richTextBlockCatalog.map.fields).toHaveProperty("show3dBuildings");
    expect(richTextBlockCatalog.map.fields).not.toHaveProperty(
      "show3DBuildings",
    );

    const document = create(RichTextDocumentSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      sourceLocale: "en",
      base: {
        nodes: [
          {
            block: {
              id: BLOCK_ID,
              value: {
                case: "map",
                value: { props: { show3dBuildings: true } },
              },
            },
            placement: { index: 0 },
          },
        ],
      },
      localeOverlays: [
        {
          locale: "en",
          blocks: [
            {
              blockId: BLOCK_ID,
              value: { case: "map", value: { props: {} } },
            },
          ],
        },
      ],
    });

    const row = flattenRichTextDocumentStorage(document)[0];
    expect(row).toBeDefined();
    if (row?.sharedData.$typeName !== "api.content.v1.RichTextBlockData") {
      throw new Error("expected Rich Text storage data");
    }
    const json = toJson(RichTextBlockDataSchema, row!.sharedData);
    expect(json).toMatchObject({ map: { props: { show3dBuildings: true } } });
    expect(JSON.stringify(json)).not.toContain("show3DBuildings");
  });

  it("keeps browser mutation validation on the attributed public surface", () => {
    const rich = create(RichTextBlockMutationBatchSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      expectedRevision: SECTION_ID,
      baseMutations: [
        { operation: { case: "delete", value: { blockId: BLOCK_ID } } },
      ],
    });
    expect(() => validateRichTextBlockMutationBatch(rich)).toThrow(
      /contributor Member/,
    );

    const page = create(PageSectionMutationBatchSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      expectedRevision: SECTION_ID,
      baseMutations: [
        { operation: { case: "delete", value: { sectionId: SECTION_ID } } },
      ],
    });
    expect(() => validatePageSectionMutationBatch(page)).toThrow(
      /contributor Member/,
    );
  });

  it("carries the locale ExpectedKind and validates it independently", () => {
    const batch = create(RichTextBlockMutationBatchSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      expectedRevision: SECTION_ID,
      contributorMemberIds: [SECTION_ID],
      localeMutationGroups: [
        {
          locale: "ko",
          mutations: [
            {
              operation: {
                case: "upsert",
                value: {
                  block: {
                    blockId: BLOCK_ID,
                    value: {
                      case: "codeBlock",
                      value: { props: {}, content: "const answer = 42" },
                    },
                  },
                },
              },
            },
          ],
        },
      ],
    });
    const storage = flattenRichTextMutationBatchStorage(batch);
    const upsert = storage.localeGroups[0]?.upserts[0];
    expect(upsert?.expectedKind).toBe("code-block");
    expect(() =>
      normalizeContentStorageLocale(
        "compact",
        "code-block",
        upsert!.localizedData,
      ),
    ).toThrow(/forbidden by profile/);
    const normalized = normalizeContentStorageLocale(
      "post",
      "code-block",
      upsert!.localizedData,
    );
    if (normalized.$typeName !== "api.content.v1.RichTextBlockLocaleData") {
      throw new Error("expected Rich Text locale storage data");
    }
    expect(toJson(RichTextBlockLocaleDataSchema, normalized)).toMatchObject({
      codeBlock: { content: "const answer = 42" },
    });
  });

  it("preserves present empty locale defaults without materializing missing values", () => {
    const pageBatch = (caption: string | undefined) =>
      create(PageSectionMutationBatchSchema, {
        blockCatalogFingerprint: contentBlockCatalogFingerprint,
        expectedRevision: SECTION_ID,
        contributorMemberIds: [SECTION_ID],
        localeMutationGroups: [
          {
            locale: "ko",
            mutations: [
              {
                operation: {
                  case: "upsert",
                  value: {
                    section: {
                      sectionId: SECTION_ID,
                      value: {
                        case: "externalVideo",
                        value: { props: { caption } },
                      },
                    },
                  },
                },
              },
            ],
          },
        ],
      });

    const explicitPage = flattenPageMutationBatchStorage(pageBatch(""));
    const missingPage = flattenPageMutationBatchStorage(pageBatch(undefined));
    const explicitPageData =
      explicitPage.localeGroups[0]!.upserts[0]!.localizedData;
    const missingPageData =
      missingPage.localeGroups[0]!.upserts[0]!.localizedData;
    if (
      explicitPageData.$typeName !== "api.content.v1.PageSectionLocaleData" ||
      missingPageData.$typeName !== "api.content.v1.PageSectionLocaleData"
    ) {
      throw new Error("expected Page locale storage data");
    }
    const explicitPageJSON = toJson(
      PageSectionLocaleDataSchema,
      explicitPageData,
    ) as { externalVideo?: { props?: { caption?: string } } };
    const missingPageJSON = toJson(
      PageSectionLocaleDataSchema,
      missingPageData,
    ) as { externalVideo?: { props?: { caption?: string } } };
    expect(
      Object.hasOwn(explicitPageJSON.externalVideo!.props!, "caption"),
    ).toBe(true);
    expect(explicitPageJSON.externalVideo!.props!.caption).toBe("");
    expect(
      Object.hasOwn(missingPageJSON.externalVideo!.props!, "caption"),
    ).toBe(false);

    const richBatch = (title: string | undefined) =>
      create(RichTextBlockMutationBatchSchema, {
        blockCatalogFingerprint: contentBlockCatalogFingerprint,
        profile: RichTextProfile.POST,
        expectedRevision: SECTION_ID,
        contributorMemberIds: [SECTION_ID],
        localeMutationGroups: [
          {
            locale: "ko",
            mutations: [
              {
                operation: {
                  case: "upsert",
                  value: {
                    block: {
                      blockId: BLOCK_ID,
                      value: {
                        case: "codeBlock",
                        value: { props: { title }, content: "" },
                      },
                    },
                  },
                },
              },
            ],
          },
        ],
      });
    const explicitRich = flattenRichTextMutationBatchStorage(richBatch(""));
    const missingRich = flattenRichTextMutationBatchStorage(
      richBatch(undefined),
    );
    const explicitRichData =
      explicitRich.localeGroups[0]!.upserts[0]!.localizedData;
    const missingRichData =
      missingRich.localeGroups[0]!.upserts[0]!.localizedData;
    if (
      explicitRichData.$typeName !== "api.content.v1.RichTextBlockLocaleData" ||
      missingRichData.$typeName !== "api.content.v1.RichTextBlockLocaleData"
    ) {
      throw new Error("expected Rich Text locale storage data");
    }
    const explicitRichJSON = toJson(
      RichTextBlockLocaleDataSchema,
      explicitRichData,
    ) as { codeBlock?: { props?: { title?: string } } };
    const missingRichJSON = toJson(
      RichTextBlockLocaleDataSchema,
      missingRichData,
    ) as { codeBlock?: { props?: { title?: string } } };
    expect(Object.hasOwn(explicitRichJSON.codeBlock!.props!, "title")).toBe(
      true,
    );
    expect(explicitRichJSON.codeBlock!.props!.title).toBe("");
    expect(Object.hasOwn(missingRichJSON.codeBlock!.props!, "title")).toBe(
      false,
    );
  });

  it("preserves nested Page units and empty Rich Text content shapes", () => {
    const nestedPage = create(PageSectionLocaleDataSchema, {
      value: {
        case: "immersiveScene",
        value: {
          props: {},
          units: [
            { unitId: BLOCK_ID, props: { title: "" } },
            { unitId: SECTION_ID, props: {} },
          ],
        },
      },
    });
    const normalizedPage = normalizeContentStorageLocale(
      "page",
      "immersive-scene",
      nestedPage,
    );
    if (
      normalizedPage.$typeName !== "api.content.v1.PageSectionLocaleData" ||
      normalizedPage.value.case !== "immersiveScene"
    ) {
      throw new Error("expected immersive Page locale storage data");
    }
    expect(
      Object.hasOwn(normalizedPage.value.value.units[0]!.props!, "title"),
    ).toBe(true);
    expect(normalizedPage.value.value.units[0]!.props!.title).toBe("");
    expect(
      Object.hasOwn(normalizedPage.value.value.units[1]!.props!, "title"),
    ).toBe(false);

    const pageDocument = create(PageDocumentSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      sourceLocale: "en",
      base: {
        nodes: [
          {
            section: {
              id: SECTION_ID,
              settings: {},
              value: {
                case: "immersiveScene",
                value: {
                  props: {},
                  units: [
                    { id: BLOCK_ID, props: {} },
                    { id: SECOND_BLOCK_ID, props: {} },
                  ],
                },
              },
            },
            placement: { index: 0 },
          },
        ],
      },
      localeOverlays: [
        {
          locale: "en",
          sections: [
            {
              sectionId: SECTION_ID,
              value: {
                case: "immersiveScene",
                value: {
                  props: {},
                  units: [
                    { unitId: BLOCK_ID, props: { title: "" } },
                    { unitId: SECOND_BLOCK_ID, props: {} },
                  ],
                },
              },
            },
          ],
        },
      ],
    });
    const pageRow = flattenPageDocumentStorage(pageDocument)[0];
    const pageLocaleData = pageRow?.locales[0]?.data;
    if (
      pageRow?.sharedData.$typeName !== "api.content.v1.PageSectionData" ||
      pageLocaleData?.$typeName !== "api.content.v1.PageSectionLocaleData"
    ) {
      throw new Error("expected Page storage row");
    }
    const pageSharedJSON = toJson(
      PageSectionDataSchema,
      pageRow.sharedData,
    ) as { immersiveScene?: { props?: { textureSize?: string } } };
    expect(
      Object.hasOwn(pageSharedJSON.immersiveScene!.props!, "textureSize"),
    ).toBe(false);
    const pageLocaleJSON = toJson(
      PageSectionLocaleDataSchema,
      pageLocaleData,
    ) as {
      immersiveScene?: {
        units?: { props?: { title?: string } }[];
      };
    };
    expect(
      Object.hasOwn(pageLocaleJSON.immersiveScene!.units![0]!.props!, "title"),
    ).toBe(true);
    expect(
      Object.hasOwn(pageLocaleJSON.immersiveScene!.units![1]!.props!, "title"),
    ).toBe(false);

    const table = create(RichTextBlockLocaleDataSchema, {
      value: {
        case: "table",
        value: {
          props: {},
          content: {
            rows: [
              {
                rowId: BLOCK_ID,
                cells: [{ cellId: SECTION_ID, content: [] }],
              },
            ],
          },
        },
      },
    });
    const normalizedTable = normalizeContentStorageLocale(
      "post",
      "table",
      table,
    );
    if (
      normalizedTable.$typeName !== "api.content.v1.RichTextBlockLocaleData" ||
      normalizedTable.value.case !== "table"
    ) {
      throw new Error("expected Table locale storage data");
    }
    expect(normalizedTable.value.value.content?.rows).toHaveLength(1);
    expect(
      normalizedTable.value.value.content?.rows[0]?.cells[0]?.content,
    ).toEqual([]);
  });

  it("validates descriptor-owned inline styles across locale storage and mutations", () => {
    const malformed = create(RichTextBlockLocaleDataSchema, {
      value: {
        case: "paragraph",
        value: {
          props: {},
          content: [
            {
              value: {
                case: "link",
                value: {
                  href: "https://geul.io",
                  content: [
                    {
                      text: "invalid",
                      styles: { backgroundColor: "RGB" },
                    },
                  ],
                },
              },
            },
          ],
        },
      },
    });
    expect(() =>
      normalizeContentStorageLocale("post", "paragraph", malformed),
    ).toThrow(/invalid editor color/);

    const malformedTable = create(RichTextBlockLocaleDataSchema, {
      value: {
        case: "table",
        value: {
          props: {},
          content: {
            rows: [
              {
                rowId: BLOCK_ID,
                cells: [
                  {
                    cellId: SECTION_ID,
                    content: [
                      {
                        value: {
                          case: "text",
                          value: {
                            text: "invalid",
                            styles: { textColor: "RGB" },
                          },
                        },
                      },
                    ],
                  },
                ],
              },
            ],
          },
        },
      },
    });
    expect(() =>
      normalizeContentStorageLocale("post", "table", malformedTable),
    ).toThrow(/invalid editor color/);

    const valid = create(RichTextBlockLocaleDataSchema, {
      value: {
        case: "paragraph",
        value: {
          props: {},
          content: [
            {
              value: {
                case: "text",
                value: {
                  text: "valid",
                  styles: { bold: true, textColor: "#A1B2C3" },
                },
              },
            },
          ],
        },
      },
    });
    const normalized = normalizeContentStorageLocale(
      "post",
      "paragraph",
      valid,
    );
    if (normalized.$typeName !== "api.content.v1.RichTextBlockLocaleData") {
      throw new Error("expected Rich Text locale storage data");
    }
    expect(toJson(RichTextBlockLocaleDataSchema, normalized)).toMatchObject({
      paragraph: {
        content: [{ text: { styles: { bold: true, textColor: "#a1b2c3" } } }],
      },
    });

    const inlineMath = create(RichTextBlockLocaleDataSchema, {
      value: {
        case: "paragraph",
        value: {
          props: {},
          content: [
            { value: { case: "mathInline", value: { source: "x^2" } } },
          ],
        },
      },
    });
    expect(() =>
      normalizeContentStorageLocale("post", "paragraph", inlineMath),
    ).not.toThrow();
    expect(() =>
      normalizeContentStorageLocale("page", "paragraph", inlineMath),
    ).toThrow(/inline math is forbidden/);

    const mutation = create(RichTextBlockMutationBatchSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      expectedRevision: SECTION_ID,
      contributorMemberIds: [SECTION_ID],
      localeMutationGroups: [
        {
          locale: "ko",
          mutations: [
            {
              operation: {
                case: "upsert",
                value: {
                  block: {
                    blockId: BLOCK_ID,
                    value: malformed.value,
                  },
                },
              },
            },
          ],
        },
      ],
    });
    expect(() => validateRichTextBlockMutationBatch(mutation)).toThrow(
      /invalid editor color/,
    );
    expect(() => flattenRichTextMutationBatchStorage(mutation)).toThrow(
      /invalid editor color/,
    );
  });

  it("uses WRITE versus RESTORE_SNAPSHOT while flattening storage", () => {
    const document = create(RichTextDocumentSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      sourceLocale: "en",
      base: {
        nodes: [
          {
            block: {
              id: BLOCK_ID,
              value: {
                case: "file",
                value: {
                  props: {
                    attachment: {
                      state: {
                        case: "missingAttachment",
                        value: {
                          formerFileId: SECTION_ID,
                          mediaKind: MissingAttachmentMediaKind.FILE,
                        },
                      },
                    },
                  },
                },
              },
            },
            placement: { index: 0 },
          },
        ],
      },
      localeOverlays: [
        {
          locale: "en",
          blocks: [
            {
              blockId: BLOCK_ID,
              value: { case: "file", value: { props: {} } },
            },
          ],
        },
      ],
    });
    expect(() => flattenRichTextDocumentStorage(document)).toThrow(
      /snapshot-restore only/,
    );
    expect(
      flattenRichTextDocumentStorage(
        document,
        ContentValidationMode.RESTORE_SNAPSHOT,
      ),
    ).toHaveLength(1);
  });

  it("extracts exact active and missing storage attachment selectors", () => {
    const active = create(RichTextBlockDataSchema, {
      value: {
        case: "file",
        value: {
          props: {
            attachment: {
              state: { case: "activeFileId", value: BLOCK_ID },
            },
          },
        },
      },
    });
    expect(extractContentStorageFileReferences("file", active)).toEqual([
      {
        referencePath: "file",
        fileId: BLOCK_ID,
        missing: false,
        missingAttachmentMediaKind: MissingAttachmentMediaKind.UNSPECIFIED,
        allowedMimeTypes: [],
        allowedMimePrefixes: [],
      },
    ]);

    const shader = create(RichTextBlockDataSchema, {
      value: {
        case: "shader",
        value: {
          props: {
            stages: [
              {
                channels: [
                  {
                    file: {
                      state: {
                        case: "missingAttachment",
                        value: {
                          formerFileId: SECTION_ID,
                          mediaKind: MissingAttachmentMediaKind.VIDEO,
                        },
                      },
                    },
                    faces: [
                      {
                        state: { case: "activeFileId", value: BLOCK_ID },
                      },
                    ],
                  },
                ],
              },
            ],
          },
        },
      },
    });
    expect(extractContentStorageFileReferences("shader", shader)).toMatchObject(
      [
        {
          referencePath: "shader.stages.0.channels.0.faces.0",
          missing: false,
          missingAttachmentMediaKind: MissingAttachmentMediaKind.UNSPECIFIED,
        },
        {
          referencePath: "shader.stages.0.channels.0.file",
          missing: true,
          missingAttachmentMediaKind: MissingAttachmentMediaKind.VIDEO,
        },
      ],
    );

    const immersive = create(PageSectionDataSchema, {
      value: {
        case: "immersiveScene",
        value: {
          units: [
            {
              id: SECTION_ID,
              props: {
                textureFile: {
                  state: {
                    case: "missingAttachment",
                    value: {
                      formerFileId: BLOCK_ID,
                      mediaKind: MissingAttachmentMediaKind.IMAGE,
                    },
                  },
                },
              },
            },
          ],
        },
      },
    });
    expect(
      extractContentStorageFileReferences("immersive-scene", immersive),
    ).toMatchObject([
      {
        referencePath: `immersive_scene:${SECTION_ID}:texture`,
        missing: true,
        missingAttachmentMediaKind: MissingAttachmentMediaKind.IMAGE,
      },
    ]);

    const invalid = create(RichTextBlockDataSchema, {
      value: {
        case: "file",
        value: {
          props: {
            attachment: {
              state: {
                case: "missingAttachment",
                value: { formerFileId: BLOCK_ID },
              },
            },
          },
        },
      },
    });
    expect(() => extractContentStorageFileReferences("file", invalid)).toThrow(
      /media kind is required/,
    );
  });

  it("normalizes shared storage for every profile and catalog kind", () => {
    const protoCase = (kind: string) =>
      kind.replace(/-([a-z])/g, (_, letter: string) => letter.toUpperCase());
    const validRichShared = (kind: string) => {
      if (kind === "file")
        return create(RichTextBlockDataSchema, {
          value: {
            case: "file",
            value: {
              props: {
                attachment: {
                  state: { case: "activeFileId", value: BLOCK_ID },
                },
              },
            },
          },
        });
      if (kind === "shader")
        return create(RichTextBlockDataSchema, {
          value: {
            case: "shader",
            value: {
              props: {
                stages: Array.from({ length: 9 }, () => ({
                  kind: ShaderProps_StagesItem_Kind.COMMON,
                  source: "",
                  channels: Array.from({ length: 4 }, () => ({
                    kind: ShaderProps_StagesItem_ChannelsItem_Kind.NONE,
                  })),
                })),
              },
            },
          },
        });
      return fromJson(RichTextBlockDataSchema, {
        [protoCase(kind)]: { props: {} },
      });
    };
    const validPageShared = (kind: string) => {
      const props: Record<string, unknown> = {};
      if (kind === "external-video") props.uri = "https://example.com/video";
      if (kind === "form") props.formId = BLOCK_ID;
      if (kind === "columns")
        props.columns = [
          { id: BLOCK_ID, ratio: 1 },
          { id: SECTION_ID, ratio: 1 },
        ];
      return fromJson(PageSectionDataSchema, {
        [protoCase(kind)]: { props },
      } as never);
    };

    for (const [profile, definition] of Object.entries(richTextProfiles)) {
      for (const kind of richTextBlockKinds) {
        const normalize = () =>
          normalizeContentStorageShared(
            profile as keyof typeof richTextProfiles,
            kind,
            validRichShared(kind),
          );
        if ((definition.blocks as readonly string[]).includes(kind))
          expect(normalize, `${profile}/${kind}`).not.toThrow();
        else expect(normalize, `${profile}/${kind}`).toThrow(/forbidden/);
      }
    }
    for (const kind of pageSectionKinds)
      expect(
        () =>
          normalizeContentStorageShared("page", kind, validPageShared(kind)),
        `page/${kind}`,
      ).not.toThrow();
  });

  it("owns canonical shared validation, modes, refs, and flattener reuse", () => {
    const upperFileId = BLOCK_ID.toUpperCase();
    const active = create(RichTextBlockDataSchema, {
      value: {
        case: "file",
        value: {
          props: {
            attachment: {
              state: { case: "activeFileId", value: upperFileId },
            },
            previewWidth: 100,
          },
        },
      },
    });
    const shared = normalizeContentStorageShared("post", "file", active);
    if (shared.sharedData.$typeName !== "api.content.v1.RichTextBlockData")
      throw new Error("expected rich-text shared storage data");
    expect(toJson(RichTextBlockDataSchema, shared.sharedData)).toEqual({
      file: { props: { attachment: { activeFileId: BLOCK_ID } } },
    });
    const locale = create(RichTextBlockLocaleDataSchema, {
      value: { case: "file", value: { props: {} } },
    });
    const full = normalizeContentStorageBlock("post", "file", active, locale);
    expect(shared.fileReferences).toEqual(full.fileReferences);
    expect(shared.fileReferences[0]).toMatchObject({
      fileId: BLOCK_ID,
      missing: false,
      missingAttachmentMediaKind: MissingAttachmentMediaKind.UNSPECIFIED,
    });

    const missing = create(RichTextBlockDataSchema, {
      value: {
        case: "file",
        value: {
          props: {
            attachment: {
              state: {
                case: "missingAttachment",
                value: {
                  formerFileId: SECTION_ID,
                  mediaKind: MissingAttachmentMediaKind.VIDEO,
                },
              },
            },
          },
        },
      },
    });
    expect(() =>
      normalizeContentStorageShared("post", "file", missing),
    ).toThrow(/restore.only/);
    const restored = normalizeContentStorageShared(
      "post",
      "file",
      missing,
      ContentValidationMode.RESTORE_SNAPSHOT,
    );
    expect(restored.fileReferences[0]).toMatchObject({
      fileId: SECTION_ID,
      missing: true,
      missingAttachmentMediaKind: MissingAttachmentMediaKind.VIDEO,
    });

    const runtime = fromJson(RichTextBlockDataSchema, {
      paragraph: { props: {} },
    });
    Object.assign(runtime.value.value!.props!, {
      url: "https://runtime.invalid",
    });
    expect(() =>
      normalizeContentStorageShared("post", "paragraph", runtime),
    ).toThrow(/runtime projection field/);
    expect(() =>
      normalizeContentStorageShared(
        "post",
        "unknown",
        fromJson(RichTextBlockDataSchema, { paragraph: { props: {} } }),
      ),
    ).toThrow(/kind/);
    expect(() =>
      normalizeContentStorageShared(
        "post",
        "file",
        fromJson(RichTextBlockDataSchema, { file: { props: {} } }),
      ),
    ).toThrow(/required/);

    const mutation = create(RichTextBlockMutationBatchSchema, {
      blockCatalogFingerprint: contentBlockCatalogFingerprint,
      profile: RichTextProfile.POST,
      expectedRevision: SECTION_ID,
      contributorMemberIds: [SECTION_ID],
      baseMutations: [
        {
          operation: {
            case: "upsert",
            value: {
              node: {
                block: { id: BLOCK_ID, value: missing.value },
                placement: { index: 0 },
              },
            },
          },
        },
      ],
    });
    expect(() => flattenRichTextMutationBatchStorage(mutation)).toThrow(
      /restore.only/,
    );
    const flattened = flattenRichTextMutationBatchStorage(
      mutation,
      ContentValidationMode.RESTORE_SNAPSHOT,
    );
    expect(flattened.baseUpserts[0]?.sharedData).toEqual(restored.sharedData);
  });
});
