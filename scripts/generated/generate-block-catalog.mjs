import { execFileSync } from "node:child_process";
import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import YAML from "yaml";

const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const sourcePath = path.join(root, "config/content/block-catalog.yaml");
const protoPath = path.join(root, "proto/api/content/v1/block_content.proto");
const goPath = path.join(root, "gen/api/content/v1/block_catalog.go");
const tsPath = path.join(
  root,
  "packages/proto/gen/api/content/v1/block_catalog.ts",
);
const check = process.argv.includes("--check");
const protoOnly = process.argv.includes("--proto-only");

const source = YAML.parse(fs.readFileSync(sourcePath, "utf8"));
const own = new Set(["shared", "locale", "source"]);
const namePattern = /^[a-z][A-Za-z0-9-]*$/;
const kindPattern = /^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$/;
const profilePattern = /^[a-z][a-z0-9_]*$/;

function assert(condition, message) {
  if (!condition) throw new Error(`block catalog: ${message}`);
}

assert(source?.schema_version === 1, "schema_version must be 1");
assert(Array.isArray(source.ownerships), "ownerships must be an array");
assert(
  source.ownerships.every((value) => own.has(value)),
  "unknown ownership",
);
assert(source.rich_text?.blocks, "rich_text.blocks is required");
assert(source.page?.sections, "page.sections is required");
assert(source.profiles, "profiles is required");

function mergedFields(definition) {
  const result = {};
  for (const setName of definition.field_sets ?? []) {
    assert(source.field_sets[setName], `unknown field set ${setName}`);
    Object.assign(result, structuredClone(source.field_sets[setName]));
  }
  Object.assign(result, structuredClone(definition.fields ?? {}));
  return result;
}

function validateField(field, fieldPath, inheritedOwnership) {
  assert(field && typeof field === "object", `${fieldPath} must be an object`);
  const ownership = field.ownership ?? inheritedOwnership;
  assert(own.has(ownership), `${fieldPath} must declare ownership`);
  const types = new Set([
    "array",
    "boolean",
    "editor_color",
    "enum",
    "enum_int",
    "file_attachment",
    "hex_color",
    "integer",
    "number",
    "object",
    "string",
    "uri",
    "uuid",
  ]);
  assert(
    types.has(field.type),
    `${fieldPath} has unsupported type ${field.type}`,
  );
  assert(
    !field.type.endsWith("_string"),
    `${fieldPath} must use a canonical scalar, not a string encoding`,
  );
  if (field.type === "enum" || field.type === "enum_int") {
    assert(
      Array.isArray(field.values) && field.values.length > 0,
      `${fieldPath} enum needs values`,
    );
    assert(
      new Set(field.values.map(String)).size === field.values.length,
      `${fieldPath} enum values must be unique`,
    );
  }
  if (field.type === "array") {
    assert(field.items, `${fieldPath} array needs items`);
    const identity = field.item_identity;
    if (identity)
      assert(
        ["field", "fixed", "position", "value"].includes(identity.strategy),
        `${fieldPath} has unsupported item identity strategy`,
      );
    if (identity?.strategy === "field") {
      assert(
        field.items.type === "object" &&
          typeof identity.field === "string" &&
          field.items.fields?.[identity.field],
        `${fieldPath} field identity must reference an item field`,
      );
    }
    if (identity?.strategy === "fixed") {
      assert(
        Array.isArray(identity.values) &&
          identity.values.length > 0 &&
          new Set(identity.values).size === identity.values.length,
        `${fieldPath} fixed identity values must be non-empty and unique`,
      );
      assert(
        field.exact_length === identity.values.length,
        `${fieldPath} fixed identity values must match exact_length`,
      );
    }
    if (identity?.strategy === "position") {
      assert(
        identity.mutation === "whole-list-replace",
        `${fieldPath} positional identity requires whole-list-replace`,
      );
    }
    if (identity?.strategy === "value") {
      assert(
        field.items.type !== "array" && field.items.type !== "object",
        `${fieldPath} value identity requires scalar items`,
      );
    }
    validateField(
      { ...field.items, ownership: field.items.ownership ?? ownership },
      `${fieldPath}[]`,
      ownership,
    );
  }
  if (field.type === "object") {
    assert(
      field.fields && typeof field.fields === "object",
      `${fieldPath} object needs fields`,
    );
    for (const [name, nested] of Object.entries(field.fields)) {
      validateField(
        { ...nested, ownership: nested.ownership ?? ownership },
        `${fieldPath}.${name}`,
        ownership,
      );
    }
  }
  if (field.translatable)
    assert(
      ownership === "locale",
      `${fieldPath} translatable field must be locale-owned`,
    );
}

function validateDefinition(definition, definitionPath) {
  const fields = mergedFields(definition);
  for (const [name, field] of Object.entries(fields)) {
    assert(
      namePattern.test(name),
      `${definitionPath}.${name} has an invalid name`,
    );
    validateField(field, `${definitionPath}.${name}`);
  }
  return { ...definition, fields };
}

const richBlocks = Object.fromEntries(
  Object.entries(source.rich_text.blocks).map(([name, definition]) => {
    assert(
      kindPattern.test(name),
      `rich-text Block kind must be canonical kebab-case: ${name}`,
    );
    return [name, validateDefinition(definition, `rich_text.blocks.${name}`)];
  }),
);
const pageSections = Object.fromEntries(
  Object.entries(source.page.sections).map(([name, definition]) => {
    assert(
      kindPattern.test(name),
      `Page section kind must be canonical kebab-case: ${name}`,
    );
    return [name, validateDefinition(definition, `page.sections.${name}`)];
  }),
);
const pageImmersiveUnit = validateDefinition(
  source.page.immersive_unit,
  "page.immersive_unit",
);

for (const [profile, definition] of Object.entries(source.profiles)) {
  assert(profilePattern.test(profile), `invalid profile name ${profile}`);
  assert(
    Array.isArray(definition.blocks),
    `${profile}.blocks must be an array`,
  );
  assert(
    definition.blocks.every((kind) => richBlocks[kind]),
    `${profile} references an unknown block kind`,
  );
}
assert(
  source.profiles.page.inline_math === false &&
    !source.profiles.page.blocks.includes("map"),
  "page must forbid map and inline math",
);
for (const profile of ["policy", "email"]) {
  const forbidden = new Set([
    "file",
    "p5-sketch",
    "three-scene",
    "shader",
    "map",
    "math",
  ]);
  assert(
    source.profiles[profile].blocks.every((kind) => !forbidden.has(kind)) &&
      source.profiles[profile].inline_math === false,
    `${profile} must forbid media, executable, map, and math content`,
  );
}

const normalized = {
  ...source,
  rich_text: { ...source.rich_text, blocks: richBlocks },
  page: { ...source.page, sections: pageSections },
};
assert(
  JSON.stringify(Object.keys(source.profiles)) ===
    JSON.stringify([
      "post",
      "page",
      "work",
      "program_event",
      "compact",
      "email",
      "policy",
    ]),
  "profiles must use the canonical behavior vocabulary",
);
const fingerprint = crypto
  .createHash("sha256")
  .update(JSON.stringify(normalized))
  .digest("hex");

const pascal = (value) =>
  String(value)
    .replaceAll(/([a-z0-9])([A-Z])/g, "$1 $2")
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((part) => part[0].toUpperCase() + part.slice(1))
    .join("");
const snake = (value) =>
  (String(value) === "show3dBuildings" ? "show_3d_buildings" : String(value))
    .replaceAll(/([a-z0-9])([A-Z])/g, "$1_$2")
    .replaceAll(/[^A-Za-z0-9]+/g, "_")
    .toLowerCase();
const protoJSONName = (value) =>
  snake(value).replaceAll(/_([a-z0-9])/g, (_match, letter) =>
    letter.toUpperCase(),
  );
const upperSnake = (value) => {
  const normalizedValue = snake(value).toUpperCase();
  return /^[0-9]/.test(normalizedValue)
    ? `X_${normalizedValue}`
    : normalizedValue;
};

function protoFieldEmitter() {
  const nested = [];

  function emitNestedEnum(typeName, fieldName, values, indent) {
    const prefix = upperSnake(fieldName);
    nested.push(`${indent}enum ${typeName} {`);
    nested.push(`${indent}  ${prefix}_UNSPECIFIED = 0;`);
    values.forEach((value, index) => {
      const suffix =
        typeof value === "number" ? String(value) : upperSnake(value);
      nested.push(`${indent}  ${prefix}_${suffix} = ${index + 1};`);
    });
    nested.push(`${indent}}`);
  }

  function typeFor(field, fieldName, indent) {
    switch (field.type) {
      case "string":
      case "editor_color":
      case "hex_color":
      case "uri":
      case "uuid":
        return "string";
      case "boolean":
        return "bool";
      case "integer":
        return "int32";
      case "number":
        return "double";
      case "file_attachment":
        return "FileAttachment";
      case "enum":
      case "enum_int": {
        const enumName = pascal(fieldName);
        emitNestedEnum(enumName, fieldName, field.values, indent);
        return enumName;
      }
      case "object": {
        const messageName = `${pascal(fieldName)}Value`;
        nested.push(emitProtoMessage(messageName, field.fields, `${indent}`));
        return messageName;
      }
      case "array": {
        const item = field.items;
        if (item.type === "object") {
          const messageName = `${pascal(fieldName)}Item`;
          nested.push(emitProtoMessage(messageName, item.fields, `${indent}`));
          return messageName;
        }
        if (item.type === "enum" || item.type === "enum_int") {
          const enumName = `${pascal(fieldName)}Item`;
          emitNestedEnum(enumName, `${fieldName}_item`, item.values, indent);
          return enumName;
        }
        return typeFor(item, `${fieldName}Item`, indent);
      }
      default:
        throw new Error(`unsupported proto field type ${field.type}`);
    }
  }

  function emit(fields, indent) {
    const lines = [];
    let number = 1;
    for (const [fieldName, field] of Object.entries(fields)) {
      const type = typeFor(field, fieldName, indent);
      const repeated = field.type === "array" ? "repeated " : "";
      const optional =
        repeated ||
        field.required ||
        field.type === "object" ||
        field.type === "file_attachment"
          ? ""
          : "optional ";
      lines.push(
        `${indent}${repeated || optional}${type} ${snake(fieldName)} = ${number};`,
      );
      number += 1;
    }
    return lines;
  }

  return { emit, nested };
}

function emitProtoMessage(name, fields, indent = "") {
  const emitter = protoFieldEmitter();
  const fieldLines = emitter.emit(fields, `${indent}  `);
  const body = [...emitter.nested, ...fieldLines].join("\n");
  return `${indent}message ${name} {\n${body}\n${indent}}`;
}

function fieldsForOwnership(fields, ownerships) {
  return Object.fromEntries(
    Object.entries(fields).filter(([, field]) =>
      ownerships.has(field.ownership),
    ),
  );
}

const baseOwnerships = new Set(["shared", "source"]);
const localeOwnerships = new Set(["locale"]);

const richBasePropMessages = Object.entries(richBlocks)
  .map(([kind, definition]) =>
    emitProtoMessage(
      `${pascal(kind)}Props`,
      fieldsForOwnership(definition.fields, baseOwnerships),
    ),
  )
  .join("\n\n");
const richLocalePropMessages = Object.entries(richBlocks)
  .map(([kind, definition]) =>
    emitProtoMessage(
      `${pascal(kind)}LocaleProps`,
      fieldsForOwnership(definition.fields, localeOwnerships),
    ),
  )
  .join("\n\n");

const richBaseValueMessages = Object.entries(richBlocks)
  .map(([kind, definition]) => {
    const fields = [`  ${pascal(kind)}Props props = 1;`];
    if (definition.content === "table")
      fields.push("  RichTextTableBase content = 2;");
    return `message ${pascal(kind)}Block {\n${fields.join("\n")}\n}`;
  })
  .join("\n\n");
const richLocaleValueMessages = Object.entries(richBlocks)
  .map(([kind, definition]) => {
    const fields = [`  ${pascal(kind)}LocaleProps props = 1;`];
    if (definition.content === "inline")
      fields.push("  repeated RichTextInline content = 2;");
    if (definition.content === "locale_text")
      fields.push("  string content = 2;");
    if (definition.content === "table")
      fields.push("  RichTextTableLocale content = 2;");
    return `message ${pascal(kind)}BlockLocale {\n${fields.join("\n")}\n}`;
  })
  .join("\n\n");

const richOneof = Object.keys(richBlocks)
  .map(
    (kind, index) => `    ${pascal(kind)}Block ${snake(kind)} = ${index + 2};`,
  )
  .join("\n");
const richLocaleOneof = Object.keys(richBlocks)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}BlockLocale ${snake(kind)} = ${index + 2};`,
  )
  .join("\n");

const richDataOneof = Object.keys(richBlocks)
  .map(
    (kind, index) => `    ${pascal(kind)}Block ${snake(kind)} = ${index + 1};`,
  )
  .join("\n");
const richLocaleDataOneof = Object.keys(richBlocks)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}BlockLocale ${snake(kind)} = ${index + 1};`,
  )
  .join("\n");

const pageBasePropMessages = Object.entries(pageSections)
  .map(([kind, definition]) =>
    emitProtoMessage(
      `${pascal(kind)}SectionProps`,
      fieldsForOwnership(definition.fields, baseOwnerships),
    ),
  )
  .join("\n\n");
const pageLocalePropMessages = Object.entries(pageSections)
  .map(([kind, definition]) =>
    emitProtoMessage(
      `${pascal(kind)}SectionLocaleProps`,
      fieldsForOwnership(definition.fields, localeOwnerships),
    ),
  )
  .join("\n\n");

const pageBaseValueMessages = Object.entries(pageSections)
  .map(([kind, definition]) => {
    const fields = [`  ${pascal(kind)}SectionProps props = 1;`];
    if (definition.rich_text_profile)
      fields.push("  RichTextBlockGraph blocks = 2;");
    if (definition.immersive_units)
      fields.push("  repeated PageImmersiveUnit units = 2;");
    return `message ${pascal(kind)}Section {\n${fields.join("\n")}\n}`;
  })
  .join("\n\n");
const pageLocaleValueMessages = Object.entries(pageSections)
  .map(([kind, definition]) => {
    const fields = [`  ${pascal(kind)}SectionLocaleProps props = 1;`];
    if (definition.rich_text_profile)
      fields.push("  RichTextLocaleOverlay blocks = 2;");
    if (definition.immersive_units)
      fields.push("  repeated PageImmersiveUnitLocale units = 2;");
    return `message ${pascal(kind)}SectionLocale {\n${fields.join("\n")}\n}`;
  })
  .join("\n\n");

const pageOneof = Object.keys(pageSections)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}Section ${snake(kind)} = ${index + 3};`,
  )
  .join("\n");
const pageLocaleOneof = Object.keys(pageSections)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}SectionLocale ${snake(kind)} = ${index + 2};`,
  )
  .join("\n");

const pageDataOneof = Object.keys(pageSections)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}Section ${snake(kind)} = ${index + 2};`,
  )
  .join("\n");
const pageLocaleDataOneof = Object.keys(pageSections)
  .map(
    (kind, index) =>
      `    ${pascal(kind)}SectionLocale ${snake(kind)} = ${index + 1};`,
  )
  .join("\n");

const profileEnums = Object.keys(source.profiles)
  .map(
    (profile, index) =>
      `  RICH_TEXT_PROFILE_${upperSnake(profile)} = ${index + 1};`,
  )
  .join("\n");

const pageSettingsMessage = emitProtoMessage(
  "PageSectionSettings",
  source.page.settings,
);
const pageImmersiveUnitPropsMessage = emitProtoMessage(
  "PageImmersiveUnitProps",
  fieldsForOwnership(pageImmersiveUnit.fields, baseOwnerships),
);
const pageImmersiveUnitLocalePropsMessage = emitProtoMessage(
  "PageImmersiveUnitLocaleProps",
  fieldsForOwnership(pageImmersiveUnit.fields, localeOwnerships),
);

const protoOutput = `// Code generated by scripts/generated/generate-block-catalog.mjs. DO NOT EDIT.
// Canonical source: config/content/block-catalog.yaml

syntax = "proto3";

package api.content.v1;

import "api/common/v1/media.proto";

option go_package = "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1;contentv1";

enum RichTextProfile {
  RICH_TEXT_PROFILE_UNSPECIFIED = 0;
${profileEnums}
}

enum ContentValidationMode {
  CONTENT_VALIDATION_MODE_UNSPECIFIED = 0;
  CONTENT_VALIDATION_MODE_WRITE = 1;
  CONTENT_VALIDATION_MODE_RESTORE_SNAPSHOT = 2;
}

enum MissingAttachmentMediaKind {
  MISSING_ATTACHMENT_MEDIA_KIND_UNSPECIFIED = 0;
  MISSING_ATTACHMENT_MEDIA_KIND_IMAGE = 1;
  MISSING_ATTACHMENT_MEDIA_KIND_AUDIO = 2;
  MISSING_ATTACHMENT_MEDIA_KIND_VIDEO = 3;
  MISSING_ATTACHMENT_MEDIA_KIND_FILE = 4;
}

message MissingAttachment {
  string former_file_id = 1;
  MissingAttachmentMediaKind media_kind = 2;
}

message FileAttachment {
  oneof state {
    string active_file_id = 1;
    MissingAttachment missing_attachment = 2;
  }
}

message RichTextStyle {
  optional bool bold = 1;
  optional bool italic = 2;
  optional bool underline = 3;
  optional bool strike = 4;
  optional bool code = 5;
  optional string text_color = 6;
  optional string background_color = 7;
}

message RichTextStyledText {
  string text = 1;
  RichTextStyle styles = 2;
}

message RichTextHardBreak {}

message RichTextLink {
  string href = 1;
  repeated RichTextStyledText content = 2;
}

message RichTextInlineMath {
  string source = 1;
}

message RichTextSiteVariable {
  string key = 1;
  RichTextStyle styles = 2;
  bool linked = 3;
}

message RichTextInline {
  oneof value {
    RichTextStyledText text = 1;
    RichTextHardBreak hard_break = 2;
    RichTextLink link = 3;
    RichTextInlineMath math_inline = 4;
    RichTextSiteVariable site_variable = 5;
  }
}

message RichTextTableCellProps {
  optional int32 colspan = 1;
  optional int32 rowspan = 2;
  optional string background_color = 3;
  optional string text_color = 4;
  enum TextAlignment {
    TEXT_ALIGNMENT_UNSPECIFIED = 0;
    TEXT_ALIGNMENT_LEFT = 1;
    TEXT_ALIGNMENT_CENTER = 2;
    TEXT_ALIGNMENT_RIGHT = 3;
  }
  optional TextAlignment text_alignment = 5;
}

message RichTextTableCellBase {
  // Durable semantic identity. Locale content references this value and never
  // derives cell identity from array position.
  string id = 1;
  bool header = 2;
  RichTextTableCellProps props = 3;
}

message RichTextTableRowBase {
  // Durable semantic identity. Locale content references this value and never
  // derives row identity from array position.
  string id = 1;
  repeated RichTextTableCellBase cells = 2;
}

message RichTextTableBase {
  repeated double column_widths = 1;
  optional int32 header_rows = 2;
  optional int32 header_columns = 3;
  repeated RichTextTableRowBase rows = 4;
}

message RichTextTableCellLocale {
  string cell_id = 1;
  repeated RichTextInline content = 2;
}

message RichTextTableRowLocale {
  string row_id = 1;
  repeated RichTextTableCellLocale cells = 2;
}

message RichTextTableLocale { repeated RichTextTableRowLocale rows = 1; }

${richBasePropMessages}

${richLocalePropMessages}

${richBaseValueMessages}

${richLocaleValueMessages}

message RichTextBlock {
  string id = 1;
  oneof value {
${richOneof}
  }
}

message RichTextBlockLocale {
  string block_id = 1;
  oneof value {
${richLocaleOneof}
  }
}

// Closed row-local storage payloads intentionally exclude identity and placement.
// geul-contract-root: generated-runtime
message RichTextBlockData {
  oneof value {
${richDataOneof}
  }
}

// geul-contract-root: generated-runtime
message RichTextBlockLocaleData {
  oneof value {
${richLocaleDataOneof}
  }
}

message ContentBlockPlacement {
  optional string parent_block_id = 1;
  uint32 index = 2;
}

message RichTextBlockNode {
  RichTextBlock block = 1;
  ContentBlockPlacement placement = 2;
}

message RichTextBlockGraph { repeated RichTextBlockNode nodes = 1; }

message RichTextLocaleOverlay {
  string locale = 1;
  repeated RichTextBlockLocale blocks = 2;
}

message RichTextDocument {
  string block_catalog_fingerprint = 1;
  RichTextProfile profile = 2;
  string source_locale = 3;
  RichTextBlockGraph base = 4;
  repeated RichTextLocaleOverlay locale_overlays = 5;
}

message LocalizedRichTextDocument {
  string block_catalog_fingerprint = 1;
  RichTextProfile profile = 2;
  string locale = 3;
  RichTextBlockGraph base = 4;
  RichTextLocaleOverlay locale_overlay = 5;
}

message UpsertRichTextBlock {
  RichTextBlockNode node = 1;
}

message DeleteRichTextBlock { string block_id = 1; }

message MoveRichTextBlock {
  string block_id = 1;
  ContentBlockPlacement placement = 2;
}

message RichTextBlockMutation {
  oneof operation {
    UpsertRichTextBlock upsert = 1;
    DeleteRichTextBlock delete = 2;
    MoveRichTextBlock move = 3;
  }
}

message UpsertRichTextBlockLocale { RichTextBlockLocale block = 1; }
message DeleteRichTextBlockLocale { string block_id = 1; }

message RichTextBlockLocaleMutation {
  oneof operation {
    UpsertRichTextBlockLocale upsert = 1;
    DeleteRichTextBlockLocale delete = 2;
  }
}

message RichTextLocaleMutationGroup {
  string locale = 1;
  repeated RichTextBlockLocaleMutation mutations = 2;
}

message RichTextBlockMutationBatch {
  string block_catalog_fingerprint = 1;
  RichTextProfile profile = 2;
  string expected_revision = 3;
  repeated RichTextBlockMutation base_mutations = 4;
  repeated RichTextLocaleMutationGroup locale_mutation_groups = 5;
  repeated string contributor_member_ids = 6;
}

${pageSettingsMessage}

${pageImmersiveUnitPropsMessage}

${pageImmersiveUnitLocalePropsMessage}

message PageImmersiveUnit {
  string id = 1;
  PageImmersiveUnitProps props = 2;
}

message PageImmersiveUnitLocale {
  string unit_id = 1;
  PageImmersiveUnitLocaleProps props = 2;
}

${pageBasePropMessages}

${pageLocalePropMessages}

${pageBaseValueMessages}

${pageLocaleValueMessages}

message PageSection {
  string id = 1;
  PageSectionSettings settings = 2;
  oneof value {
${pageOneof}
  }
}

message PageSectionLocale {
  string section_id = 1;
  oneof value {
${pageLocaleOneof}
  }
}

// Closed row-local storage payloads intentionally exclude identity and placement.
// geul-contract-root: generated-runtime
message PageSectionData {
  PageSectionSettings settings = 1;
  oneof value {
${pageDataOneof}
  }
}

// geul-contract-root: generated-runtime
message PageSectionLocaleData {
  oneof value {
${pageLocaleDataOneof}
  }
}

message PageSectionPlacement {
  optional string parent_section_id = 1;
  optional string column_id = 2;
  uint32 index = 3;
}

message PageSectionNode {
  PageSection section = 1;
  PageSectionPlacement placement = 2;
}

message PageSectionGraph { repeated PageSectionNode nodes = 1; }

message PageLocaleOverlay {
  string locale = 1;
  repeated PageSectionLocale sections = 2;
}

message PageDocument {
  string block_catalog_fingerprint = 1;
  string source_locale = 2;
  PageSectionGraph base = 3;
  repeated PageLocaleOverlay locale_overlays = 4;
}

message LocalizedPageDocument {
  string block_catalog_fingerprint = 1;
  string locale = 2;
  PageSectionGraph base = 3;
  PageLocaleOverlay locale_overlay = 4;
}

message UpsertPageSection {
  PageSectionNode node = 1;
}

message DeletePageSection { string section_id = 1; }

message MovePageSection {
  string section_id = 1;
  PageSectionPlacement placement = 2;
}

message MutatePageRichTextBlock {
  string section_id = 1;
  RichTextBlockMutation mutation = 2;
}

message PageSectionMutation {
  oneof operation {
    UpsertPageSection upsert = 1;
    DeletePageSection delete = 2;
    MovePageSection move = 3;
    MutatePageRichTextBlock mutate_rich_text_block = 4;
  }
}

message UpsertPageSectionLocale { PageSectionLocale section = 1; }
message DeletePageSectionLocale { string section_id = 1; }
message MutatePageRichTextBlockLocale {
  string section_id = 1;
  RichTextBlockLocaleMutation mutation = 2;
}

message PageSectionLocaleMutation {
  oneof operation {
    UpsertPageSectionLocale upsert = 1;
    DeletePageSectionLocale delete = 2;
    MutatePageRichTextBlockLocale mutate_rich_text_block = 3;
  }
}

message PageLocaleMutationGroup {
  string locale = 1;
  repeated PageSectionLocaleMutation mutations = 2;
}

message PageSectionMutationBatch {
  string block_catalog_fingerprint = 1;
  string expected_revision = 2;
  repeated PageSectionMutation base_mutations = 3;
  repeated PageLocaleMutationGroup locale_mutation_groups = 4;
  repeated string contributor_member_ids = 5;
}

message ContentBlockMediaSelector {
  string block_id = 1;
  string reference_path = 2;
}

// geul-contract-root: generated-runtime
message ContentBlockMediaReference {
  string block_id = 1;
  string reference_path = 2;
  string active_file_id = 3;
}

enum ContentBlockDownloadAvailability {
  CONTENT_BLOCK_DOWNLOAD_AVAILABILITY_UNSPECIFIED = 0;
  CONTENT_BLOCK_DOWNLOAD_AVAILABILITY_AVAILABLE = 1;
  CONTENT_BLOCK_DOWNLOAD_AVAILABILITY_UNAVAILABLE = 2;
}

enum ContentBlockDownloadAction {
  CONTENT_BLOCK_DOWNLOAD_ACTION_UNSPECIFIED = 0;
  CONTENT_BLOCK_DOWNLOAD_ACTION_DOWNLOAD = 1;
  CONTENT_BLOCK_DOWNLOAD_ACTION_SIGN_IN = 2;
  CONTENT_BLOCK_DOWNLOAD_ACTION_NONE = 3;
}

message ContentBlockMediaItem {
  ContentBlockMediaSelector selector = 1;
  FileAttachment attachment = 2;
  optional api.common.v1.MediaDelivery delivery = 3;
  ContentBlockDownloadAvailability download_availability = 4;
  ContentBlockDownloadAction download_action = 5;
}
`;

function writeOrCheck(file, content, label) {
  if (check) {
    const current = fs.existsSync(file) ? fs.readFileSync(file, "utf8") : "";
    if (current !== content)
      throw new Error(
        `generated ${label} is stale: ${path.relative(root, file)}`,
      );
    return;
  }
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, content);
}

writeOrCheck(protoPath, protoOutput, "Block catalog proto");

if (!protoOnly) {
  const runtime = generateRuntimeCatalog(normalized, fingerprint);
  const formattedGo = execFileSync("gofmt", [], {
    input: runtime.go,
    encoding: "utf8",
  });
  const formattedTs = execFileSync(
    "pnpm",
    ["exec", "prettier", "--stdin-filepath", tsPath],
    { input: runtime.ts, encoding: "utf8", cwd: root },
  );
  writeOrCheck(goPath, formattedGo, "Go Block catalog");
  writeOrCheck(tsPath, formattedTs, "TypeScript Block catalog");
}

function generateRuntimeCatalog(catalog, catalogFingerprint) {
  const descriptor = {
    limits: catalog.limits,
    runtimeForbiddenFields: catalog.runtime_forbidden_fields,
    fileReferencePolicies: catalog.file_reference_policies,
    richTextBlocks: Object.fromEntries(
      Object.entries(catalog.rich_text.blocks).map(([kind, definition]) => [
        kind,
        { fields: definition.fields, content: definition.content ?? "none" },
      ]),
    ),
    richTextInline: catalog.rich_text.inline,
    richTextTable: catalog.rich_text.table,
    profiles: catalog.profiles,
    pageSettings: catalog.page.settings,
    pageImmersiveUnit: catalog.page.immersive_unit.fields,
    pageSections: Object.fromEntries(
      Object.entries(catalog.page.sections).map(([kind, definition]) => [
        kind,
        {
          fields: definition.fields,
          richTextProfile: definition.rich_text_profile ?? null,
          columns: definition.columns === true,
        },
      ]),
    ),
    pageColumnChildKinds: catalog.page.column_child_kinds,
  };
  return generateRuntimeFiles(descriptor, catalogFingerprint);
}

function generateRuntimeFiles(descriptor, catalogFingerprint) {
  // Kept in this generator so the YAML remains the only handwritten schema.
  // The implementation body is appended below to keep proto generation usable
  // before buf has materialized language bindings.
  return {
    go: generateGoRuntimeLegacy(descriptor, catalogFingerprint),
    ts: generateTsRuntime(descriptor, catalogFingerprint),
  };
}

function collaborativeFieldPaths(fields, prefix = "props.") {
  const result = [];
  for (const [name, field] of Object.entries(fields)) {
    const path = `${prefix}${name}`;
    if (field.type === "string") result.push(path);
    if (field.type === "object")
      result.push(...collaborativeFieldPaths(field.fields, `${path}.`));
    if (field.type === "array" && field.items?.type === "object") {
      result.push(
        ...collaborativeFieldPaths(field.items.fields, `${path}[*].`),
      );
    }
  }
  return result;
}

function generateGoRuntimeLegacy(descriptor, catalogFingerprint) {
  // The .tmpl is handwritten generator source. Only the rendered Go file gets
  // the generated-file banner and is forbidden from direct editing.
  const template = fs.readFileSync(
    path.join(root, "scripts/generated/runtime/block_catalog_runtime.go.tmpl"),
    "utf8",
  );
  return template
    .replace(
      "__GENERATED_FILE_HEADER__",
      "// Code generated by scripts/generated/generate-block-catalog.mjs. DO NOT EDIT.",
    )
    .replaceAll("__CONTENT_BLOCK_CATALOG_FINGERPRINT__", catalogFingerprint)
    .replace("__CONTENT_BLOCK_RUNTIME_DESCRIPTOR__", JSON.stringify(descriptor))
    .replace(
      "__SITE_VARIABLE_KEYS__",
      descriptor.richTextInline.kinds.siteVariable.fields.key.values
        .map((key) => JSON.stringify(key))
        .join(", "),
    )
    .replace(
      "__SITE_VARIABLE_LINKED_KEYS__",
      descriptor.richTextInline.kinds.siteVariable.linked_keys
        .map((key) => JSON.stringify(key))
        .join(", "),
    );
}

function generateTsRuntime(descriptor, catalogFingerprint) {
  const profileByNumber = Object.fromEntries(
    Object.keys(descriptor.profiles).map((profile, index) => [
      index + 1,
      profile,
    ]),
  );
  return `// Code generated by scripts/generated/generate-block-catalog.mjs. DO NOT EDIT.

import { clone, fromJson, toBinary, toJson } from "@bufbuild/protobuf";
import type { JsonValue } from "@bufbuild/protobuf";
import type {
  ContentBlockMediaReference, FileAttachment, LocalizedPageDocument,
  LocalizedRichTextDocument, PageDocument, PageLocaleOverlay, PageSection,
  PageSectionLocale, PageSectionMutationBatch, RichTextBlock,
  RichTextBlockData, RichTextBlockLocale, RichTextBlockLocaleData,
  RichTextBlockMutationBatch, RichTextDocument,
  RichTextInline, RichTextLocaleOverlay,
  PageSectionData, PageSectionLocaleData,
} from "./block_content_pb.ts";
import {
  ContentValidationMode, LocalizedPageDocumentSchema,
  LocalizedRichTextDocumentSchema, PageDocumentSchema, RichTextDocumentSchema,
  RichTextBlockDataSchema, RichTextBlockLocaleDataSchema,
  PageSectionDataSchema, PageSectionLocaleDataSchema,
  MissingAttachmentMediaKind, RichTextProfile,
} from "./block_content_pb.ts";

export const contentBlockCatalogFingerprint = ${JSON.stringify(catalogFingerprint)} as const;
export const contentBlockRuntimeForbiddenFields = ${JSON.stringify(descriptor.runtimeForbiddenFields, null, 2)} as const;
export const richTextBlockKinds = ${JSON.stringify(Object.keys(descriptor.richTextBlocks), null, 2)} as const;
export type RichTextBlockKind = (typeof richTextBlockKinds)[number];
export const pageSectionKinds = ${JSON.stringify(Object.keys(descriptor.pageSections), null, 2)} as const;
export type PageSectionKind = (typeof pageSectionKinds)[number];
export const siteVariableKeys = ${JSON.stringify(descriptor.richTextInline.kinds.siteVariable.fields.key.values)} as const;
export type SiteVariableKey = (typeof siteVariableKeys)[number];
export function isSiteVariableKey(key: string): key is SiteVariableKey { return (siteVariableKeys as readonly string[]).includes(key); }
const linkedSiteVariableKeys: readonly string[] = ${JSON.stringify(descriptor.richTextInline.kinds.siteVariable.linked_keys)};
export const richTextProfiles = ${JSON.stringify(descriptor.profiles, null, 2)} as const;
export const richTextBlockCatalog = ${JSON.stringify(descriptor.richTextBlocks, null, 2)} as const;
const richTextInlineStyleCatalog = ${JSON.stringify(descriptor.richTextInline.styles, null, 2)} as const;
const richTextLinkFieldCatalog = ${JSON.stringify(descriptor.richTextInline.kinds.link.fields, null, 2)} as const;
export const pageSectionCatalog = ${JSON.stringify(descriptor.pageSections, null, 2)} as const;
export const pageSectionSettingsCatalog = ${JSON.stringify(descriptor.pageSettings, null, 2)} as const;
export const pageImmersiveUnitCatalog = ${JSON.stringify(descriptor.pageImmersiveUnit, null, 2)} as const;
export const pageColumnChildKinds = ${JSON.stringify(descriptor.pageColumnChildKinds, null, 2)} as const;
export const contentBlockFileReferencePolicies = ${JSON.stringify(descriptor.fileReferencePolicies, null, 2)} as const;
export const richTextBlockKindByProtoCase = ${JSON.stringify(Object.fromEntries(Object.keys(descriptor.richTextBlocks).map((kind) => [protoJSONName(kind), kind])), null, 2)} as const;
export const pageSectionKindByProtoCase = ${JSON.stringify(Object.fromEntries(Object.keys(descriptor.pageSections).map((kind) => [protoJSONName(kind), kind])), null, 2)} as const;
export const richTextCollaborativeTextCatalog = ${JSON.stringify(Object.fromEntries(Object.entries(descriptor.richTextBlocks).map(([kind, rule]) => [kind, [...collaborativeFieldPaths(rule.fields), ...(rule.content === "inline" || rule.content === "locale_text" || rule.content === "table" ? ["content"] : [])]])), null, 2)} as const;
export const pageSectionCollaborativeTextCatalog = ${JSON.stringify(Object.fromEntries(Object.entries(descriptor.pageSections).map(([kind, rule]) => [kind, [...collaborativeFieldPaths(rule.fields), ...(kind === "immersive-scene" ? collaborativeFieldPaths(descriptor.pageImmersiveUnit, "units[*].props.") : [])]])), null, 2)} as const;
const profileByNumber = ${JSON.stringify(profileByNumber, null, 2)} as const;

type Ownership = "shared" | "locale" | "source";
type FieldSpec = Readonly<Record<string, unknown>> & { readonly type: string; readonly ownership?: Ownership };
type FieldCatalog = Readonly<Record<string, FieldSpec>>;
export function richTextBlockFieldOwnership(kind: RichTextBlockKind, field: string): Ownership | undefined { return (richTextBlockCatalog[kind].fields as FieldCatalog)[field]?.ownership; }
export function pageSectionFieldOwnership(kind: PageSectionKind, field: string): Ownership | undefined { return (pageSectionCatalog[kind].fields as FieldCatalog)[field]?.ownership; }
export function richTextCollaborativeTextFields(kind: RichTextBlockKind): readonly string[] { return richTextCollaborativeTextCatalog[kind]; }
export function pageSectionCollaborativeTextFields(kind: PageSectionKind): readonly string[] { return pageSectionCollaborativeTextCatalog[kind]; }
function collaborativePatternMatches(pattern: string, path: string): boolean {
  const expected = pattern.split("."); const actual = path.split(".");
  if (expected.length !== actual.length) return false;
  return expected.every((segment, index) => {
    if (!segment.endsWith("[*]")) return segment === actual[index];
    const name = segment.slice(0, -3); const value = actual[index];
    if (!value.startsWith(name + "[") || !value.endsWith("]")) return false;
    return /^\\d+$/.test(value.slice(name.length + 1, -1));
  });
}
export function isRichTextCollaborativeTextPath(kind: RichTextBlockKind, path: string): boolean {
  if (richTextCollaborativeTextCatalog[kind].filter((field) => field !== "content").some((field) => collaborativePatternMatches(field, path))) return true;
  const content = richTextBlockCatalog[kind].content;
  if (content === "locale_text") return path === "content";
  const inlineLeaf = String.raw\`(?:text\\.text|link\\.content\\[\\d+\\]\\.text|mathInline\\.source)\`;
  if (content === "inline") return new RegExp(String.raw\`^content\\[\\d+\\]\\.${"${inlineLeaf}"}$\`).test(path);
  if (content === "table") return new RegExp(String.raw\`^content\\.rows\\[\\d+\\]\\.cells\\[\\d+\\]\\.content\\[\\d+\\]\\.${"${inlineLeaf}"}$\`).test(path);
  return false;
}
export function isPageSectionCollaborativeTextPath(kind: PageSectionKind, path: string): boolean { return pageSectionCollaborativeTextCatalog[kind].some((field) => collaborativePatternMatches(field, path)); }

export interface ContentBlockValidationIssue { readonly path: string; readonly code: string; readonly message: string }
export class ContentBlockValidationError extends Error {
  readonly issue: ContentBlockValidationIssue;
  constructor(issue: ContentBlockValidationIssue) { super(\`${"${issue.path}"}: ${"${issue.message}"} (${"${issue.code}"})\`); this.name = "ContentBlockValidationError"; this.issue = issue; }
}

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const hexColorPattern = /^#(?:[0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$/i;
const editorColorTokens = new Set(["default", "gray", "brown", "red", "orange", "yellow", "green", "blue", "purple", "pink"]);
const fail = (path: string, code: string, message: string): never => { throw new ContentBlockValidationError({ path, code, message }); };
const required = <T>(value: T | null | undefined, path: string, message: string): T => value ?? fail(path, "required", message);
const requireUuid = (value: string, path: string): void => { if (!uuidPattern.test(value)) fail(path, "uuid", "value must be a UUID"); };
const requireUnique = (values: readonly string[], path: string): void => { if (new Set(values).size !== values.length) fail(path, "duplicate", "values must be unique"); };
const validMissingAttachmentMediaKinds = new Set<MissingAttachmentMediaKind>([MissingAttachmentMediaKind.IMAGE, MissingAttachmentMediaKind.AUDIO, MissingAttachmentMediaKind.VIDEO, MissingAttachmentMediaKind.FILE]);

function validateAttachment(value: unknown, path: string, mode: ContentValidationMode): void {
  const attachment = required(value as FileAttachment | undefined, path, "file attachment is required");
  if (attachment.state.case === undefined) fail(path, "required", "file attachment state is required");
  if (attachment.state.case === "activeFileId") { requireUuid(attachment.state.value, path + ".activeFileId"); attachment.state = { case: "activeFileId", value: attachment.state.value.toLowerCase() }; }
  else { const missing = required(attachment.state.value, path, "missing attachment is required"); requireUuid(missing.formerFileId, path + ".missingAttachment.formerFileId"); missing.formerFileId = missing.formerFileId.toLowerCase(); if (!validMissingAttachmentMediaKinds.has(missing.mediaKind)) fail(path + ".missingAttachment.mediaKind", "required", "missing attachment media kind is required"); if (mode !== ContentValidationMode.RESTORE_SNAPSHOT) fail(path, "restore_only", "missing attachment is snapshot-restore only"); }
}

function validateFields(value: unknown, fields: FieldCatalog, ownerships: ReadonlySet<Ownership>, path: string, mode: ContentValidationMode): void {
  const record = (value ?? {}) as Record<string, unknown>;
  for (const [name, spec] of Object.entries(fields)) {
    if (!ownerships.has(spec.ownership ?? "shared")) continue;
    const fieldValue = record[name]; const required = spec.required === true;
    if (fieldValue === undefined || fieldValue === null) { if (required) fail(path + "." + name, "required", "field is required"); continue; }
    if (spec.type === "array" && Array.isArray(fieldValue) && fieldValue.length === 0 && !required && !Object.hasOwn(spec, "default")) continue;
    if (spec.type === "uuid") requireUuid(String(fieldValue), path + "." + name);
    else if (spec.type === "uri") { try { const uri = String(fieldValue); if (spec.root_relative === true && uri.startsWith("/") && !uri.startsWith("//") && !uri.includes("\\\\") && !/[\\u0000-\\u0020\\u007f]/.test(uri)) { new URL(uri, "https://example.invalid"); continue; } const parsed = new URL(uri); const allowed = (spec.schemes as readonly string[] | undefined) ?? []; if (!allowed.includes(parsed.protocol.slice(0, -1))) fail(path + "." + name, "uri_scheme", "URI scheme is forbidden"); } catch { fail(path + "." + name, "uri", "invalid URI"); } }
    else if (spec.type === "editor_color") { const color = String(fieldValue); if (!editorColorTokens.has(color) && !hexColorPattern.test(color)) fail(path + "." + name, "color", "invalid editor color"); }
    else if (spec.type === "hex_color") { const color = String(fieldValue); if (!(spec.allow_empty === true && color === "") && !hexColorPattern.test(color)) fail(path + "." + name, "color", "invalid hex color"); }
    else if (spec.type === "file_attachment") validateAttachment(fieldValue, path + "." + name, mode);
    else if (spec.type === "enum" || spec.type === "enum_int") { const count = (spec.values as readonly unknown[]).length; if (!Number.isInteger(fieldValue) || Number(fieldValue) < 1 || Number(fieldValue) > count) fail(path + "." + name, "enum", "invalid enum value"); }
    else if (spec.type === "integer" || spec.type === "number") { const numeric = Number(fieldValue); if (!Number.isFinite(numeric) || (spec.type === "integer" && !Number.isSafeInteger(numeric))) fail(path + "." + name, "number", "invalid numeric value"); if (typeof spec.min === "number" && numeric < spec.min) fail(path + "." + name, "min", "value is below minimum"); if (typeof spec.max === "number" && numeric > spec.max) fail(path + "." + name, "max", "value exceeds maximum"); if (typeof spec.min_exclusive === "number" && numeric <= spec.min_exclusive) fail(path + "." + name, "min_exclusive", "value must be greater than minimum"); }
    else if (spec.type === "string") { if (typeof fieldValue !== "string") fail(path + "." + name, "type", "expected string"); const text = fieldValue as string; if (typeof spec.max_length === "number" && text.length > spec.max_length) fail(path + "." + name, "max_length", "string is too long"); }
    else if (spec.type === "boolean" && typeof fieldValue !== "boolean") fail(path + "." + name, "type", "expected boolean");
    else if (spec.type === "array") { if (!Array.isArray(fieldValue)) fail(path + "." + name, "type", "expected array"); const values = fieldValue as unknown[]; if (typeof spec.exact_length === "number" && values.length !== spec.exact_length) fail(path + "." + name, "length", "array length mismatch"); if (typeof spec.min_length === "number" && values.length < spec.min_length) fail(path + "." + name, "length", "array is too short"); if (typeof spec.max_length === "number" && values.length > spec.max_length) fail(path + "." + name, "length", "array is too long"); if (spec.unique === true && new Set(values.map((item) => JSON.stringify(item))).size !== values.length) fail(path + "." + name, "duplicate", "array values must be unique"); const itemSpec = spec.items as FieldSpec; values.forEach((item, index) => validateFields({ item }, { item: { ...itemSpec, ownership: spec.ownership } }, ownerships, \`${"${path}"}.${"${name}"}[${"${index}"}]\`, mode)); }
    else if (spec.type === "object") validateFields(fieldValue, spec.fields as FieldCatalog, ownerships, path + "." + name, mode);
  }
}

// Link presentation media follows the profile's paragraph video permission; restricted profiles stay media-free.
function validateProfileLinkMedia(kind: string, props: unknown, profile: keyof typeof richTextProfiles, path: string): void {
  if (kind !== "paragraph" || richTextProfiles[profile].paragraph_external_video) return;
  for (const field of ["linkThumbnail", "linkIcon"] as const) {
    const value = (props as Partial<Record<typeof field, unknown>> | undefined)?.[field];
    if (value !== undefined && value !== null) fail(path + "." + field, "profile", "link media is forbidden by profile");
  }
}
function validateInlineStyles(value: unknown, profile: keyof typeof richTextProfiles, path: string, mode: ContentValidationMode): void {
  const styles = mutableRecord(value);
  for (const [mark, active] of Object.entries(styles)) {
    if (mark !== "$typeName" && active !== undefined && active !== false && !richTextProfiles[profile].marks.includes(mark as never)) fail(path + "." + mark, "profile", "mark is forbidden");
  }
  validateFields(styles, richTextInlineStyleCatalog as FieldCatalog, new Set(["locale"]), path, mode);
}
function validateInline(nodes: readonly RichTextInline[], profile: keyof typeof richTextProfiles, path: string, mode: ContentValidationMode): void {
  nodes.forEach((node, index) => {
    const itemPath = path + "[" + index + "]";
    if (node.value.case === undefined) fail(itemPath, "required", "inline value is required");
    if (node.value.case === "mathInline" && !richTextProfiles[profile].inline_math) fail(itemPath + ".mathInline", "profile", "inline math is forbidden");
    if (node.value.case === "text" && node.value.value.styles) validateInlineStyles(node.value.value.styles, profile, itemPath + ".text.styles", mode);
    if (node.value.case === "siteVariable") {
      const variable = node.value.value;
      if (!isSiteVariableKey(variable.key)) fail(itemPath + ".siteVariable.key", "site_variable_key", "unknown site variable key");
      if (variable.linked && !linkedSiteVariableKeys.includes(variable.key)) fail(itemPath + ".siteVariable.linked", "site_variable_linked", "site variable cannot be linked");
      if (variable.styles) validateInlineStyles(variable.styles, profile, itemPath + ".siteVariable.styles", mode);
    }
    if (node.value.case === "link") {
      validateFields(node.value.value, richTextLinkFieldCatalog as FieldCatalog, new Set(["locale"]), itemPath + ".link", mode);
      node.value.value.content.forEach((text, textIndex) => {
        if (text.styles) validateInlineStyles(text.styles, profile, itemPath + ".link.content[" + textIndex + "].styles", mode);
      });
    }
  });
}
function validateLocalizedRichContent(value: unknown, kind: RichTextBlockKind, profile: keyof typeof richTextProfiles, path: string, mode: ContentValidationMode): void {
  const content = mutableRecord(value).content;
  if (richTextBlockCatalog[kind].content === "inline") {
    validateInline((content ?? []) as readonly RichTextInline[], profile, path + ".content", mode);
    return;
  }
  if (richTextBlockCatalog[kind].content !== "table") return;
  const rows = (mutableRecord(content).rows ?? []) as readonly unknown[];
  rows.forEach((row, rowIndex) => {
    const cells = (mutableRecord(row).cells ?? []) as readonly unknown[];
    cells.forEach((cell, cellIndex) => validateInline(
      (mutableRecord(cell).content ?? []) as readonly RichTextInline[],
      profile,
      path + ".content.rows[" + rowIndex + "].cells[" + cellIndex + "].content",
      mode,
    ));
  });
}
function normalizeRichInlineJSON(value: unknown): void {
  if (Array.isArray(value)) {
    value.forEach(normalizeRichInlineJSON);
    return;
  }
  if (value === null || typeof value !== "object") return;
  const record = mutableRecord(value);
  if (record.styles !== undefined) normalizeFieldRecord(mutableRecord(record.styles), richTextInlineStyleCatalog as FieldCatalog, new Set(["locale"]), false);
  Object.values(record).forEach(normalizeRichInlineJSON);
}

function profileName(profile: RichTextProfile): keyof typeof richTextProfiles { const name = profileByNumber[String(profile) as keyof typeof profileByNumber]; if (!name) fail("$.profile", "profile", "unknown profile"); return name; }

function validateRichGraph(documentProfile: RichTextProfile, base: NonNullable<RichTextDocument["base"]>, overlays: readonly RichTextLocaleOverlay[], sourceLocale: string, mode: ContentValidationMode, path: string): void {
  const profile = profileName(documentProfile); const nodes = base.nodes; if (nodes.length > ${descriptor.limits.max_document_blocks}) fail(path, "limit", "too many blocks");
  const byId = new Map<string, RichTextBlock>(); const slots = new Set<string>();
  nodes.forEach((node, index) => { const nodePath = \`${"${path}"}.nodes[${"${index}"}]\`; const block = required(node.block, nodePath + ".block", "block is required"); const placement = required(node.placement, nodePath + ".placement", "placement is required"); requireUuid(block.id, nodePath + ".block.id"); if (byId.has(block.id)) fail(nodePath, "duplicate_id", "duplicate block id"); byId.set(block.id, block); if (block.value.case === undefined) fail(nodePath, "required", "block value is required"); const kind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!kind) fail(nodePath, "kind", "unknown rich-text Block kind"); const blockValue = required(block.value.value, nodePath, "block value is required"); if (!richTextProfiles[profile].blocks.includes(kind as never)) fail(nodePath, "profile", "block kind is forbidden"); validateFields(blockValue.props, richTextBlockCatalog[kind].fields as FieldCatalog, new Set(["shared", "source"]), nodePath + ".props", mode); validateProfileLinkMedia(kind, blockValue.props, profile, nodePath + ".props"); if (placement.parentBlockId !== undefined) requireUuid(placement.parentBlockId, nodePath + ".placement.parentBlockId"); const slot = (placement.parentBlockId ?? "root") + ":" + placement.index; if (slots.has(slot)) fail(nodePath + ".placement", "duplicate_position", "duplicate sibling position"); slots.add(slot); });
  for (const node of nodes) if (node.placement?.parentBlockId && !byId.has(node.placement.parentBlockId)) fail(path, "orphan", "parent block does not exist");
  const locales = overlays.map((overlay) => overlay.locale); requireUnique(locales, path + ".localeOverlays"); if (!locales.includes(sourceLocale)) fail(path + ".sourceLocale", "source_overlay", "source locale overlay is required");
  overlays.forEach((overlay, overlayIndex) => { if (!overlay.locale.trim()) fail(path, "locale", "locale is required"); const seen = new Set<string>(); overlay.blocks.forEach((localized, index) => { const itemPath = \`${"${path}"}.localeOverlays[${"${overlayIndex}"}].blocks[${"${index}"}]\`; requireUuid(localized.blockId, itemPath + ".blockId"); if (seen.has(localized.blockId)) fail(itemPath, "duplicate_id", "duplicate locale block"); seen.add(localized.blockId); const baseBlock = required(byId.get(localized.blockId), itemPath, "locale block has no base block"); if (localized.value.case === undefined || localized.value.case !== baseBlock.value.case) fail(itemPath, "kind_mismatch", "locale kind must match base kind"); const kind = richTextBlockKindByProtoCase[localized.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!kind) fail(itemPath, "kind", "unknown rich-text locale Block kind"); const localizedValue = required(localized.value.value, itemPath, "locale value is required"); validateFields(localizedValue.props, richTextBlockCatalog[kind].fields as FieldCatalog, new Set(["locale"]), itemPath + ".props", mode); validateLocalizedRichContent(localizedValue, kind, profile, itemPath, mode); }); });
}

function mutableRecord(value: unknown): Record<string, unknown> { return (value ?? {}) as Record<string, unknown>; }
function runtimeUpperSnake(value: unknown): string { const normalized = String(value).replace(/([a-z0-9])([A-Z])/g, "$1_$2").replace(/[^A-Za-z0-9]+/g, "_").toUpperCase(); return /^[0-9]/.test(normalized) ? "X_" + normalized : normalized; }
function runtimeEnumSuffix(type: string, value: unknown): string { return type === "enum_int" && typeof value === "number" ? String(value) : runtimeUpperSnake(value); }
function enumDefault(fieldName: string, value: unknown, item: boolean, type: string): string { return runtimeUpperSnake(fieldName) + (item ? "_ITEM_" : "_") + runtimeEnumSuffix(type, value); }
function normalizedDefault(fieldName: string, spec: FieldSpec): unknown { if (spec.type === "enum" || spec.type === "enum_int") return enumDefault(fieldName, spec.default, false, spec.type); if (spec.type === "array" && Array.isArray(spec.default)) { const item = spec.items as FieldSpec; return spec.default.map((value) => item.type === "enum" || item.type === "enum_int" ? enumDefault(fieldName, value, true, item.type) : value); } return structuredClone(spec.default); }
function normalizeFieldRecord(value: Record<string, unknown>, fields: FieldCatalog, ownerships: ReadonlySet<Ownership>, materializeDefaults: boolean): void {
  for (const [name, spec] of Object.entries(fields)) { if (!ownerships.has(spec.ownership ?? "shared")) continue; if (materializeDefaults && value[name] === undefined && Object.hasOwn(spec, "default")) value[name] = normalizedDefault(name, spec); const current = value[name]; if (typeof current === "number" && Object.is(current, -0)) value[name] = 0; if (spec.type === "array" && Array.isArray(current)) { const item = spec.items as FieldSpec; if (item.type === "object") current.forEach((entry) => normalizeFieldRecord(mutableRecord(entry), item.fields as FieldCatalog, ownerships, materializeDefaults)); if (spec.unique === true) current.sort((left, right) => JSON.stringify(left).localeCompare(JSON.stringify(right))); } if (spec.type === "object" && current !== undefined) normalizeFieldRecord(mutableRecord(current), spec.fields as FieldCatalog, ownerships, materializeDefaults); if ((spec.type === "editor_color" || spec.type === "hex_color") && typeof current === "string" && current.startsWith("#")) value[name] = current.toLowerCase(); }
}
function findVariant(record: Record<string, unknown>, cases: readonly string[]): [string, Record<string, unknown>] { for (const name of cases) { const value = record[name]; if (value !== undefined) return [name, mutableRecord(value)]; } return fail("$.value", "required", "typed value is required"); }
function normalizeRichJSON(base: Record<string, unknown>, overlays: unknown[]): void { const protoCases = Object.keys(richTextBlockKindByProtoCase); const nodes = (base.nodes ?? []) as unknown[]; for (const node of nodes) { const block = mutableRecord(mutableRecord(node).block); const [protoCase, payload] = findVariant(block, protoCases); const kind = richTextBlockKindByProtoCase[protoCase as keyof typeof richTextBlockKindByProtoCase]; const props = mutableRecord(payload.props); payload.props = props; normalizeFieldRecord(props, richTextBlockCatalog[kind].fields as FieldCatalog, new Set(["shared", "source"]), true); } for (const overlay of overlays) { for (const localized of (mutableRecord(overlay).blocks ?? []) as unknown[]) { const block = mutableRecord(localized); const [protoCase, payload] = findVariant(block, protoCases); const kind = richTextBlockKindByProtoCase[protoCase as keyof typeof richTextBlockKindByProtoCase]; const props = mutableRecord(payload.props); payload.props = props; normalizeFieldRecord(props, richTextBlockCatalog[kind].fields as FieldCatalog, new Set(["locale"]), false); normalizeRichInlineJSON(payload.content); } } }
function normalizePageJSON(base: Record<string, unknown>, overlays: unknown[]): void { const localizedByLocale = new Map<string, Map<string, Record<string, unknown>>>(); for (const overlay of overlays) { const record = mutableRecord(overlay); localizedByLocale.set(String(record.locale), new Map(((record.sections ?? []) as unknown[]).map((section) => { const value = mutableRecord(section); return [String(value.sectionId), value]; }))); } for (const node of (base.nodes ?? []) as unknown[]) { const section = mutableRecord(mutableRecord(node).section); const settings = mutableRecord(section.settings); section.settings = settings; normalizeFieldRecord(settings, pageSectionSettingsCatalog as FieldCatalog, new Set(["shared"]), true); const [protoCase, payload] = findVariant(section, Object.keys(pageSectionKindByProtoCase)); const kind = pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase]; const props = mutableRecord(payload.props); payload.props = props; normalizeFieldRecord(props, pageSectionCatalog[kind].fields as FieldCatalog, new Set(["shared", "source"]), true); if (protoCase === "richText") { const nestedOverlays: unknown[] = []; for (const [locale, sections] of localizedByLocale) { const localized = sections.get(String(section.id)); const rich = localized === undefined ? undefined : mutableRecord(localized).richText; nestedOverlays.push(rich === undefined ? { locale, blocks: [] } : mutableRecord(rich).blocks); } normalizeRichJSON(mutableRecord(payload.blocks), nestedOverlays); } } for (const sections of localizedByLocale.values()) for (const localized of sections.values()) { const [protoCase, payload] = findVariant(localized, Object.keys(pageSectionKindByProtoCase)); const kind = pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase]; const props = mutableRecord(payload.props); payload.props = props; normalizeFieldRecord(props, pageSectionCatalog[kind].fields as FieldCatalog, new Set(["locale"]), false); if (protoCase === "immersiveScene") for (const unit of (payload.units ?? []) as unknown[]) normalizeFieldRecord(mutableRecord(mutableRecord(unit).props), pageImmersiveUnitCatalog as FieldCatalog, new Set(["locale"]), false); } }

export function normalizeRichTextDocument(document: RichTextDocument, mode = ContentValidationMode.WRITE): RichTextDocument { validateRichTextDocument(document, mode); const raw = mutableRecord(toJson(RichTextDocumentSchema, document)); normalizeRichJSON(mutableRecord(raw.base), (raw.localeOverlays ?? []) as unknown[]); const result = fromJson(RichTextDocumentSchema, raw as JsonValue); validateRichTextDocument(result, mode); return result; }
export function normalizePageDocument(document: PageDocument, mode = ContentValidationMode.WRITE): PageDocument { validatePageDocument(document, mode); const raw = mutableRecord(toJson(PageDocumentSchema, document)); normalizePageJSON(mutableRecord(raw.base), (raw.localeOverlays ?? []) as unknown[]); const result = fromJson(PageDocumentSchema, raw as JsonValue); validatePageDocument(result, mode); return result; }

export function validateRichTextDocument(document: RichTextDocument, mode = ContentValidationMode.WRITE): void { if (document.blockCatalogFingerprint !== contentBlockCatalogFingerprint) fail("$.blockCatalogFingerprint", "fingerprint_mismatch", "fingerprint mismatch"); if (!document.sourceLocale.trim()) fail("$.sourceLocale", "required", "source locale is required"); validateRichGraph(document.profile, required(document.base, "$.base", "base graph is required"), document.localeOverlays, document.sourceLocale, mode, "$"); }
export function validateLocalizedRichTextDocument(document: LocalizedRichTextDocument, mode = ContentValidationMode.RESTORE_SNAPSHOT): void { const overlay = required(document.localeOverlay, "$.localeOverlay", "locale overlay is required"); if (document.blockCatalogFingerprint !== contentBlockCatalogFingerprint || document.locale !== overlay.locale) fail("$", "localized_document", "invalid localized document"); validateRichGraph(document.profile, required(document.base, "$.base", "base graph is required"), [overlay], document.locale, mode, "$"); }

function validatePageGraph(base: NonNullable<PageDocument["base"]>, overlays: readonly PageLocaleOverlay[], sourceLocale: string, mode: ContentValidationMode, path: string): void {
  const byId = new Map<string, PageSection>(); const slots = new Set<string>();
  base.nodes.forEach((node, index) => { const nodePath = \`${"${path}"}.nodes[${"${index}"}]\`; const section = required(node.section, nodePath + ".section", "section is required"); const placement = required(node.placement, nodePath + ".placement", "placement is required"); requireUuid(section.id, nodePath + ".section.id"); if (byId.has(section.id)) fail(nodePath, "duplicate_id", "duplicate section id"); byId.set(section.id, section); if (section.value.case === undefined) fail(nodePath, "required", "section value is required"); const protoCase = section.value.case; const sectionValue = required(section.value.value, nodePath, "section value is required"); const kind = pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase]; if (!kind) fail(nodePath, "kind", "unknown section kind"); validateFields(section.settings, pageSectionSettingsCatalog as FieldCatalog, new Set(["shared"]), nodePath + ".settings", mode); validateFields(sectionValue.props, pageSectionCatalog[kind].fields as FieldCatalog, new Set(["shared", "source"]), nodePath + ".props", mode); if (section.value.case === "immersiveScene") { const ids = section.value.value.units.map((unit) => unit.id); ids.forEach((id) => requireUuid(id, nodePath + ".units.id")); requireUnique(ids, nodePath + ".units"); } const parent = placement.parentSectionId; const column = placement.columnId; if ((parent === undefined) !== (column === undefined)) fail(nodePath + ".placement", "column_placement", "parent and column must be set together"); const slot = (parent ?? "root") + ":" + (column ?? "root") + ":" + placement.index; if (slots.has(slot)) fail(nodePath + ".placement", "duplicate_position", "duplicate section position"); slots.add(slot); });
  for (const node of base.nodes) { const placement = required(node.placement, path + ".placement", "placement is required"); const section = required(node.section, path + ".section", "section is required"); const parentId = placement.parentSectionId; if (!parentId) continue; const parent = required(byId.get(parentId), path, "column parent does not exist"); if (parent.value.case !== "columns") fail(path, "parent", "column parent must be a Columns section"); const columnsPayload = required(parent.value.value, path, "Columns payload is required") as { props?: { columns?: readonly { id: string }[] } }; const kind = pageSectionKindByProtoCase[section.value.case as keyof typeof pageSectionKindByProtoCase]; if (!kind || !pageColumnChildKinds.includes(kind as never)) fail(path, "nesting", "section kind is forbidden inside columns"); const columns = columnsPayload.props?.columns ?? []; if (!columns.some((column) => column.id === placement.columnId)) fail(path, "column", "placement column does not exist"); }
  const locales = overlays.map((overlay) => overlay.locale); requireUnique(locales, path + ".localeOverlays"); if (!locales.includes(sourceLocale)) fail(path, "source_overlay", "source locale overlay is required"); overlays.forEach((overlay) => { const seen = new Set<string>(); overlay.sections.forEach((localized) => { requireUuid(localized.sectionId, path + ".locale.sectionId"); if (seen.has(localized.sectionId)) fail(path, "duplicate_id", "duplicate locale section"); seen.add(localized.sectionId); const baseSection = required(byId.get(localized.sectionId), path, "locale section has no base section"); if (localized.value.case === undefined || localized.value.case !== baseSection.value.case) fail(path, "kind_mismatch", "locale section must match base section"); const protoCase = localized.value.case; const localizedValue = required(localized.value.value, path, "locale section payload is required"); const kind = pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase]; if (!kind) fail(path, "kind", "unknown locale section kind"); validateFields(localizedValue.props, pageSectionCatalog[kind].fields as FieldCatalog, new Set(["locale"]), path + ".locale.props", mode); if (localized.value.case === "immersiveScene" && baseSection.value.case === "immersiveScene") { const baseIds = new Set(baseSection.value.value.units.map((unit) => unit.id)); const unitIds = localized.value.value.units.map((unit) => unit.unitId); requireUnique(unitIds, path + ".locale.units"); if (unitIds.some((id) => !baseIds.has(id))) fail(path, "orphan", "locale immersive unit has no base unit"); } }); });
}

export function validatePageDocument(document: PageDocument, mode = ContentValidationMode.WRITE): void { if (document.blockCatalogFingerprint !== contentBlockCatalogFingerprint || !document.sourceLocale.trim()) fail("$", "required", "invalid Page document envelope"); validatePageGraph(required(document.base, "$.base", "base graph is required"), document.localeOverlays, document.sourceLocale, mode, "$"); }
export function validateLocalizedPageDocument(document: LocalizedPageDocument, mode = ContentValidationMode.RESTORE_SNAPSHOT): void { const overlay = required(document.localeOverlay, "$.localeOverlay", "locale overlay is required"); if (document.blockCatalogFingerprint !== contentBlockCatalogFingerprint || document.locale !== overlay.locale) fail("$", "localized_document", "invalid localized Page document"); validatePageGraph(required(document.base, "$.base", "base graph is required"), [overlay], document.locale, mode, "$"); }

function validateMembers(values: readonly string[], path: string): void { if (values.length === 0) fail(path, "required", "at least one contributor Member is required"); values.forEach((value, index) => requireUuid(value, \`${"${path}"}[${"${index}"}]\`)); requireUnique(values, path); }
export function validateRichTextBlockMutationBatch(batch: RichTextBlockMutationBatch, mode = ContentValidationMode.WRITE): void { void flattenRichTextMutationStorage(batch, mode, false); }
export function validatePageSectionMutationBatch(batch: PageSectionMutationBatch, mode = ContentValidationMode.WRITE): void { if (batch.blockCatalogFingerprint !== contentBlockCatalogFingerprint) fail("$", "fingerprint_mismatch", "fingerprint mismatch"); requireUuid(batch.expectedRevision, "$.expectedRevision"); validateMembers(batch.contributorMemberIds, "$.contributorMemberIds"); requireUnique(batch.localeMutationGroups.map((group) => group.locale), "$.localeMutationGroups"); if (batch.baseMutations.length + batch.localeMutationGroups.reduce((sum, group) => sum + group.mutations.length, 0) === 0) fail("$.mutations", "required", "at least one mutation is required"); for (const mutation of batch.baseMutations) if (mutation.operation.case === "upsert") { const node = required(mutation.operation.value.node, "$.baseMutations", "upsert node is required"); const section = required(node.section, "$.baseMutations.section", "upsert section is required"); required(node.placement, "$.baseMutations.placement", "upsert placement is required"); if (section.value.case === "richText" && section.value.value.blocks?.nodes.length) fail("$.baseMutations", "descendant_ambiguity", "section upsert cannot contain rich-text descendants"); } void mode; }

export function materializeLocalizedRichTextDocument(document: RichTextDocument, locale: string): LocalizedRichTextDocument { validateRichTextDocument(document, ContentValidationMode.RESTORE_SNAPSHOT); const copy = clone(RichTextDocumentSchema, document); return { $typeName: "api.content.v1.LocalizedRichTextDocument", blockCatalogFingerprint: document.blockCatalogFingerprint, profile: document.profile, locale, base: required(copy.base, "$.base", "base graph is required"), localeOverlay: required(copy.localeOverlays.find((item) => item.locale === locale), "$.locale", "locale overlay not found") }; }
export function materializeLocalizedPageDocument(document: PageDocument, locale: string): LocalizedPageDocument { validatePageDocument(document, ContentValidationMode.RESTORE_SNAPSHOT); const copy = clone(PageDocumentSchema, document); return { $typeName: "api.content.v1.LocalizedPageDocument", blockCatalogFingerprint: document.blockCatalogFingerprint, locale, base: required(copy.base, "$.base", "base graph is required"), localeOverlay: required(copy.localeOverlays.find((item) => item.locale === locale), "$.locale", "locale overlay not found") }; }

export type ContentBlockStorageFamily = "richText" | "pageSection";
export interface ContentBlockStorageLocale { readonly locale: string; readonly data: RichTextBlockLocaleData | PageSectionLocaleData }
export interface ContentBlockStorageRow { readonly family: ContentBlockStorageFamily; readonly blockId: string; readonly parentBlockId?: string; readonly containerSlot: string; readonly position: number; readonly kind: string; readonly sharedData: RichTextBlockData | PageSectionData; readonly locales: readonly ContentBlockStorageLocale[] }
export interface ContentStorageFileReference { readonly referencePath: string; readonly fileId: string; readonly missing: boolean; readonly missingAttachmentMediaKind: MissingAttachmentMediaKind; readonly allowedMimeTypes: readonly string[]; readonly allowedMimePrefixes: readonly string[] }
export interface ValidatedContentStorageShared { readonly sharedData: RichTextBlockData | PageSectionData; readonly fileReferences: readonly ContentStorageFileReference[] }
export interface ValidatedContentStorageBlock extends ValidatedContentStorageShared { readonly localizedData: RichTextBlockLocaleData | PageSectionLocaleData }
export interface ContentStorageMove { readonly blockId: string; readonly parentBlockId?: string; readonly containerSlot: string; readonly position: number }
export interface ContentStorageLocaleUpsert { readonly blockId: string; readonly expectedKind: string; readonly localizedData: RichTextBlockLocaleData | PageSectionLocaleData }
export interface ContentStorageLocaleMutationGroup { readonly locale: string; readonly upserts: readonly ContentStorageLocaleUpsert[]; readonly deletes: readonly string[] }
export interface ContentStorageMutationBatch { readonly expectedRevision: string; readonly baseUpserts: readonly ContentBlockStorageRow[]; readonly deletes: readonly string[]; readonly moves: readonly ContentStorageMove[]; readonly localeGroups: readonly ContentStorageLocaleMutationGroup[]; readonly contributorMemberIds: readonly string[] }

function contentStorageAttachmentReference(attachment: FileAttachment | undefined, referencePath: string, policyName: keyof typeof contentBlockFileReferencePolicies): ContentStorageFileReference | undefined {
  if (attachment?.state.case === undefined) return undefined;
  const policy = contentBlockFileReferencePolicies[policyName];
  const allowedMimeTypes = "mime_types" in policy ? [...policy.mime_types] : [];
  const allowedMimePrefixes = "mime_prefixes" in policy ? [...policy.mime_prefixes] : [];
  if (attachment.state.case === "activeFileId") {
    requireUuid(attachment.state.value, "$." + referencePath + ".activeFileId");
    return { referencePath, fileId: attachment.state.value.toLowerCase(), missing: false, missingAttachmentMediaKind: MissingAttachmentMediaKind.UNSPECIFIED, allowedMimeTypes, allowedMimePrefixes };
  }
  const missing = required(attachment.state.value, "$." + referencePath, "missing attachment is required");
  requireUuid(missing.formerFileId, "$." + referencePath + ".missingAttachment.formerFileId");
  if (!validMissingAttachmentMediaKinds.has(missing.mediaKind)) fail("$." + referencePath + ".missingAttachment.mediaKind", "required", "missing attachment media kind is required");
  return { referencePath, fileId: missing.formerFileId.toLowerCase(), missing: true, missingAttachmentMediaKind: missing.mediaKind, allowedMimeTypes, allowedMimePrefixes };
}

export function extractContentStorageFileReferences(kind: string, sharedData: RichTextBlockData | PageSectionData): ContentStorageFileReference[] {
  const protoCase = required(sharedData.value.case, "$.sharedData.value", "typed storage value is required");
  const actualKind = (richTextBlockKindByProtoCase as Readonly<Record<string, string>>)[protoCase] ?? (pageSectionKindByProtoCase as Readonly<Record<string, string>>)[protoCase];
  if (actualKind !== kind) fail("$.kind", "kind_mismatch", "row payload kind does not match storage kind");
  const result: ContentStorageFileReference[] = [];
  const append = (attachment: FileAttachment | undefined, referencePath: string, policyName: keyof typeof contentBlockFileReferencePolicies): void => { const reference = contentStorageAttachmentReference(attachment, referencePath, policyName); if (reference) result.push(reference); };
  if (sharedData.value.case === "file") append(sharedData.value.value.props?.attachment, "file", "file");
  if (sharedData.value.case === "paragraph") { append(sharedData.value.value.props?.linkThumbnail, "paragraph.linkThumbnail", "link_thumbnail"); append(sharedData.value.value.props?.linkIcon, "paragraph.linkIcon", "link_icon"); }
  if (sharedData.value.case === "shader") sharedData.value.value.props?.stages.forEach((stage, stageIndex) => stage.channels.forEach((channel, channelIndex) => { append(channel.file, "shader.stages." + stageIndex + ".channels." + channelIndex + ".file", "shader_channel"); channel.faces.forEach((face, faceIndex) => append(face, "shader.stages." + stageIndex + ".channels." + channelIndex + ".faces." + faceIndex, "shader_cubemap_face")); }));
  if (sharedData.value.case === "immersiveScene") sharedData.value.value.units.forEach((unit) => { const props = unit.props; for (const [field, role, policy] of [["meshFile", "mesh", "immersive.mesh"], ["meshOptimizationSourceFile", "optimization_source", "immersive.optimization_source"], ["meshOptimizationFile", "optimized_mesh", "immersive.optimized_mesh"], ["textureFile", "texture", "immersive.texture"], ["darkTextureFile", "dark_texture", "immersive.dark_texture"]] as const) append(props?.[field], "immersive_scene:" + unit.id + ":" + role, policy); });
  return result.sort((left, right) => left.referencePath.localeCompare(right.referencePath));
}

function rejectContentStorageRuntimeFields(value: unknown, path: string): void {
  if (Array.isArray(value)) { value.forEach((item, index) => rejectContentStorageRuntimeFields(item, path + "[" + index + "]")); return; }
  if (value === null || typeof value !== "object") return;
  for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
    if ((contentBlockRuntimeForbiddenFields as readonly string[]).includes(key)) fail(path + "." + key, "runtime_field", "runtime projection field is forbidden in shared storage");
    rejectContentStorageRuntimeFields(child, path + "." + key);
  }
}

function normalizePresentStorageFields(value: Record<string, unknown>, fields: FieldCatalog, ownerships: ReadonlySet<Ownership>): void {
  for (const [name, spec] of Object.entries(fields)) {
    if (!ownerships.has(spec.ownership ?? "shared")) continue;
    const current = value[name]; if (current === undefined || current === null) continue;
    if (typeof current === "number" && Object.is(current, -0)) value[name] = 0;
    if (spec.type === "uuid" && typeof current === "string") value[name] = current.toLowerCase();
    if ((spec.type === "editor_color" || spec.type === "hex_color") && typeof current === "string" && current.startsWith("#")) value[name] = current.toLowerCase();
    if (spec.type === "array" && Array.isArray(current)) { const item = spec.items as FieldSpec; if (item.type === "object") current.forEach((entry) => normalizePresentStorageFields(mutableRecord(entry), item.fields as FieldCatalog, ownerships)); if (spec.unique === true) current.sort((left, right) => JSON.stringify(left).localeCompare(JSON.stringify(right))); }
    if (spec.type === "object") normalizePresentStorageFields(mutableRecord(current), spec.fields as FieldCatalog, ownerships);
  }
}

function normalizeCanonicalStorageJSONFields(value: Record<string, unknown>, fields: FieldCatalog, ownerships: ReadonlySet<Ownership>): void {
  for (const [name, spec] of Object.entries(fields)) {
    if (!ownerships.has(spec.ownership ?? "shared")) continue;
    const current = value[name]; if (current === undefined || current === null) continue;
    const defaultMatches = Object.hasOwn(spec, "default") && (JSON.stringify(current) === JSON.stringify(spec.default) || ((spec.type === "enum" || spec.type === "enum_int") && typeof current === "string" && current.endsWith("_" + runtimeEnumSuffix(spec.type, spec.default))));
    if (defaultMatches) { delete value[name]; continue; }
    if (typeof current === "number" && Object.is(current, -0)) value[name] = 0;
    if ((spec.type === "editor_color" || spec.type === "hex_color") && typeof current === "string" && current.startsWith("#")) value[name] = current.toLowerCase();
    if (spec.type === "array" && Array.isArray(current)) { const item = spec.items as FieldSpec; if (item.type === "object") current.forEach((entry) => normalizeCanonicalStorageJSONFields(mutableRecord(entry), item.fields as FieldCatalog, ownerships)); if (spec.unique === true) current.sort((left, right) => JSON.stringify(left).localeCompare(JSON.stringify(right))); }
    if (spec.type === "object") normalizeCanonicalStorageJSONFields(mutableRecord(current), spec.fields as FieldCatalog, ownerships);
  }
}

export function normalizeContentStorageShared(profile: keyof typeof richTextProfiles, kind: string, sharedData: RichTextBlockData | PageSectionData, mode = ContentValidationMode.WRITE): ValidatedContentStorageShared {
  if (!(profile in richTextProfiles)) fail("$.profile", "profile", "unknown rich-text profile");
  rejectContentStorageRuntimeFields(sharedData, "$.shared");
  const rich = sharedData.$typeName === "api.content.v1.RichTextBlockData";
  if (rich) {
    if (!richTextBlockKinds.includes(kind as RichTextBlockKind) || !richTextProfiles[profile].blocks.includes(kind as never)) fail("$.kind", "profile", "Block kind is forbidden by profile");
    const copy = clone(RichTextBlockDataSchema, sharedData as RichTextBlockData);
    const protoCase = required(copy.value.case, "$.shared", "shared payload kind is required");
    if (richTextBlockKindByProtoCase[protoCase as keyof typeof richTextBlockKindByProtoCase] !== kind) fail("$.kind", "kind_mismatch", "shared payload kind does not match storage kind");
    const payload = mutableRecord(copy.value.value); const props = mutableRecord(payload.props);
    validateFields(props, richTextBlockCatalog[kind as RichTextBlockKind].fields as FieldCatalog, new Set(["shared", "source"]), "$.shared.props", mode);
    validateProfileLinkMedia(kind, props, profile, "$.shared.props");
    normalizePresentStorageFields(props, richTextBlockCatalog[kind as RichTextBlockKind].fields as FieldCatalog, new Set(["shared", "source"]));
    const raw = mutableRecord(toJson(RichTextBlockDataSchema, copy)); const rawPayload = mutableRecord(raw[protoCase]); normalizeCanonicalStorageJSONFields(mutableRecord(rawPayload.props), richTextBlockCatalog[kind as RichTextBlockKind].fields as FieldCatalog, new Set(["shared", "source"]));
    const normalized = fromJson(RichTextBlockDataSchema, raw as JsonValue);
    return { sharedData: normalized, fileReferences: extractContentStorageFileReferences(kind, normalized) };
  }
  if (profile !== "page" || !pageSectionKinds.includes(kind as PageSectionKind)) fail("$.kind", "kind", "unknown Page storage kind");
  const copy = clone(PageSectionDataSchema, sharedData as PageSectionData);
  const protoCase = required(copy.value.case, "$.shared", "shared payload kind is required");
  if (pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase] !== kind) fail("$.kind", "kind_mismatch", "shared payload kind does not match storage kind");
  const settings = mutableRecord(copy.settings);
  validateFields(settings, pageSectionSettingsCatalog as FieldCatalog, new Set(["shared"]), "$.shared.settings", mode);
  normalizePresentStorageFields(settings, pageSectionSettingsCatalog as FieldCatalog, new Set(["shared"]));
  const payload = mutableRecord(copy.value.value); const props = mutableRecord(payload.props);
  validateFields(props, pageSectionCatalog[kind as PageSectionKind].fields as FieldCatalog, new Set(["shared", "source"]), "$.shared.props", mode);
  normalizePresentStorageFields(props, pageSectionCatalog[kind as PageSectionKind].fields as FieldCatalog, new Set(["shared", "source"]));
  if (kind === "rich-text" && payload.blocks !== undefined && Object.keys(mutableRecord(payload.blocks)).length !== 0) fail("$.shared.richText.blocks", "nested_authority", "rich-text descendants must be separate storage rows");
  if (kind === "immersive-scene") {
    const units = (payload.units ?? []) as unknown[]; const seen = new Set<string>();
    units.forEach((rawUnit, index) => { const unit = mutableRecord(rawUnit); const id = String(unit.id ?? ""); requireUuid(id, "$.shared.units[" + index + "].id"); const normalizedID = id.toLowerCase(); if (seen.has(normalizedID)) fail("$.shared.units", "duplicate", "unit IDs must be unique"); seen.add(normalizedID); unit.id = normalizedID; const unitProps = mutableRecord(unit.props); validateFields(unitProps, pageImmersiveUnitCatalog as FieldCatalog, new Set(["shared", "source"]), "$.shared.units[" + index + "].props", mode); normalizePresentStorageFields(unitProps, pageImmersiveUnitCatalog as FieldCatalog, new Set(["shared", "source"])); });
  }
  const raw = mutableRecord(toJson(PageSectionDataSchema, copy)); normalizeCanonicalStorageJSONFields(mutableRecord(raw.settings), pageSectionSettingsCatalog as FieldCatalog, new Set(["shared"])); const rawPayload = mutableRecord(raw[protoCase]); normalizeCanonicalStorageJSONFields(mutableRecord(rawPayload.props), pageSectionCatalog[kind as PageSectionKind].fields as FieldCatalog, new Set(["shared", "source"])); if (kind === "immersive-scene") for (const rawUnit of (rawPayload.units ?? []) as unknown[]) normalizeCanonicalStorageJSONFields(mutableRecord(mutableRecord(rawUnit).props), pageImmersiveUnitCatalog as FieldCatalog, new Set(["shared", "source"]));
  const normalized = fromJson(PageSectionDataSchema, raw as JsonValue);
  return { sharedData: normalized, fileReferences: extractContentStorageFileReferences(kind, normalized) };
}

export function normalizeContentStorageBlock(profile: keyof typeof richTextProfiles, kind: string, sharedData: RichTextBlockData | PageSectionData, localizedData: RichTextBlockLocaleData | PageSectionLocaleData, mode = ContentValidationMode.RESTORE_SNAPSHOT): ValidatedContentStorageBlock {
  const shared = normalizeContentStorageShared(profile, kind, sharedData, mode);
  return { ...shared, localizedData: normalizeContentStorageLocale(profile, kind, localizedData, mode) };
}
function richStorageData(block: RichTextBlock): RichTextBlockData { return { $typeName: "api.content.v1.RichTextBlockData", value: block.value }; }
function richStorageLocaleData(block: RichTextBlockLocale): RichTextBlockLocaleData { return { $typeName: "api.content.v1.RichTextBlockLocaleData", value: block.value }; }
function pageStorageData(section: PageSection): PageSectionData { const value = section.value.case === "richText" ? { case: "richText" as const, value: { ...section.value.value, blocks: undefined } } : section.value; return { $typeName: "api.content.v1.PageSectionData", settings: section.settings, value }; }
function pageStorageLocaleData(section: PageSectionLocale): PageSectionLocaleData { const value = section.value.case === "richText" ? { case: "richText" as const, value: { ...section.value.value, blocks: undefined } } : section.value; return { $typeName: "api.content.v1.PageSectionLocaleData", value }; }
export function normalizeContentStorageLocale(profile: keyof typeof richTextProfiles, kind: string, localizedData: RichTextBlockLocaleData | PageSectionLocaleData, mode = ContentValidationMode.WRITE): RichTextBlockLocaleData | PageSectionLocaleData {
  const rich = localizedData.$typeName === "api.content.v1.RichTextBlockLocaleData";
  if (rich) {
    if (!richTextBlockKinds.includes(kind as RichTextBlockKind) || !richTextProfiles[profile].blocks.includes(kind as never)) fail("$.kind", "profile", "Block kind is forbidden by profile");
    const typedData = localizedData as RichTextBlockLocaleData;
    const protoCase = required(typedData.value.case, "$.locale", "locale payload kind is required");
    if (richTextBlockKindByProtoCase[protoCase as keyof typeof richTextBlockKindByProtoCase] !== kind) fail("$.kind", "kind_mismatch", "locale payload kind does not match expected storage kind");
    const typedPayload = required(typedData.value.value, "$.locale", "locale payload is required");
    validateLocalizedRichContent(typedPayload, kind as RichTextBlockKind, profile, "$.locale", mode);
    const raw = richStorageJSON(typedData);
    const payload = mutableRecord(raw[protoCase]); const props = mutableRecord(payload.props); payload.props = props;
    validateFields(props, richTextBlockCatalog[kind as RichTextBlockKind].fields as FieldCatalog, new Set(["locale"]), "$.locale.props", mode);
    normalizeFieldRecord(props, richTextBlockCatalog[kind as RichTextBlockKind].fields as FieldCatalog, new Set(["locale"]), false);
    normalizeRichInlineJSON(payload.content);
    return fromJson(RichTextBlockLocaleDataSchema, raw as JsonValue);
  }
  if (profile !== "page" || !pageSectionKinds.includes(kind as PageSectionKind)) fail("$.kind", "kind", "unknown Page storage kind");
  const raw = pageStorageJSON(localizedData as PageSectionLocaleData);
  const protoCase = required((localizedData as PageSectionLocaleData).value.case, "$.locale", "locale payload kind is required");
  if (pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase] !== kind) fail("$.kind", "kind_mismatch", "locale payload kind does not match expected storage kind");
  const payload = mutableRecord(raw[protoCase]); const props = mutableRecord(payload.props); payload.props = props;
  validateFields(props, pageSectionCatalog[kind as PageSectionKind].fields as FieldCatalog, new Set(["locale"]), "$.locale.props", mode);
  normalizeFieldRecord(props, pageSectionCatalog[kind as PageSectionKind].fields as FieldCatalog, new Set(["locale"]), false);
  if (kind === "immersive-scene") {
    const units = (payload.units ?? []) as unknown[];
    const seen = new Set<string>();
    units.forEach((rawUnit, index) => {
      const unit = mutableRecord(rawUnit);
      const id = String(unit.unitId ?? "");
      requireUuid(id, "$.locale.units[" + index + "].unitId");
      const normalizedID = id.toLowerCase();
      if (seen.has(normalizedID)) fail("$.locale.units", "duplicate", "unit IDs must be unique");
      seen.add(normalizedID);
      unit.unitId = normalizedID;
      const unitProps = mutableRecord(unit.props);
      unit.props = unitProps;
      validateFields(unitProps, pageImmersiveUnitCatalog as FieldCatalog, new Set(["locale"]), "$.locale.units[" + index + "].props", mode);
      normalizeFieldRecord(unitProps, pageImmersiveUnitCatalog as FieldCatalog, new Set(["locale"]), false);
    });
  }
  if (kind === "rich-text" && Object.keys(mutableRecord(payload.blocks)).length !== 0) fail("$.locale.richText.blocks", "nested_authority", "rich-text descendants must be separate storage rows");
  return fromJson(PageSectionLocaleDataSchema, raw as JsonValue);
}
function compareStorageRows(left: ContentBlockStorageRow, right: ContentBlockStorageRow): number {
  return (left.parentBlockId ?? "").localeCompare(right.parentBlockId ?? "") ||
    left.containerSlot.localeCompare(right.containerSlot) ||
    left.position - right.position ||
    left.blockId.localeCompare(right.blockId);
}
export function flattenRichTextDocumentStorage(document: RichTextDocument, mode = ContentValidationMode.WRITE): ContentBlockStorageRow[] {
  const normalized = normalizeRichTextDocument(document, mode);
  const profile = profileName(normalized.profile);
  const locales = new Map<string, Map<string, RichTextBlockLocale>>(normalized.localeOverlays.map((overlay) => [overlay.locale, new Map(overlay.blocks.map((block) => [block.blockId, block]))]));
  return required(normalized.base, "$.base", "base graph is required").nodes.map((node) => {
    const block = required(node.block, "$.node.block", "block is required");
    const placement = required(node.placement, "$.node.placement", "placement is required");
    const catalogKind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase];
    if (!catalogKind) fail("$.node.block", "kind", "unknown rich-text Block kind");
    const shared = normalizeContentStorageShared(profile, catalogKind, richStorageData(block), mode);
    return { family: "richText", blockId: block.id, parentBlockId: placement.parentBlockId, containerSlot: "content", position: placement.index, kind: catalogKind, sharedData: shared.sharedData, locales: [...locales].flatMap(([locale, values]) => { const value = values.get(block.id); return value ? [{ locale, data: normalizeContentStorageLocale(profile, catalogKind, richStorageLocaleData(value), mode) as RichTextBlockLocaleData }] : []; }) };
  });
}
export function flattenPageDocumentStorage(document: PageDocument, mode = ContentValidationMode.WRITE): ContentBlockStorageRow[] {
  const normalized = normalizePageDocument(document, mode);
  const locales = new Map<string, Map<string, PageSectionLocale>>(normalized.localeOverlays.map((overlay) => [overlay.locale, new Map(overlay.sections.map((section) => [section.sectionId, section]))]));
  const rows: ContentBlockStorageRow[] = [];
  for (const node of required(normalized.base, "$.base", "base graph is required").nodes) {
    const section = required(node.section, "$.node.section", "section is required");
    const placement = required(node.placement, "$.node.placement", "placement is required");
    const catalogKind = pageSectionKindByProtoCase[section.value.case as keyof typeof pageSectionKindByProtoCase];
    if (!catalogKind) fail("$.node.section", "kind", "unknown section kind");
    const shared = normalizeContentStorageShared("page", catalogKind, pageStorageData(section), mode);
    rows.push({ family: "pageSection", blockId: section.id, parentBlockId: placement.parentSectionId, containerSlot: placement.columnId ? "column-" + placement.columnId : "sections", position: placement.index, kind: catalogKind, sharedData: shared.sharedData, locales: [...locales].flatMap(([locale, values]) => { const value = values.get(section.id); return value ? [{ locale, data: normalizeContentStorageLocale("page", catalogKind, pageStorageLocaleData(value), mode) as PageSectionLocaleData }] : []; }) });
    if (section.value.case !== "richText" || !section.value.value.blocks) continue;
    const nested: RichTextDocument = { $typeName: "api.content.v1.RichTextDocument", blockCatalogFingerprint: contentBlockCatalogFingerprint, profile: RichTextProfile.PAGE, sourceLocale: normalized.sourceLocale, base: section.value.value.blocks, localeOverlays: normalized.localeOverlays.map((overlay) => { const localized = locales.get(overlay.locale)?.get(section.id); return localized?.value.case === "richText" && localized.value.value.blocks ? localized.value.value.blocks : { $typeName: "api.content.v1.RichTextLocaleOverlay", locale: overlay.locale, blocks: [] }; }) };
    for (const inner of flattenRichTextDocumentStorage(nested, mode)) rows.push({ ...inner, parentBlockId: inner.parentBlockId ?? section.id, containerSlot: "content" });
  }
  return rows;
}

function mutationContributors(values: readonly string[], system: boolean): string[] {
  if (system) { if (values.length !== 0) fail("$.contributorMemberIds", "system_contributors", "system mutation must not attribute contributor Members"); return []; }
  validateMembers(values, "$.contributorMemberIds");
  return values.map((value) => value.toLowerCase());
}
function flattenRichTextMutationStorage(batch: RichTextBlockMutationBatch, mode: ContentValidationMode, system: boolean): ContentStorageMutationBatch {
  if (batch.blockCatalogFingerprint !== contentBlockCatalogFingerprint) fail("$.blockCatalogFingerprint", "fingerprint_mismatch", "fingerprint mismatch");
  requireUuid(batch.expectedRevision, "$.expectedRevision");
  const profile = profileName(batch.profile);
  const baseUpserts: ContentBlockStorageRow[] = []; const deletes: string[] = []; const moves: ContentStorageMove[] = []; const localeGroups: ContentStorageLocaleMutationGroup[] = [];
  for (const mutation of batch.baseMutations) {
    if (mutation.operation.case === "upsert") {
      const node = required(mutation.operation.value.node, "$.baseMutations", "upsert node is required"); const block = required(node.block, "$.baseMutations.block", "Block is required"); const placement = required(node.placement, "$.baseMutations.placement", "placement is required");
      const kind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!kind) fail("$.kind", "kind", "unknown rich-text Block kind");
      const shared = normalizeContentStorageShared(profile, kind, richStorageData(block), mode);
      baseUpserts.push({ family: "richText", blockId: block.id.toLowerCase(), parentBlockId: placement.parentBlockId?.toLowerCase(), containerSlot: "content", position: placement.index, kind, sharedData: shared.sharedData, locales: [] });
    } else if (mutation.operation.case === "delete") deletes.push(mutation.operation.value.blockId.toLowerCase());
    else if (mutation.operation.case === "move") { const placement = required(mutation.operation.value.placement, "$.baseMutations.move.placement", "placement is required"); moves.push({ blockId: mutation.operation.value.blockId.toLowerCase(), parentBlockId: placement.parentBlockId?.toLowerCase(), containerSlot: "content", position: placement.index }); }
    else fail("$.baseMutations", "operation", "unknown mutation");
  }
  requireUnique(batch.localeMutationGroups.map((group) => group.locale), "$.localeMutationGroups");
  for (const group of batch.localeMutationGroups) {
    if (!group.locale.trim()) fail("$.localeMutationGroups.locale", "required", "locale is required"); const upserts: ContentStorageLocaleUpsert[] = []; const localeDeletes: string[] = [];
    for (const mutation of group.mutations) {
      if (mutation.operation.case === "upsert") { const block = required(mutation.operation.value.block, "$.localeMutationGroups", "locale Block is required"); const expectedKind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!expectedKind) fail("$.kind", "kind", "unknown rich-text locale kind"); const localizedData = normalizeContentStorageLocale(profile, expectedKind, richStorageLocaleData(block), mode) as RichTextBlockLocaleData; upserts.push({ blockId: block.blockId.toLowerCase(), expectedKind, localizedData }); }
      else if (mutation.operation.case === "delete") localeDeletes.push(mutation.operation.value.blockId.toLowerCase());
      else fail("$.localeMutationGroups", "operation", "unknown locale mutation");
    }
    localeGroups.push({ locale: group.locale, upserts, deletes: localeDeletes });
  }
  if (baseUpserts.length + deletes.length + moves.length + localeGroups.length === 0) fail("$.mutations", "required", "at least one mutation is required");
  return { expectedRevision: batch.expectedRevision.toLowerCase(), baseUpserts, deletes, moves, localeGroups, contributorMemberIds: mutationContributors(batch.contributorMemberIds, system) };
}
export function flattenRichTextMutationBatchStorage(batch: RichTextBlockMutationBatch, mode = ContentValidationMode.WRITE): ContentStorageMutationBatch { return flattenRichTextMutationStorage(batch, mode, false); }
export function flattenRichTextSystemMutationBatchStorage(batch: RichTextBlockMutationBatch, mode = ContentValidationMode.WRITE): ContentStorageMutationBatch { return flattenRichTextMutationStorage(batch, mode, true); }

function flattenPageMutationStorage(batch: PageSectionMutationBatch, mode: ContentValidationMode, system: boolean): ContentStorageMutationBatch {
  if (batch.blockCatalogFingerprint !== contentBlockCatalogFingerprint) fail("$.blockCatalogFingerprint", "fingerprint_mismatch", "fingerprint mismatch");
  requireUuid(batch.expectedRevision, "$.expectedRevision");
  const baseUpserts: ContentBlockStorageRow[] = []; const deletes: string[] = []; const moves: ContentStorageMove[] = []; const localeGroups: ContentStorageLocaleMutationGroup[] = [];
  for (const mutation of batch.baseMutations) {
    if (mutation.operation.case === "upsert") { const node = required(mutation.operation.value.node, "$.baseMutations", "upsert node is required"); const section = required(node.section, "$.baseMutations.section", "section is required"); const placement = required(node.placement, "$.baseMutations.placement", "placement is required"); if (section.value.case === "richText" && section.value.value.blocks?.nodes.length) fail("$.baseMutations", "descendant_ambiguity", "section upsert cannot contain rich-text descendants"); const kind = pageSectionKindByProtoCase[section.value.case as keyof typeof pageSectionKindByProtoCase]; if (!kind) fail("$.kind", "kind", "unknown Page section kind"); const shared = normalizeContentStorageShared("page", kind, pageStorageData(section), mode); baseUpserts.push({ family: "pageSection", blockId: section.id.toLowerCase(), parentBlockId: placement.parentSectionId?.toLowerCase(), containerSlot: placement.columnId ? "column-" + placement.columnId.toLowerCase() : "sections", position: placement.index, kind, sharedData: shared.sharedData, locales: [] }); }
    else if (mutation.operation.case === "delete") deletes.push(mutation.operation.value.sectionId.toLowerCase());
    else if (mutation.operation.case === "move") { const placement = required(mutation.operation.value.placement, "$.baseMutations.move.placement", "placement is required"); moves.push({ blockId: mutation.operation.value.sectionId.toLowerCase(), parentBlockId: placement.parentSectionId?.toLowerCase(), containerSlot: placement.columnId ? "column-" + placement.columnId.toLowerCase() : "sections", position: placement.index }); }
    else if (mutation.operation.case === "mutateRichTextBlock") { const sectionId = mutation.operation.value.sectionId.toLowerCase(); const rich = required(mutation.operation.value.mutation, "$.baseMutations", "rich-text mutation is required"); if (rich.operation.case === "upsert") { const node = required(rich.operation.value.node, "$.baseMutations", "rich-text upsert is required"); const block = required(node.block, "$.baseMutations.block", "Block is required"); const placement = required(node.placement, "$.baseMutations.placement", "placement is required"); const kind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!kind) fail("$.kind", "kind", "unknown rich-text Block kind"); const shared = normalizeContentStorageShared("page", kind, richStorageData(block), mode); baseUpserts.push({ family: "richText", blockId: block.id.toLowerCase(), parentBlockId: placement.parentBlockId?.toLowerCase() ?? sectionId, containerSlot: "content", position: placement.index, kind, sharedData: shared.sharedData, locales: [] }); } else if (rich.operation.case === "delete") deletes.push(rich.operation.value.blockId.toLowerCase()); else if (rich.operation.case === "move") { const placement = required(rich.operation.value.placement, "$.baseMutations.move.placement", "placement is required"); moves.push({ blockId: rich.operation.value.blockId.toLowerCase(), parentBlockId: placement.parentBlockId?.toLowerCase() ?? sectionId, containerSlot: "content", position: placement.index }); } else fail("$.baseMutations", "operation", "unknown nested rich-text mutation"); }
    else fail("$.baseMutations", "operation", "unknown Page mutation");
  }
  requireUnique(batch.localeMutationGroups.map((group) => group.locale), "$.localeMutationGroups");
  for (const group of batch.localeMutationGroups) { if (!group.locale.trim()) fail("$.localeMutationGroups.locale", "required", "locale is required"); const upserts: ContentStorageLocaleUpsert[] = []; const localeDeletes: string[] = [];
    for (const mutation of group.mutations) {
      if (mutation.operation.case === "upsert") { const section = required(mutation.operation.value.section, "$.localeMutationGroups", "locale section is required"); const expectedKind = pageSectionKindByProtoCase[section.value.case as keyof typeof pageSectionKindByProtoCase]; if (!expectedKind) fail("$.kind", "kind", "unknown Page locale kind"); const localizedData = normalizeContentStorageLocale("page", expectedKind, pageStorageLocaleData(section), mode) as PageSectionLocaleData; upserts.push({ blockId: section.sectionId.toLowerCase(), expectedKind, localizedData }); }
      else if (mutation.operation.case === "delete") localeDeletes.push(mutation.operation.value.sectionId.toLowerCase());
      else if (mutation.operation.case === "mutateRichTextBlock") { const rich = required(mutation.operation.value.mutation, "$.localeMutationGroups", "rich-text locale mutation is required"); if (rich.operation.case === "upsert") { const block = required(rich.operation.value.block, "$.localeMutationGroups", "locale Block is required"); const expectedKind = richTextBlockKindByProtoCase[block.value.case as keyof typeof richTextBlockKindByProtoCase]; if (!expectedKind) fail("$.kind", "kind", "unknown rich-text locale kind"); const localizedData = normalizeContentStorageLocale("page", expectedKind, richStorageLocaleData(block), mode) as RichTextBlockLocaleData; upserts.push({ blockId: block.blockId.toLowerCase(), expectedKind, localizedData }); } else if (rich.operation.case === "delete") localeDeletes.push(rich.operation.value.blockId.toLowerCase()); else fail("$.localeMutationGroups", "operation", "unknown nested locale mutation"); }
      else fail("$.localeMutationGroups", "operation", "unknown Page locale mutation");
    }
    localeGroups.push({ locale: group.locale, upserts, deletes: localeDeletes });
  }
  if (baseUpserts.length + deletes.length + moves.length + localeGroups.length === 0) fail("$.mutations", "required", "at least one mutation is required");
  return { expectedRevision: batch.expectedRevision.toLowerCase(), baseUpserts, deletes, moves, localeGroups, contributorMemberIds: mutationContributors(batch.contributorMemberIds, system) };
}
export function flattenPageMutationBatchStorage(batch: PageSectionMutationBatch, mode = ContentValidationMode.WRITE): ContentStorageMutationBatch { return flattenPageMutationStorage(batch, mode, false); }
export function flattenPageSystemMutationBatchStorage(batch: PageSectionMutationBatch, mode = ContentValidationMode.WRITE): ContentStorageMutationBatch { return flattenPageMutationStorage(batch, mode, true); }

function richStorageJSON(data: RichTextBlockData | RichTextBlockLocaleData): Record<string, unknown> {
  return mutableRecord(data.$typeName.endsWith("LocaleData") ? toJson(RichTextBlockLocaleDataSchema, data as RichTextBlockLocaleData) : toJson(RichTextBlockDataSchema, data as RichTextBlockData));
}
function pageStorageJSON(data: PageSectionData | PageSectionLocaleData): Record<string, unknown> {
  return mutableRecord(data.$typeName.endsWith("LocaleData") ? toJson(PageSectionLocaleDataSchema, data as PageSectionLocaleData) : toJson(PageSectionDataSchema, data as PageSectionData));
}
export function materializeRichTextDocumentStorage(profile: RichTextProfile, sourceLocale: string, rows: readonly ContentBlockStorageRow[]): RichTextDocument {
  profileName(profile);
  if (!sourceLocale.trim()) fail("$.sourceLocale", "required", "source locale is required");
  const seen = new Set<string>();
  const localeBlocks = new Map<string, Record<string, unknown>[]>();
  localeBlocks.set(sourceLocale, []);
  const nodes = [...rows].sort(compareStorageRows).map((row) => {
    if (row.family !== "richText" || !richTextBlockKinds.includes(row.kind as RichTextBlockKind)) fail("$.kind", "kind", "non-rich row in rich-text document");
    requireUuid(row.blockId, "$.blockId");
    if (seen.has(row.blockId)) fail("$.blockId", "duplicate_id", "duplicate Block ID");
    seen.add(row.blockId);
    if (row.parentBlockId !== undefined) requireUuid(row.parentBlockId, "$.parentBlockId");
    const shared = richStorageJSON(row.sharedData as RichTextBlockData);
    const protoCase = (row.sharedData as RichTextBlockData).value.case;
    if (richTextBlockKindByProtoCase[protoCase as keyof typeof richTextBlockKindByProtoCase] !== row.kind) fail("$.kind", "kind_mismatch", "row payload kind does not match storage kind");
    if (!row.locales.some((locale) => locale.locale === sourceLocale)) fail("$.sourceLocale", "source_overlay", "each base Block requires source locale data");
    for (const locale of row.locales) {
      const values = localeBlocks.get(locale.locale) ?? [];
      values.push({ blockId: row.blockId, ...richStorageJSON(locale.data as RichTextBlockLocaleData) });
      localeBlocks.set(locale.locale, values);
    }
    return { block: { id: row.blockId, ...shared }, placement: { ...(row.parentBlockId === undefined ? {} : { parentBlockId: row.parentBlockId }), index: row.position } };
  });
  const raw = { blockCatalogFingerprint: contentBlockCatalogFingerprint, profile, sourceLocale, base: { nodes }, localeOverlays: [...localeBlocks].sort(([left], [right]) => left.localeCompare(right)).map(([locale, blocks]) => ({ locale, blocks })) } as JsonValue;
  return normalizeRichTextDocument(fromJson(RichTextDocumentSchema, raw), ContentValidationMode.RESTORE_SNAPSHOT);
}
export function materializeLocalizedRichTextDocumentStorage(profile: RichTextProfile, locale: string, rows: readonly ContentBlockStorageRow[]): LocalizedRichTextDocument {
  return materializeLocalizedRichTextDocument(materializeRichTextDocumentStorage(profile, locale, rows), locale);
}
export function materializePageDocumentStorage(sourceLocale: string, rows: readonly ContentBlockStorageRow[]): PageDocument {
  if (!sourceLocale.trim()) fail("$.sourceLocale", "required", "source locale is required");
  const sectionRows = new Map<string, ContentBlockStorageRow>();
  const richRows = new Map<string, ContentBlockStorageRow>();
  for (const row of rows) {
    requireUuid(row.blockId, "$.blockId");
    if (sectionRows.has(row.blockId) || richRows.has(row.blockId)) fail("$.blockId", "duplicate_id", "duplicate Block ID");
    if (row.family === "pageSection" && pageSectionKinds.includes(row.kind as PageSectionKind)) sectionRows.set(row.blockId, row);
    else if (row.family === "richText" && richTextBlockKinds.includes(row.kind as RichTextBlockKind)) richRows.set(row.blockId, row);
    else fail("$.kind", "kind", "unknown Page storage row kind");
  }
  const localeSections = new Map<string, Map<string, Record<string, unknown>>>();
  localeSections.set(sourceLocale, new Map());
  const sectionNodes = new Map<string, Record<string, unknown>>();
  for (const row of [...sectionRows.values()].sort(compareStorageRows)) {
    const shared = pageStorageJSON(row.sharedData as PageSectionData);
    const protoCase = (row.sharedData as PageSectionData).value.case;
    if (pageSectionKindByProtoCase[protoCase as keyof typeof pageSectionKindByProtoCase] !== row.kind) fail("$.kind", "kind_mismatch", "row payload kind does not match storage kind");
    if (!row.locales.some((locale) => locale.locale === sourceLocale)) fail("$.sourceLocale", "source_overlay", "each Page section requires source locale data");
    const placement: Record<string, unknown> = { index: row.position };
    if (row.parentBlockId !== undefined) {
      placement.parentSectionId = row.parentBlockId;
      if (!row.containerSlot.startsWith("column-")) fail("$.containerSlot", "slot", "nested Page section must name a column");
      placement.columnId = row.containerSlot.slice("column-".length);
    } else if (row.containerSlot !== "sections") fail("$.containerSlot", "slot", "root Page section must use sections slot");
    const node = { section: { id: row.blockId, ...shared }, placement };
    sectionNodes.set(row.blockId, node);
    for (const locale of row.locales) {
      const values = localeSections.get(locale.locale) ?? new Map<string, Record<string, unknown>>();
      values.set(row.blockId, { sectionId: row.blockId, ...pageStorageJSON(locale.data as PageSectionLocaleData) });
      localeSections.set(locale.locale, values);
    }
  }
  const consumedRich = new Set<string>();
  for (const [sectionID, sectionRow] of sectionRows) {
    if (sectionRow.kind !== "rich-text") continue;
    const descendants: ContentBlockStorageRow[] = [];
    for (const candidate of richRows.values()) {
      let parent = candidate.parentBlockId;
      const ancestry = new Set<string>();
      while (parent !== undefined) {
        if (parent === sectionID) { descendants.push({ ...candidate, parentBlockId: candidate.parentBlockId === sectionID ? undefined : candidate.parentBlockId }); consumedRich.add(candidate.blockId); break; }
        if (ancestry.has(parent)) fail("$.parentBlockId", "cycle", "rich-text storage parent cycle");
        ancestry.add(parent);
        parent = richRows.get(parent)?.parentBlockId;
      }
    }
    const rich = materializeRichTextDocumentStorage(RichTextProfile.PAGE, sourceLocale, descendants);
    const richJSON = mutableRecord(toJson(RichTextDocumentSchema, rich));
    const node = required(sectionNodes.get(sectionID), "$.section", "rich-text section row is required");
    const section = mutableRecord(node.section);
    const sharedRich = mutableRecord(section.richText);
    sharedRich.blocks = richJSON.base;
    section.richText = sharedRich;
    for (const rawOverlay of (richJSON.localeOverlays ?? []) as unknown[]) {
      const richOverlay = mutableRecord(rawOverlay);
      const locale = String(richOverlay.locale);
      const values = localeSections.get(locale) ?? new Map<string, Record<string, unknown>>();
      const localized = values.get(sectionID) ?? { sectionId: sectionID, richText: {} };
      const localizedRich = mutableRecord(localized.richText);
      localizedRich.blocks = richOverlay;
      localized.richText = localizedRich;
      values.set(sectionID, localized);
      localeSections.set(locale, values);
    }
  }
  if (consumedRich.size !== richRows.size) fail("$.parentBlockId", "orphan", "rich-text Block is not owned by a Page rich-text section");
  const raw = { blockCatalogFingerprint: contentBlockCatalogFingerprint, sourceLocale, base: { nodes: [...sectionNodes.values()] }, localeOverlays: [...localeSections].sort(([left], [right]) => left.localeCompare(right)).map(([locale, sections]) => ({ locale, sections: [...sections.values()].sort((left, right) => String(left.sectionId).localeCompare(String(right.sectionId))) })) } as JsonValue;
  return normalizePageDocument(fromJson(PageDocumentSchema, raw), ContentValidationMode.RESTORE_SNAPSHOT);
}
export function materializeLocalizedPageDocumentStorage(locale: string, rows: readonly ContentBlockStorageRow[]): LocalizedPageDocument {
  return materializeLocalizedPageDocument(materializePageDocumentStorage(locale, rows), locale);
}

function activeReference(attachment: FileAttachment | undefined, blockId: string, referencePath: string): ContentBlockMediaReference | undefined { return attachment?.state.case === "activeFileId" ? { $typeName: "api.content.v1.ContentBlockMediaReference", blockId, referencePath, activeFileId: attachment.state.value } : undefined; }
export function extractRichTextFileReferences(document: RichTextDocument): ContentBlockMediaReference[] { validateRichTextDocument(document, ContentValidationMode.RESTORE_SNAPSHOT); const result: ContentBlockMediaReference[] = []; for (const node of document.base!.nodes) { const block = node.block!; if (block.value.case === "file") { const ref = activeReference(block.value.value.props?.attachment, block.id, "file"); if (ref) result.push(ref); } if (block.value.case === "paragraph") { for (const [attachment, path] of [[block.value.value.props?.linkThumbnail, "paragraph.linkThumbnail"], [block.value.value.props?.linkIcon, "paragraph.linkIcon"]] as const) { const ref = activeReference(attachment, block.id, path); if (ref) result.push(ref); } } if (block.value.case === "shader") block.value.value.props?.stages.forEach((stage, stageIndex) => stage.channels.forEach((channel, channelIndex) => { const one = activeReference(channel.file, block.id, \`shader.stages.${"${stageIndex}"}.channels.${"${channelIndex}"}.file\`); if (one) result.push(one); channel.faces.forEach((face, faceIndex) => { const ref = activeReference(face, block.id, \`shader.stages.${"${stageIndex}"}.channels.${"${channelIndex}"}.faces.${"${faceIndex}"}\`); if (ref) result.push(ref); }); })); } return result; }
export function extractPageFileReferences(document: PageDocument): ContentBlockMediaReference[] { validatePageDocument(document, ContentValidationMode.RESTORE_SNAPSHOT); const result: ContentBlockMediaReference[] = []; for (const node of document.base!.nodes) { const section = node.section!; if (section.value.case === "richText" && section.value.value.blocks) { const rich: RichTextDocument = { $typeName: "api.content.v1.RichTextDocument", blockCatalogFingerprint: contentBlockCatalogFingerprint, profile: RichTextProfile.PAGE, sourceLocale: document.sourceLocale, base: section.value.value.blocks, localeOverlays: document.localeOverlays.flatMap((overlay) => { const localized = overlay.sections.find((item) => item.sectionId === section.id); return localized?.value.case === "richText" && localized.value.value.blocks ? [localized.value.value.blocks] : []; }) }; result.push(...extractRichTextFileReferences(rich)); } if (section.value.case === "immersiveScene") section.value.value.units.forEach((unit) => { const props = unit.props; for (const [field, path] of [["meshFile","immersive.mesh"],["meshOptimizationSourceFile","immersive.optimization_source"],["meshOptimizationFile","immersive.optimized_mesh"],["textureFile","immersive.texture"],["darkTextureFile","immersive.dark_texture"]] as const) { const ref = activeReference(props?.[field], section.id, path + "." + unit.id); if (ref) result.push(ref); } }); } return result; }

function canonicalStorageData(row: ContentBlockStorageRow, data: RichTextBlockData | PageSectionData | RichTextBlockLocaleData | PageSectionLocaleData): Record<string, unknown> { const value = row.family === "richText" ? (data.$typeName.endsWith("LocaleData") ? toJson(RichTextBlockLocaleDataSchema, data as RichTextBlockLocaleData) : toJson(RichTextBlockDataSchema, data as RichTextBlockData)) : (data.$typeName.endsWith("LocaleData") ? toJson(PageSectionLocaleDataSchema, data as PageSectionLocaleData) : toJson(PageSectionDataSchema, data as PageSectionData)); return Object.fromEntries(Object.entries(mutableRecord(value)).sort(([left], [right]) => left.localeCompare(right))); }
export function canonicalStorageDocumentBytes(profile: string, rows: readonly ContentBlockStorageRow[]): Uint8Array { const blocks = [...rows].sort((left, right) => left.blockId.localeCompare(right.blockId)).map((row) => ({ id: row.blockId, parent_id: row.parentBlockId ?? null, container_slot: row.containerSlot, position: row.position, kind: row.kind, shared_data: canonicalStorageData(row, row.sharedData), locales: [...row.locales].sort((left, right) => left.locale.localeCompare(right.locale)).map((locale) => ({ locale: locale.locale, data: canonicalStorageData(row, locale.data) })) })); return new TextEncoder().encode(JSON.stringify({ profile, blocks })); }
async function sha256Hex(value: Uint8Array): Promise<string> { const digest = await globalThis.crypto.subtle.digest("SHA-256", value as BufferSource); return [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, "0")).join(""); }
export function canonicalRichTextDocumentBytes(document: RichTextDocument): Uint8Array { return canonicalStorageDocumentBytes(profileName(document.profile), flattenRichTextDocumentStorage(document, ContentValidationMode.RESTORE_SNAPSHOT)); }
export function canonicalPageDocumentBytes(document: PageDocument): Uint8Array { return canonicalStorageDocumentBytes("page", flattenPageDocumentStorage(document, ContentValidationMode.RESTORE_SNAPSHOT)); }
export function canonicalRichTextDocumentHash(document: RichTextDocument): Promise<string> { return sha256Hex(canonicalRichTextDocumentBytes(document)); }
export function canonicalPageDocumentHash(document: PageDocument): Promise<string> { return sha256Hex(canonicalPageDocumentBytes(document)); }
export function canonicalLocalizedRichTextDocumentBytes(document: LocalizedRichTextDocument): Uint8Array { validateLocalizedRichTextDocument(document); return toBinary(LocalizedRichTextDocumentSchema, document); }
export function canonicalLocalizedPageDocumentBytes(document: LocalizedPageDocument): Uint8Array { validateLocalizedPageDocument(document); return toBinary(LocalizedPageDocumentSchema, document); }

export function richTextMutationTouchesSource(batch: RichTextBlockMutationBatch, sourceLocale: string): boolean { return batch.baseMutations.some((mutation) => mutation.operation.case === "delete" || mutation.operation.case === "move" || (mutation.operation.case === "upsert" && mutation.operation.value.node?.placement !== undefined)) || batch.localeMutationGroups.some((group) => group.locale === sourceLocale && group.mutations.length > 0); }
export function pageMutationTouchesSource(batch: PageSectionMutationBatch, sourceLocale: string): boolean { return batch.baseMutations.some((mutation) => mutation.operation.case === "delete" || mutation.operation.case === "move" || (mutation.operation.case === "upsert" && mutation.operation.value.node?.placement !== undefined)) || batch.localeMutationGroups.some((group) => group.locale === sourceLocale && group.mutations.length > 0); }
`;
}
