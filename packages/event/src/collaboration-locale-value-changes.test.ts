import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";
import {
  AIDocumentFieldPathSegmentSchema,
  AIDocumentFieldTargetSchema,
} from "@echovisionlab/geul-proto/secure/ai_pb.ts";
import { ApplyPageBlockBatchRequestSchema } from "@echovisionlab/geul-proto/intra/page_pb.ts";

function fieldPath(fieldHandle: string) {
  return create(AIDocumentFieldPathSegmentSchema, {
    selector: { case: "fieldHandle", value: fieldHandle },
  });
}

function itemPath(itemHandle: string) {
  return create(AIDocumentFieldPathSegmentSchema, {
    selector: { case: "itemHandle", value: itemHandle },
  });
}

describe("resident collaboration affected locale values", () => {
  it("round-trips explicit-empty and exact table-cell leaf targets", () => {
    const explicitEmpty = create(AIDocumentFieldTargetSchema, {
      owner: { case: "blockHandle", value: "paragraph-1" },
      fieldHandle: "content",
    });
    const tableCell = create(AIDocumentFieldTargetSchema, {
      owner: { case: "blockHandle", value: "table-1" },
      fieldHandle: "tableContent",
      path: [
        fieldPath("rows"),
        itemPath("row-1"),
        fieldPath("cells"),
        itemPath("cell-1"),
        fieldPath("content"),
      ],
    });
    const request = create(ApplyPageBlockBatchRequestSchema, {
      pageId: "page-1",
      locale: "en",
      affectedLocaleValues: [explicitEmpty, tableCell],
    });

    const decoded = fromBinary(
      ApplyPageBlockBatchRequestSchema,
      toBinary(ApplyPageBlockBatchRequestSchema, request),
    );

    expect(decoded.affectedLocaleValues).toEqual([explicitEmpty, tableCell]);
  });
});
