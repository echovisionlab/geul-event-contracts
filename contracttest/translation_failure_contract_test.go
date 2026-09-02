package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationFailuresExposeOnlyBoundedReasons(t *testing.T) {
	job := (&managev1.TranslationJob{}).ProtoReflect().Descriptor()
	requireNoField(t, job, "error")
	requireNoField(t, job, "failure_reason")

	lifecycle := (&managev1.TranslationLifecycleEvent{}).ProtoReflect().Descriptor()
	requireNoField(t, lifecycle, "error")
	requireTranslationFailureReasonField(t, lifecycle, "failure_reason", 7)

	enum := managev1.TranslationFailureReason(0).Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{"TRANSLATION_FAILURE_REASON_UNSPECIFIED", 0},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_CONFIGURATION", 1},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_AUTHENTICATION", 2},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_RATE_LIMITED", 3},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_UNAVAILABLE", 4},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_REJECTED", 5},
		{"TRANSLATION_FAILURE_REASON_PROVIDER_RESPONSE_INVALID", 6},
		{"TRANSLATION_FAILURE_REASON_TARGET_APPLY_FAILED", 7},
		{"TRANSLATION_FAILURE_REASON_OG_HANDOFF_FAILED", 8},
		{"TRANSLATION_FAILURE_REASON_INTERNAL", 9},
	}
	if enum.Values().Len() != len(want) {
		t.Fatalf("TranslationFailureReason values = %d, want %d", enum.Values().Len(), len(want))
	}
	for i, expected := range want {
		value := enum.Values().Get(i)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Errorf("TranslationFailureReason[%d] = %s/%d, want %s/%d", i, value.Name(), value.Number(), expected.name, expected.number)
		}
	}
	if enum.Values().ByName("TRANSLATION_FAILURE_REASON_SOURCE_NO_LONGER_CURRENT") != nil {
		t.Fatal("source-currentness must not remain a translation failure reason")
	}
}

func requireTranslationFailureReasonField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireField(t, message, name, number, protoreflect.EnumKind)
	if got, want := field.Enum().FullName(), protoreflect.FullName("api.manage.v1.TranslationFailureReason"); got != want {
		t.Errorf("%s.%s enum = %s, want %s", message.FullName(), name, got, want)
	}
}
