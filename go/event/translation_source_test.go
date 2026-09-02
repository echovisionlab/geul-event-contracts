package event

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationJobFieldsAreApplicationFacts(t *testing.T) {
	message := (&managev1.TranslationJob{}).ProtoReflect().Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{"id", 1},
		{"target", 2},
		{"target_locale", 3},
		{"source_locale", 4},
		{"request_artifact_digest", 5},
		{"status", 6},
		{"operation_id", 7},
		{"provider", 8},
		{"model", 9},
		{"requested_at", 10},
		{"started_at", 11},
		{"requested_by_member_id", 12},
	}
	if message.Fields().Len() != len(want) {
		t.Fatalf("TranslationJob field count = %d, want %d", message.Fields().Len(), len(want))
	}
	for index, expected := range want {
		field := message.Fields().Get(index)
		if field.Name() != expected.name || field.Number() != expected.number {
			t.Fatalf("TranslationJob field %d = %s/%d, want %s/%d", index, field.Name(), field.Number(), expected.name, expected.number)
		}
	}
	for _, forbidden := range []protoreflect.Name{
		"translation_spec_version",
		"source_hash",
		"source_revision",
		"source_epoch",
		"cancel_requested",
		"completed_at",
		"duration_ms",
		"failure_reason",
		"attempts",
		"last_heartbeat_at",
		"superseded_by_revision",
		"superseded_by_source_epoch",
	} {
		if field := message.Fields().ByName(forbidden); field != nil {
			t.Fatalf("TranslationJob.%s duplicate or terminal state unexpectedly remains", forbidden)
		}
	}
}

func TestTranslationLifecycleCarriesNoSecondSourceIdentity(t *testing.T) {
	message := (&managev1.TranslationLifecycleEvent{}).ProtoReflect().Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{"job_id", 1},
		{"entity_type", 2},
		{"entity_id", 3},
		{"target_locale", 4},
		{"status", 5},
		{"timestamp_ms", 6},
		{"failure_reason", 7},
	}
	if message.Fields().Len() != len(want) {
		t.Fatalf("TranslationLifecycleEvent field count = %d, want %d", message.Fields().Len(), len(want))
	}
	for index, expected := range want {
		field := message.Fields().Get(index)
		if field.Name() != expected.name || field.Number() != expected.number {
			t.Fatalf("TranslationLifecycleEvent field %d = %s/%d, want %s/%d", index, field.Name(), field.Number(), expected.name, expected.number)
		}
	}
	for _, forbidden := range []protoreflect.Name{
		"source_locale",
		"translation_spec_version",
		"source_hash",
		"source_revision",
		"source_epoch",
		"request_artifact_digest",
	} {
		if field := message.Fields().ByName(forbidden); field != nil {
			t.Fatalf("TranslationLifecycleEvent.%s must not duplicate request identity", forbidden)
		}
	}
}

func TestValidateTranslationJobRequestArtifact(t *testing.T) {
	if err := ValidateTranslationJobRequestArtifact(nil); err == nil {
		t.Fatal("nil translation job unexpectedly accepted")
	}
	if err := ValidateTranslationJobRequestArtifact(&managev1.TranslationJob{}); err == nil {
		t.Fatal("translation job without target unexpectedly accepted")
	}
	job := &managev1.TranslationJob{
		Target: &managev1.TranslationTarget{
			EntityType: managev1.TranslationEntityType_TRANSLATION_ENTITY_TYPE_PAGE,
			EntityId:   "entity-1",
		},
	}
	if err := ValidateTranslationJobRequestArtifact(job); err == nil {
		t.Fatal("translation job without request artifact digest unexpectedly accepted")
	}
	job.RequestArtifactDigest = "artifact-digest"
	if err := ValidateTranslationJobRequestArtifact(job); err != nil {
		t.Fatalf("valid translation job rejected: %v", err)
	}
}
