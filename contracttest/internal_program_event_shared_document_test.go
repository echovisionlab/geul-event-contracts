package contracttest_test

import (
	"testing"

	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestInternalProgramEventServiceOwnsSharedDocument(t *testing.T) {
	service := intrav1.File_api_intra_v1_program_event_proto.Services().ByName("InternalProgramEventService")
	if service == nil {
		t.Fatal("api.intra.v1.InternalProgramEventService descriptor is missing")
	}

	for _, name := range []protoreflect.Name{"ApplyProgramEventBlockBatch", "LoadProgramEventBlockDocument"} {
		if service.Methods().ByName(name) == nil {
			t.Errorf("InternalProgramEventService.%s is missing", name)
		}
	}

	apply := (&intrav1.ApplyProgramEventBlockBatchRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, apply, 5)
	requireMessageField(t, apply, "event_id", 1, protoreflect.StringKind, "")
	batch := requireMessageField(t, apply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch").Message()
	requireMessageField(t, apply, "locale", 3, protoreflect.StringKind, "")
	requireMessageField(t, apply, "expected_target_revision", 4, protoreflect.StringKind, "")
	requireMessageField(t, apply, "affected_locale_values", 5, protoreflect.MessageKind, "api.manage.v1.AIDocumentFieldTarget")
	requireMessageField(t, batch, "expected_revision", 3, protoreflect.StringKind, "")
	contributors := requireMessageField(t, batch, "contributor_member_ids", 6, protoreflect.StringKind, "")
	if contributors.Cardinality() != protoreflect.Repeated {
		t.Error("RichTextBlockMutationBatch.contributor_member_ids must be repeated")
	}
	load := (&intrav1.LoadProgramEventBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, load, "document", 1, protoreflect.MessageKind, "api.content.v1.LocalizedRichTextDocument")
	requireMessageField(t, load, "document_revision", 2, protoreflect.StringKind, "")
	requireMessageField(t, load, "locale", 5, protoreflect.StringKind, "")
	requireMessageField(t, load, "locale_exists", 6, protoreflect.BoolKind, "")
	requireMessageField(t, load, "locale_metadata", 7, protoreflect.MessageKind, "api.intra.v1.ProgramEventLocaleMetadata")
	requireMessageField(t, load, "target_revision", 8, protoreflect.StringKind, "")
	if load.Fields().ByName("source_epoch") != nil || load.Fields().ByName("snapshot_digest") != nil {
		t.Fatal("Program Event collaboration load must not duplicate source identity or snapshot digest")
	}

	metadata := requireMessageField(t, load, "source_metadata", 4, protoreflect.MessageKind, "api.intra.v1.ProgramEventLocaleMetadata").Message()
	requireMessageFieldCount(t, metadata, 3)
	requireMessageField(t, metadata, "locale", 1, protoreflect.StringKind, "")
	requireMessageField(t, metadata, "title", 2, protoreflect.StringKind, "")
	requireMessageField(t, metadata, "summary", 3, protoreflect.StringKind, "")

	assertProgramEventSourceMetadataReservations(t, metadata)

	manageLocale := (&managev1.ProgramEventLocale{}).ProtoReflect().Descriptor()
	requireMessageField(t, manageLocale, "content_html", 3, protoreflect.StringKind, "")
	requireMessageField(t, manageLocale, "content_text", 4, protoreflect.StringKind, "")
	openEvent := (&openv1.ProgramEvent{}).ProtoReflect().Descriptor()
	requireMessageField(t, openEvent, "content_html", 6, protoreflect.StringKind, "")
	requireMessageField(t, openEvent, "content_text", 7, protoreflect.StringKind, "")
}

func assertProgramEventSourceMetadataReservations(t *testing.T, metadata protoreflect.MessageDescriptor) {
	t.Helper()

	ranges := metadata.ReservedRanges()
	if ranges.Len() != 1 || ranges.Get(0)[0] != 4 || ranges.Get(0)[1] != 6 {
		t.Fatalf("ProgramEventLocaleMetadata reserved ranges = %v, want [4, 6)", ranges)
	}
	names := metadata.ReservedNames()
	if names.Len() != 2 || names.Get(0) != "content_html" || names.Get(1) != "content_text" {
		t.Fatalf("ProgramEventLocaleMetadata reserved names = %v, want content_html and content_text", names)
	}
}

func TestProgramEventSharedDocumentRoundTrip(t *testing.T) {
	source := &intrav1.ApplyProgramEventBlockBatchRequest{
		EventId: "event-1",
		Locale:  "ko",
		Batch: &contentv1.RichTextBlockMutationBatch{
			BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
			Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_PROGRAM_EVENT,
			ExpectedRevision:        "00000000-0000-4000-8000-000000000002",
			ContributorMemberIds:    []string{"00000000-0000-4000-8000-000000000001"},
		},
	}

	data, err := proto.Marshal(source)
	if err != nil {
		t.Fatalf("marshal shared Program Event document: %v", err)
	}
	decoded := &intrav1.ApplyProgramEventBlockBatchRequest{}
	if err := proto.Unmarshal(data, decoded); err != nil {
		t.Fatalf("unmarshal shared Program Event document: %v", err)
	}
	if !proto.Equal(source, decoded) {
		t.Fatal("shared Program Event document did not round-trip exactly")
	}
}
