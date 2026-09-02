import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  assertFormLocaleStableID,
  formFieldBlockHandle,
  formFieldCheckboxLabelTarget,
  formOptionBlockHandle,
  formOptionLabelTarget,
  formRootBlockHandle,
  formRootTitleTarget,
  formStepBlockHandle,
  formStepDescriptionTarget,
  formValidatorBlockHandle,
  formValidatorMessageTarget,
} from "@echovisionlab/geul-proto/intra/form_locale_catalog.ts";

interface FormLocaleParityFixture {
  readonly valid: Record<string, string>;
  readonly ids: Record<string, string>;
  readonly invalidIds: readonly string[];
}

const fixture = JSON.parse(
  readFileSync(
    new URL(
      "../../proto/gen/api/intra/v1/form_locale_catalog_parity.json",
      import.meta.url,
    ),
    "utf8",
  ),
) as FormLocaleParityFixture;

describe("generated Form locale catalog", () => {
  it("matches the shared cross-language block-handle fixtures", () => {
    expect({
      root: formRootBlockHandle,
      step: formStepBlockHandle(fixture.ids.step),
      field: formFieldBlockHandle(fixture.ids.field),
      option: formOptionBlockHandle(fixture.ids.option),
      validator: formValidatorBlockHandle(fixture.ids.validator),
    }).toEqual(fixture.valid);
  });

  it("rejects non-canonical and all-digit stable IDs", () => {
    for (const id of fixture.invalidIds) {
      expect(() => assertFormLocaleStableID(id)).toThrow();
    }
  });

  it("constructs only block-owned leaf targets without paths", () => {
    const targets = [
      formRootTitleTarget(),
      formStepDescriptionTarget(fixture.ids.step),
      formFieldCheckboxLabelTarget(fixture.ids.field),
      formOptionLabelTarget(fixture.ids.option),
      formValidatorMessageTarget(fixture.ids.validator),
    ];
    for (const target of targets) {
      expect(target.owner.case).toBe("blockHandle");
      expect(target.owner.value).toMatch(
        /^(document|form:(step|field|option|validator):)/,
      );
      expect(target.fieldHandle).not.toBe("");
      expect(target.path).toEqual([]);
    }
  });
});
