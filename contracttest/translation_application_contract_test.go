package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationJobExposesOnlyInFlightStates(t *testing.T) {
	status := managev1.TranslationJobStatus(0).Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{"TRANSLATION_JOB_STATUS_UNSPECIFIED", 0},
		{"TRANSLATION_JOB_STATUS_QUEUED", 1},
		{"TRANSLATION_JOB_STATUS_RUNNING", 2},
	}
	if status.Values().Len() != len(want) {
		t.Fatalf("TranslationJobStatus values = %d, want %d", status.Values().Len(), len(want))
	}
	for index, expected := range want {
		value := status.Values().Get(index)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Fatalf("TranslationJobStatus[%d] = %s/%d, want %s/%d", index, value.Name(), value.Number(), expected.name, expected.number)
		}
	}
}

func TestCancelTranslationJobReturnsNoDeletedJob(t *testing.T) {
	response := (&managev1.CancelTranslationJobResponse{}).ProtoReflect().Descriptor()
	if response.Fields().Len() != 0 {
		t.Fatalf("CancelTranslationJobResponse fields = %d, want 0", response.Fields().Len())
	}
}

func TestTranslationLifecycleReportsOnlyApplicationOutcome(t *testing.T) {
	status := managev1.TranslationLifecycleStatus(0).Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{"TRANSLATION_LIFECYCLE_STATUS_UNSPECIFIED", 0},
		{"TRANSLATION_LIFECYCLE_STATUS_APPLIED", 1},
		{"TRANSLATION_LIFECYCLE_STATUS_FAILED", 2},
	}
	if status.Values().Len() != len(want) {
		t.Fatalf("TranslationLifecycleStatus values = %d, want %d", status.Values().Len(), len(want))
	}
	for index, expected := range want {
		value := status.Values().Get(index)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Fatalf("TranslationLifecycleStatus[%d] = %s/%d, want %s/%d", index, value.Name(), value.Number(), expected.name, expected.number)
		}
	}
}

func TestTranslationEntryIsExistenceAndContentOnly(t *testing.T) {
	entry := (&managev1.TranslationEntry{}).ProtoReflect().Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{"target", 1},
		{"locale", 2},
		{"title", 3},
		{"summary", 4},
		{"content_html", 5},
		{"content_text", 6},
		{"content_json", 7},
		{"updated_at", 8},
		{"og_asset", 9},
	}
	if entry.Fields().Len() != len(want) {
		t.Fatalf("TranslationEntry fields = %d, want %d", entry.Fields().Len(), len(want))
	}
	for index, expected := range want {
		field := entry.Fields().Get(index)
		if field.Name() != expected.name || field.Number() != expected.number {
			t.Fatalf("TranslationEntry[%d] = %s/%d, want %s/%d", index, field.Name(), field.Number(), expected.name, expected.number)
		}
	}
}
