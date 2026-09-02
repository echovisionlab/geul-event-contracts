package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestEditorUploadContractsCarryNoDocumentOrBlockTarget(t *testing.T) {
	descriptors := []protoreflect.MessageDescriptor{
		(&managev1.FileIngestIdentity{}).ProtoReflect().Descriptor(),
		(&managev1.InitiateMultipartUploadRequest{}).ProtoReflect().Descriptor(),
		(&managev1.InitiateMultipartUploadResponse{}).ProtoReflect().Descriptor(),
		(&managev1.FindMultipartUploadCandidateRequest{}).ProtoReflect().Descriptor(),
		(&managev1.FindMultipartUploadCandidateResponse{}).ProtoReflect().Descriptor(),
		(&managev1.DownloadFromUrlRequest{}).ProtoReflect().Descriptor(),
		(&managev1.DownloadFromUrlResponse{}).ProtoReflect().Descriptor(),
	}
	for _, descriptor := range descriptors {
		if field := descriptor.Fields().ByName("block_id"); field != nil {
			t.Errorf("%s still exposes block_id", descriptor.FullName())
		}
	}
}

func TestTrackAttachmentRoundTripsCASWithoutEditorTarget(t *testing.T) {
	input := &managev1.FileIngestAttachedEvent{
		CorrelationId: "correlation-track",
		Identity: &managev1.FileIngestIdentity{
			EntityType:            managev1.TranscodeEntityType_TRANSCODE_ENTITY_TYPE_TRACK,
			EntityId:              "track-1",
			FileId:                "file-new",
			MediaKind:             managev1.FileIngestMediaKind_FILE_INGEST_MEDIA_KIND_TRACK_AUDIO,
			SlotId:                proto.String("original"),
			AttemptId:             proto.String("attempt-1"),
			ExpectedCurrentFileId: proto.String("file-current"),
		},
		SequenceNumber: 1,
		TimestampMs:    1_782_864_000_000,
		FileName:       "track.wav",
		MimeType:       "audio/wav",
		FileSize:       1024,
	}

	data, err := proto.Marshal(input)
	if err != nil {
		t.Fatalf("marshal FileIngestAttachedEvent: %v", err)
	}
	output := &managev1.FileIngestAttachedEvent{}
	if err := proto.Unmarshal(data, output); err != nil {
		t.Fatalf("unmarshal FileIngestAttachedEvent: %v", err)
	}
	if !proto.Equal(output, input) {
		t.Fatal("Track FileIngestAttachedEvent did not round-trip exactly")
	}
}

func TestFileScopedProcessingEntityTypeExists(t *testing.T) {
	if got := managev1.TranscodeEntityType_TRANSCODE_ENTITY_TYPE_FILE.Number(); got != 6 {
		t.Fatalf("File transcode entity number = %d, want 6", got)
	}
}

func TestIndependentMultipartCandidateReturnsAuthoritativeParts(t *testing.T) {
	response := (&managev1.FindMultipartUploadCandidateResponse{}).ProtoReflect().Descriptor()
	parts := response.Fields().ByName("uploaded_parts")
	if parts == nil {
		t.Fatal("FindMultipartUploadCandidateResponse.uploaded_parts is missing")
	}
	if parts.Number() != 7 || !parts.IsList() || parts.Kind() != protoreflect.MessageKind ||
		parts.Message().FullName() != "api.manage.v1.UploadPartInfo" {
		t.Fatalf("uploaded_parts has unexpected descriptor: %v", parts)
	}
}
