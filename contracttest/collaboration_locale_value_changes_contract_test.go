package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestResidentCollaborationApplyCarriesExactAffectedLocaleValues(t *testing.T) {
	contracts := []struct {
		path    string
		message string
	}{
		{"api/intra/v1/post.proto", "ApplyPostBlockBatchRequest"},
		{"api/intra/v1/page.proto", "ApplyPageBlockBatchRequest"},
		{"api/intra/v1/work.proto", "ApplyWorkBlockBatchRequest"},
		{"api/intra/v1/program_event.proto", "ApplyProgramEventBlockBatchRequest"},
		{"api/intra/v1/artist.proto", "ApplyArtistBlockBatchRequest"},
		{"api/intra/v1/label.proto", "ApplyLabelBlockBatchRequest"},
		{"api/intra/v1/release.proto", "ApplyReleaseBlockBatchRequest"},
		{"api/intra/v1/campaign.proto", "ApplyCampaignBlockBatchRequest"},
		{"api/intra/v1/email_template.proto", "ApplyEmailTemplateBlockBatchRequest"},
		{"api/intra/v1/privacy.proto", "ApplyPrivacyBlockBatchRequest"},
		{"api/intra/v1/terms.proto", "ApplyTermsBlockBatchRequest"},
	}

	for _, contract := range contracts {
		source := readSourceProto(t, contract.path)
		declaration := sourceDeclaration(t, source, "message", contract.message)
		requireSourceContains(t, declaration,
			"repeated api.manage.v1.AIDocumentFieldTarget affected_locale_values = 5;",
			"Actual changed locale-owned leaf terminals",
			"explicit-empty writes and presence removal",
			"canonical-sorted,",
			"duplicate-free targets",
			"relation_item, shared/topology/File",
			"identity-terminal (row_id, cell_id, Page unit_id)",
			"composite object/list",
			"Source/shared-only batches send an empty list",
		)
	}
}

func TestInteractiveMutationRelayUsesOnlyDocumentAndTargetRevisions(t *testing.T) {
	source := readSourceProto(t, "api/intra/v1/collaboration.proto")
	declaration := sourceDeclaration(t, source, "message", "RelayInteractiveAIDocumentMutationRequest")
	requireSourceContains(t, declaration,
		"string expected_document_revision = 5;",
		"string accepted_document_revision = 6;",
		"optional string expected_target_revision = 9;",
		"optional string accepted_target_revision = 10;",
	)
	requireSourceExcludes(t, declaration, "source_epoch", "source_hash", "snapshot_digest")
}

func TestCustomDocumentSavesDoNotAdvertiseUnavailableAffectedLocaleValues(t *testing.T) {
	for _, descriptor := range []struct {
		name   string
		fields protoreflect.FieldDescriptors
	}{
		{"Form", (&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor().Fields()},
		{"EmailLayout", (&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor().Fields()},
	} {
		if field := descriptor.fields.ByName("affected_locale_values"); field != nil {
			t.Fatalf("%s custom save unexpectedly exposes affected_locale_values", descriptor.name)
		}
	}
}

func TestAffectedLocaleValuesRoundTripExplicitEmptyAndTableCellLeaves(t *testing.T) {
	explicitEmpty := blockFieldTarget("paragraph-1", "content")
	tableCell := &managev1.AIDocumentFieldTarget{
		Owner:       &managev1.AIDocumentFieldTarget_BlockHandle{BlockHandle: "table-1"},
		FieldHandle: "tableContent",
		Path: []*managev1.AIDocumentFieldPathSegment{
			fieldPath("rows"),
			itemPath("row-1"),
			fieldPath("cells"),
			itemPath("cell-1"),
			fieldPath("content"),
		},
	}
	want := &intrav1.ApplyPageBlockBatchRequest{
		PageId:               "page-1",
		Locale:               "en",
		AffectedLocaleValues: []*managev1.AIDocumentFieldTarget{explicitEmpty, tableCell},
	}

	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("marshal affected locale values: %v", err)
	}
	got := new(intrav1.ApplyPageBlockBatchRequest)
	if err := proto.Unmarshal(wire, got); err != nil {
		t.Fatalf("unmarshal affected locale values: %v", err)
	}
	if !proto.Equal(want, got) {
		t.Fatalf("affected locale values changed across round trip:\nwant %v\n got %v", want, got)
	}
}

func blockFieldTarget(blockHandle, fieldHandle string) *managev1.AIDocumentFieldTarget {
	return &managev1.AIDocumentFieldTarget{
		Owner:       &managev1.AIDocumentFieldTarget_BlockHandle{BlockHandle: blockHandle},
		FieldHandle: fieldHandle,
	}
}

func fieldPath(handle string) *managev1.AIDocumentFieldPathSegment {
	return &managev1.AIDocumentFieldPathSegment{
		Selector: &managev1.AIDocumentFieldPathSegment_FieldHandle{FieldHandle: handle},
	}
}

func itemPath(handle string) *managev1.AIDocumentFieldPathSegment {
	return &managev1.AIDocumentFieldPathSegment{
		Selector: &managev1.AIDocumentFieldPathSegment_ItemHandle{ItemHandle: handle},
	}
}
