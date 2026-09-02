package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestContentUpdatedExactLocaleRelayWireHardCut(t *testing.T) {
	descriptor := (&managev1.ContentUpdatedEvent{}).ProtoReflect().Descriptor()
	fields := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
		kind   protoreflect.Kind
	}{
		{"locale", 9, protoreflect.StringKind},
		{"locale_exists", 10, protoreflect.BoolKind},
		{"target_revision", 11, protoreflect.StringKind},
	}
	for _, expected := range fields {
		field := descriptor.Fields().ByName(expected.name)
		if field == nil || field.Number() != expected.number || field.Kind() != expected.kind || !field.HasPresence() || field.Cardinality() == protoreflect.Repeated {
			t.Errorf("ContentUpdatedEvent.%s must be optional %s at field %d", expected.name, expected.kind, expected.number)
		}
	}
}

func TestContentUpdatedExactLocaleRelaySourceContract(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/events.proto")
	declaration := sourceDeclaration(t, source, "message", "ContentUpdatedEvent")
	requireSourceContains(t, declaration,
		"optional string locale = 9;",
		"optional bool locale_exists = 10;",
		"optional string target_revision = 11;",
		"locale tuple is absent for non-locale updates",
		"locale_exists presence",
		"false means target deletion and forbids target_revision",
		"absent for source-locale writes and target deletion",
		"must not masquerade as a source document revision change",
	)
	asyncAPI := readRepositorySource(t, "asyncapi/asyncapi.yaml")
	requireSourceContains(t, asyncAPI,
		"Locale-owned authoring carries an exact locale tuple",
		"target deletion",
		"source-locale writes",
		"document_state_changed false",
	)
}
