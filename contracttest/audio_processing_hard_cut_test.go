package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAutomaticAudioTranscodeJobShape(t *testing.T) {
	job := (&managev1.TranscodeAudioEvent{}).ProtoReflect().Descriptor()
	for field, number := range map[protoreflect.Name]protoreflect.FieldNumber{
		"event_id":           1,
		"entity_type":        2,
		"entity_id":          3,
		"file_id":            4,
		"source":             5,
		"hls_output":         6,
		"spectrogram_output": 7,
	} {
		descriptor := job.Fields().ByName(field)
		if descriptor == nil {
			t.Errorf("api.manage.v1.TranscodeAudioEvent.%s must remain", field)
			continue
		}
		if descriptor.Number() != number {
			t.Errorf(
				"api.manage.v1.TranscodeAudioEvent.%s number = %d, want %d",
				field,
				descriptor.Number(),
				number,
			)
		}
	}
}

func TestGenericTranscodeLifecycleContractsRemain(t *testing.T) {
	events := managev1.File_api_manage_v1_events_proto
	for _, name := range []protoreflect.Name{
		"TranscodeCompleteEvent",
		"TranscodeProgressEvent",
		"TranscodeCancelEvent",
	} {
		if events.Messages().ByName(name) == nil {
			t.Errorf("api.manage.v1.%s must remain", name)
		}
	}

	if got := managev1.TranscodeEventType_TRANSCODE_EVENT_TYPE_AUDIO; got != 1 {
		t.Errorf("TRANSCODE_EVENT_TYPE_AUDIO = %d, want 1", got)
	}
}
