package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestPageCollaborationUsesOneTypedAggregateDocument(t *testing.T) {
	service := intrav1.File_api_intra_v1_page_proto.Services().ByName("InternalPageService")
	if service == nil {
		t.Fatal("InternalPageService descriptor is missing")
	}

	loadRequest := (&intrav1.LoadPageBlockDocumentRequest{}).ProtoReflect().Descriptor()
	requireField(t, loadRequest, "page_id", 1, protoreflect.StringKind)
	principal := requireField(t, loadRequest, "principal", 2, protoreflect.MessageKind)
	if got, want := principal.Message().FullName(), protoreflect.FullName("api.intra.v1.CollaborationPrincipal"); got != want {
		t.Fatalf("%s.principal type = %s, want %s", loadRequest.FullName(), got, want)
	}
	requireField(t, loadRequest, "locale", 3, protoreflect.StringKind)
	for _, method := range []protoreflect.Name{
		"LoadPageBlockDocument",
		"ApplyPageBlockBatch",
		"UpdatePageLocaleMetadata",
		"UpdatePageDocumentMetadata",
		"CreatePageVersionCheckpoint",
	} {
		if service.Methods().ByName(method) == nil {
			t.Errorf("InternalPageService must expose %s", method)
		}
	}

	load := (&intrav1.LoadPageBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requirePageMessageType(t, load, "document", 1, "api.content.v1.LocalizedPageDocument")
	requireField(t, load, "document_revision", 2, protoreflect.StringKind)
	metadata := requireField(t, load, "source_metadata", 5, protoreflect.MessageKind)
	if metadata.Cardinality() == protoreflect.Repeated {
		t.Error("LoadPageBlockDocumentResponse.source_metadata must be singular")
	}
	requireDocumentLayoutField(t, load, 3)
	requirePageBlockMediaField(t, load, 4)
	requireField(t, load, "locale", 6, protoreflect.StringKind)
	requireField(t, load, "locale_exists", 7, protoreflect.BoolKind)
	requireField(t, load, "locale_metadata", 8, protoreflect.MessageKind)
	targetRevision := requireField(t, load, "target_revision", 9, protoreflect.StringKind)
	if !targetRevision.HasOptionalKeyword() {
		t.Fatal("Page target_revision must be optional")
	}
	requireNoField(t, load, "source_epoch")
	requireNoField(t, load, "snapshot_digest")

	apply := (&intrav1.ApplyPageBlockBatchRequest{}).ProtoReflect().Descriptor()
	requireField(t, apply, "page_id", 1, protoreflect.StringKind)
	requirePageMessageType(t, apply, "batch", 2, "api.content.v1.PageSectionMutationBatch")
	requireField(t, apply, "locale", 3, protoreflect.StringKind)
}

func TestPageMetadataAndCheckpointUseSeparateRevisionCAS(t *testing.T) {
	document := (&intrav1.UpdatePageDocumentMetadataRequest{}).ProtoReflect().Descriptor()
	requireField(t, document, "page_id", 1, protoreflect.StringKind)
	requireField(t, document, "expected_revision", 2, protoreflect.StringKind)
	requireDocumentLayoutField(t, document, 3)
	requireContributorIDsField(t, document, 4)
	requireField(t, document, "locale", 5, protoreflect.StringKind)

	source := (&intrav1.UpdatePageLocaleMetadataRequest{}).ProtoReflect().Descriptor()
	requireField(t, source, "page_id", 1, protoreflect.StringKind)
	requireField(t, source, "title", 2, protoreflect.StringKind)
	requireField(t, source, "set_summary", 3, protoreflect.StringKind)
	requireField(t, source, "clear_summary", 4, protoreflect.BoolKind)
	requireField(t, source, "expected_revision", 5, protoreflect.StringKind)
	requireContributorIDsField(t, source, 6)
	requireField(t, source, "locale", 7, protoreflect.StringKind)
	requireField(t, source, "expected_target_revision", 8, protoreflect.StringKind)
	requireNoField(t, source, "patch_mask")
	requireNoField(t, source, "expected_source_revision")

	checkpoint := (&intrav1.CreatePageVersionCheckpointRequest{}).ProtoReflect().Descriptor()
	requireField(t, checkpoint, "page_id", 1, protoreflect.StringKind)
	requireField(t, checkpoint, "expected_revision", 2, protoreflect.StringKind)
	requireContributorIDsField(t, checkpoint, 3)
	requireField(t, checkpoint, "locale", 4, protoreflect.StringKind)
	for _, field := range []protoreflect.Name{"version_checkpoint", "yjs_state", "content_json"} {
		requireNoField(t, checkpoint, field)
	}
}

func TestPageManageAndOpenUseTypedDocuments(t *testing.T) {
	managePage := (&managev1.Page{}).ProtoReflect().Descriptor()
	requirePageMessageType(t, managePage, "document", 4, "api.content.v1.PageDocument")
	requireField(t, managePage, "revision", 5, protoreflect.StringKind)
	requireNoField(t, managePage, "source_epoch")
	requireNoField(t, managePage, "canonical_hash")
	requireDocumentLayoutField(t, managePage, 13)
	requirePageBlockMediaField(t, managePage, 15)

	update := (&managev1.UpdatePageRequest{}).ProtoReflect().Descriptor()
	requireField(t, update, "slug", 2, protoreflect.StringKind)
	requireField(t, update, "show_title", 3, protoreflect.BoolKind)
	requireNoField(t, update, "document_layout")

	openPage := (&openv1.Page{}).ProtoReflect().Descriptor()
	requirePageMessageType(t, openPage, "document", 9, "api.content.v1.LocalizedPageDocument")
	requireField(t, openPage, "revision", 13, protoreflect.StringKind)
	requireDocumentLayoutField(t, openPage, 14)
	requireNoField(t, openPage, "canonical_hash")
	requireNoField(t, openPage, "source_epoch")
	requireNoField(t, openPage, "content_json")

	response := (&openv1.GetPageResponse{}).ProtoReflect().Descriptor()
	requireNoField(t, response, "render_status")
	requirePageBlockMediaField(t, response, 2)
	requireNoField(t, response, "media")
	requireNoField(t, response, "media_deliveries")
	requireNoField(t, response, "download_access")
}

func TestPageVersionExposesOnlyTypedSnapshotMetadata(t *testing.T) {
	descriptor := (&managev1.PageVersion{}).ProtoReflect().Descriptor()
	for _, field := range []protoreflect.Name{
		"slug", "show_title", "document_layout", "status", "featured_image_asset", "content_hash",
	} {
		requireNoField(t, descriptor, field)
	}
	title := requireField(t, descriptor, "title", 3, protoreflect.StringKind)
	if !title.HasOptionalKeyword() {
		t.Error("PageVersion.title must preserve the nullable typed snapshot title")
	}
	requireField(t, descriptor, "canonical_hash", 4, protoreflect.StringKind)
	requireField(t, descriptor, "summary", 6, protoreflect.StringKind)
}

func TestPageFeaturedImageStillUsesSignedMediaDelivery(t *testing.T) {
	page := (&managev1.Page{}).ProtoReflect().Descriptor()
	requireField(t, page, "featured_image_delivery", 14, protoreflect.MessageKind)

	response := (&managev1.SetPageFeaturedImageResponse{}).ProtoReflect().Descriptor()
	requireField(t, response, "image_delivery", 2, protoreflect.MessageKind)
}

func requirePageMessageType(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
	want protoreflect.FullName,
) {
	t.Helper()
	field := requireField(t, descriptor, name, number, protoreflect.MessageKind)
	if got := field.Message().FullName(); got != want {
		t.Fatalf("%s.%s type = %s, want %s", descriptor.FullName(), name, got, want)
	}
}

func requirePageBlockMediaField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	number protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireField(t, descriptor, "block_media", number, protoreflect.MessageKind)
	if field.Cardinality() != protoreflect.Repeated || field.Message().FullName() != "api.content.v1.ContentBlockMediaItem" {
		t.Fatalf("%s.block_media must be repeated api.content.v1.ContentBlockMediaItem", descriptor.FullName())
	}
}
