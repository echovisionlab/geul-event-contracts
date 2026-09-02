package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTrackOriginalAudioProjectionIsTypedAPIOwnedCAS(t *testing.T) {
	service := intrav1.File_api_intra_v1_file_ingest_proto.Services().ByName("InternalFileIngestService")
	method := service.Methods().ByName("AttachTrackOriginalAudio")
	if method == nil {
		t.Fatal("AttachTrackOriginalAudio RPC is missing")
	}

	request := (&intrav1.AttachTrackOriginalAudioRequest{}).ProtoReflect().Descriptor()
	assertExactFields(t, request, []fieldExpectation{
		{name: "track_id", number: 1, kind: protoreflect.StringKind},
		{name: "verified_file_id", number: 2, kind: protoreflect.StringKind},
		{name: "ingest_attempt_id", number: 3, kind: protoreflect.StringKind},
		{name: "expected_current_file_id", number: 4, kind: protoreflect.StringKind, optional: true},
	})

	response := (&intrav1.AttachTrackOriginalAudioResponse{}).ProtoReflect().Descriptor()
	assertExactFields(t, response, []fieldExpectation{
		{name: "result", number: 1, kind: protoreflect.EnumKind},
		{name: "current_file_id", number: 2, kind: protoreflect.StringKind},
		{name: "release_id", number: 3, kind: protoreflect.StringKind},
	})

	if got := intrav1.AttachTrackOriginalAudioResult_ATTACH_TRACK_ORIGINAL_AUDIO_RESULT_APPLIED.Number(); got != 1 {
		t.Fatalf("APPLIED enum number = %d, want 1", got)
	}
	if got := intrav1.AttachTrackOriginalAudioResult_ATTACH_TRACK_ORIGINAL_AUDIO_RESULT_ALREADY_APPLIED.Number(); got != 2 {
		t.Fatalf("ALREADY_APPLIED enum number = %d, want 2", got)
	}
}
