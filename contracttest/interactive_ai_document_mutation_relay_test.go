package contracttest

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestInteractiveAIDocumentMutationRelayContract(t *testing.T) {
	descriptor := (&intrav1.RelayInteractiveAIDocumentMutationRequest{}).ProtoReflect().Descriptor()
	wantFields := []struct {
		name     protoreflect.Name
		kind     protoreflect.Kind
		message  protoreflect.FullName
		repeated bool
		optional bool
	}{
		{name: "mutation_id", kind: protoreflect.StringKind},
		{name: "origin", kind: protoreflect.EnumKind},
		{name: "document", kind: protoreflect.MessageKind, message: "api.manage.v1.AIDocumentReference"},
		{name: "locale", kind: protoreflect.MessageKind, message: "api.manage.v1.AIDocumentLocale"},
		{name: "expected_document_revision", kind: protoreflect.StringKind},
		{name: "accepted_document_revision", kind: protoreflect.StringKind},
		{name: "operations", kind: protoreflect.MessageKind, message: "api.manage.v1.AIDocumentOperation", repeated: true},
		{name: "actor_member_id", kind: protoreflect.StringKind},
		{name: "expected_target_revision", kind: protoreflect.StringKind, optional: true},
		{name: "accepted_target_revision", kind: protoreflect.StringKind, optional: true},
	}

	if descriptor.Fields().Len() != len(wantFields) {
		t.Fatalf("%s fields = %d, want %d", descriptor.FullName(), descriptor.Fields().Len(), len(wantFields))
	}
	for index, want := range wantFields {
		field := descriptor.Fields().Get(index)
		if field.Name() != want.name || field.Number() != protoreflect.FieldNumber(index+1) || field.Kind() != want.kind || field.IsList() != want.repeated {
			t.Errorf("field %d = %s/%d/%s/repeated=%t, want %s/%d/%s/repeated=%t", index, field.Name(), field.Number(), field.Kind(), field.IsList(), want.name, index+1, want.kind, want.repeated)
		}
		if field.HasOptionalKeyword() != want.optional {
			t.Errorf("%s.%s optional = %t, want %t", descriptor.FullName(), field.Name(), field.HasOptionalKeyword(), want.optional)
		}
		if want.message != "" && field.Message().FullName() != want.message {
			t.Errorf("%s.%s message = %s, want %s", descriptor.FullName(), field.Name(), field.Message().FullName(), want.message)
		}
	}

	origins := intrav1.InteractiveAIDocumentMutationOrigin_name
	wantOrigins := map[int32]string{
		0: "INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_UNSPECIFIED",
		1: "INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_IN_EDITOR_AI",
		2: "INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP",
	}
	if len(origins) != len(wantOrigins) {
		t.Fatalf("origin values = %v, want %v", origins, wantOrigins)
	}
	for value, name := range wantOrigins {
		if origins[value] != name {
			t.Errorf("origin %d = %q, want %q", value, origins[value], name)
		}
	}

	for _, forbidden := range []protoreflect.Name{
		"session_id",
		"prompt",
		"pat",
		"token",
		"inverse",
		"history",
		"translation_job_id",
		"accepted_source_epoch",
	} {
		if field := descriptor.Fields().ByName(forbidden); field != nil {
			t.Errorf("%s must not expose forbidden field %s", descriptor.FullName(), forbidden)
		}
	}

	service := intrav1.File_api_intra_v1_collaboration_proto.Services().ByName("InternalCollaborationRelayService")
	if service == nil {
		t.Fatal("InternalCollaborationRelayService is missing")
	}
	method := service.Methods().ByName("RelayInteractiveAIDocumentMutation")
	if method == nil {
		t.Fatal("RelayInteractiveAIDocumentMutation is missing")
	}
	if got, want := method.Input().FullName(), descriptor.FullName(); got != want {
		t.Errorf("method input = %s, want %s", got, want)
	}
	if got, want := method.Output().FullName(), protoreflect.FullName("api.intra.v1.RelayInteractiveAIDocumentMutationResponse"); got != want {
		t.Errorf("method output = %s, want %s", got, want)
	}
}
