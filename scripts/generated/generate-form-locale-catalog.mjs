import fs from "node:fs";
import path from "node:path";
import process from "node:process";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import YAML from "yaml";

const root = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const sourcePath = path.join(root, "config/content/form-locale-catalog.yaml");
const goPath = path.join(root, "gen/api/intra/v1/form_locale_catalog.go");
const tsPath = path.join(
  root,
  "packages/proto/gen/api/intra/v1/form_locale_catalog.ts",
);
const fixturePath = path.join(
  root,
  "packages/proto/gen/api/intra/v1/form_locale_catalog_parity.json",
);
const check = process.argv.includes("--check");
const kinds = ["root", "step", "field", "option", "validator"];

const catalog = YAML.parse(fs.readFileSync(sourcePath, "utf8"));
validateCatalog(catalog);

const fixtureIDs = {
  step: "step_A-1",
  field: "field_b-2",
  option: "option_C-3",
  validator: "validator_d-4",
};
const fixtures = {
  schemaVersion: catalog.schema_version,
  valid: Object.fromEntries(
    kinds.map((kind) => [
      kind,
      expectedBlockHandle(catalog.blocks[kind], fixtureIDs[kind]),
    ]),
  ),
  ids: fixtureIDs,
  invalidIds: ["", "0", "123", "-leading", "white space", "a".repeat(97)],
};

const go = execFileSync("gofmt", [], {
  input: generateGo(catalog),
  encoding: "utf8",
});
const ts = execFileSync(
  "pnpm",
  ["exec", "prettier", "--stdin-filepath", tsPath],
  { input: generateTypeScript(catalog), encoding: "utf8", cwd: root },
);
const fixture = `${JSON.stringify(fixtures, null, 2)}\n`;

writeOrCheck(goPath, go, "Go Form locale catalog");
writeOrCheck(tsPath, ts, "TypeScript Form locale catalog");
writeOrCheck(fixturePath, fixture, "Form locale parity fixture");

function validateCatalog(value) {
  if (value?.schema_version !== 1)
    throw new Error("Form locale catalog schema_version must be 1");
  if (value?.stable_id?.pattern !== "^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$") {
    throw new Error("Form locale catalog stable ID grammar drifted");
  }
  if (value?.stable_id?.reject_all_digits !== true) {
    throw new Error("Form locale catalog must reject all-digit stable IDs");
  }
  const expectedFields = {
    root: ["title"],
    step: ["title", "description"],
    field: ["label", "description", "placeholder", "checkbox_label"],
    option: ["label"],
    validator: ["message"],
  };
  for (const kind of kinds) {
    const block = value.blocks?.[kind];
    if (!block) throw new Error(`Form locale catalog is missing ${kind}`);
    if (
      JSON.stringify(block.locale_fields) !==
      JSON.stringify(expectedFields[kind])
    ) {
      throw new Error(`Form locale catalog ${kind} fields drifted`);
    }
    if (kind === "root") {
      if (
        block.fixed_handle !== "document" ||
        block.handle_prefix !== undefined
      ) {
        throw new Error("Form root handle must remain document");
      }
    } else if (
      block.handle_prefix !== `form:${kind}:` ||
      block.stable_id_required !== true ||
      block.fixed_handle !== undefined
    ) {
      throw new Error(`Form ${kind} handle contract drifted`);
    }
  }
}

function expectedBlockHandle(block, id) {
  return block.fixed_handle ?? `${block.handle_prefix}${id}`;
}

function pascal(value) {
  return value
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join("");
}

function camel(value) {
  const name = pascal(value);
  return name.charAt(0).toLowerCase() + name.slice(1);
}

function generateGo(value) {
  const constants = [];
  const blockHelpers = [];
  const targetHelpers = [];
  for (const kind of kinds) {
    const block = value.blocks[kind];
    const kindName = pascal(kind);
    if (block.fixed_handle) {
      constants.push(
        `Form${kindName}BlockHandle = ${JSON.stringify(block.fixed_handle)}`,
      );
    } else {
      constants.push(
        `Form${kindName}BlockHandlePrefix = ${JSON.stringify(block.handle_prefix)}`,
      );
      blockHelpers.push(`func Form${kindName}BlockHandle(id string) (string, error) {
\tif err := ValidateFormLocaleStableID(id); err != nil { return "", err }
\treturn Form${kindName}BlockHandlePrefix + id, nil
}`);
    }
    for (const field of block.locale_fields) {
      constants.push(
        `Form${kindName}${pascal(field)}FieldHandle = ${JSON.stringify(field)}`,
      );
      const blockExpr = block.fixed_handle
        ? `Form${kindName}BlockHandle`
        : `blockHandle, err := Form${kindName}BlockHandle(id)\n\tif err != nil { return nil, err }`;
      const signature = block.fixed_handle
        ? "() *managev1.AIDocumentFieldTarget"
        : "(id string) (*managev1.AIDocumentFieldTarget, error)";
      const returns = block.fixed_handle
        ? `return formLocaleFieldTarget(${blockExpr}, Form${kindName}${pascal(field)}FieldHandle)`
        : `${blockExpr}\n\treturn formLocaleFieldTarget(blockHandle, Form${kindName}${pascal(field)}FieldHandle), nil`;
      targetHelpers.push(
        `func Form${kindName}${pascal(field)}Target${signature} {\n\t${returns}\n}`,
      );
    }
  }
  return `// Code generated by scripts/generated/generate-form-locale-catalog.mjs. DO NOT EDIT.

package intrav1

import (
\t"errors"
\t"regexp"

\tmanagev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
)

const (
\t${constants.join("\n\t")}
)

var formLocaleStableIDPattern = regexp.MustCompile(${JSON.stringify(value.stable_id.pattern)})
var formLocaleAllDigitsPattern = regexp.MustCompile(${JSON.stringify("^[0-9]+$")})

func ValidateFormLocaleStableID(id string) error {
\tif !formLocaleStableIDPattern.MatchString(id) || formLocaleAllDigitsPattern.MatchString(id) {
\t\treturn errors.New("form locale stable ID must match ${value.stable_id.pattern} and contain a non-digit")
\t}
\treturn nil
}

${blockHelpers.join("\n\n")}

func formLocaleFieldTarget(blockHandle, fieldHandle string) *managev1.AIDocumentFieldTarget {
\treturn &managev1.AIDocumentFieldTarget{
\t\tOwner: &managev1.AIDocumentFieldTarget_BlockHandle{BlockHandle: blockHandle},
\t\tFieldHandle: fieldHandle,
\t}
}

${targetHelpers.join("\n\n")}
`;
}

function generateTypeScript(value) {
  const constants = [];
  const blockHelpers = [];
  const targetHelpers = [];
  for (const kind of kinds) {
    const block = value.blocks[kind];
    const kindName = pascal(kind);
    const kindCamel = camel(kind);
    if (block.fixed_handle) {
      constants.push(
        `export const form${kindName}BlockHandle = ${JSON.stringify(block.fixed_handle)} as const;`,
      );
    } else {
      constants.push(
        `export const form${kindName}BlockHandlePrefix = ${JSON.stringify(block.handle_prefix)} as const;`,
      );
      blockHelpers.push(`export function form${kindName}BlockHandle(id: string): string {
  assertFormLocaleStableID(id);
  return form${kindName}BlockHandlePrefix + id;
}`);
    }
    for (const field of block.locale_fields) {
      constants.push(
        `export const form${kindName}${pascal(field)}FieldHandle = ${JSON.stringify(field)} as const;`,
      );
      const signature = block.fixed_handle ? "()" : "(id: string)";
      const blockExpr = block.fixed_handle
        ? `form${kindName}BlockHandle`
        : `form${kindName}BlockHandle(id)`;
      targetHelpers.push(`export function form${kindName}${pascal(field)}Target${signature}: AIDocumentFieldTarget {
  return formLocaleFieldTarget(${blockExpr}, form${kindName}${pascal(field)}FieldHandle);
}`);
    }
    void kindCamel;
  }
  return `// Code generated by scripts/generated/generate-form-locale-catalog.mjs. DO NOT EDIT.

import { create } from "@bufbuild/protobuf";
import type { AIDocumentFieldTarget } from "../../manage/v1/ai_pb.ts";
import { AIDocumentFieldTargetSchema } from "../../manage/v1/ai_pb.ts";

${constants.join("\n")}

const formLocaleStableIDPattern = new RegExp(${JSON.stringify(value.stable_id.pattern)});
const formLocaleAllDigitsPattern = /^[0-9]+$/;

export function assertFormLocaleStableID(id: string): void {
  if (!formLocaleStableIDPattern.test(id) || formLocaleAllDigitsPattern.test(id)) {
    throw new Error("form locale stable ID must match ${value.stable_id.pattern} and contain a non-digit");
  }
}

${blockHelpers.join("\n\n")}

function formLocaleFieldTarget(blockHandle: string, fieldHandle: string): AIDocumentFieldTarget {
  return create(AIDocumentFieldTargetSchema, {
    owner: { case: "blockHandle", value: blockHandle },
    fieldHandle,
  });
}

${targetHelpers.join("\n\n")}
`;
}

function writeOrCheck(file, content, label) {
  if (check) {
    const current = fs.existsSync(file) ? fs.readFileSync(file, "utf8") : "";
    if (current !== content) {
      throw new Error(
        `generated ${label} is stale: ${path.relative(root, file)}`,
      );
    }
    return;
  }
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, content);
}
