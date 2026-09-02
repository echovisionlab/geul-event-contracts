package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFormCollaborationUsesOneServerDerivedSourceDocument(t *testing.T) {
	service := intrav1.File_api_intra_v1_form_proto.Services().ByName("InternalFormService")
	if service.Methods().ByName("SaveDocument") == nil || service.Methods().ByName("LoadDocument") == nil {
		t.Fatal("Form source document RPCs are missing")
	}
	if service.Methods().ByName("SaveTranslationDocument") != nil || service.Methods().ByName("LoadTranslationDocument") != nil {
		t.Fatal("Form target translation document RPCs must be absent")
	}

	sharedSave := (&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor()
	meta := requireMessageField(t, sharedSave, "meta", 2, protoreflect.MessageKind, "api.intra.v1.FormMeta")
	if !meta.HasOptionalKeyword() {
		t.Errorf("%s.meta must be optional", sharedSave.FullName())
	}
	requireMessageField(t, sharedSave, "locale", 4, protoreflect.StringKind, "")
	requireMessageField(t, sharedSave, "expected_document_revision", 6, protoreflect.StringKind, "")
	requireOptionalField(t, sharedSave, "expected_target_revision", 7, protoreflect.StringKind)
	sharedSaveResponse := (&intrav1.SaveFormDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, sharedSaveResponse, "locale", 2, protoreflect.StringKind, "")
	requireMessageField(t, sharedSaveResponse, "document_revision", 3, protoreflect.StringKind, "")
	requireOptionalField(t, sharedSaveResponse, "target_revision", 4, protoreflect.StringKind)

	sharedLoadRequest := (&intrav1.LoadFormDocumentRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, sharedLoadRequest, "locale", 2, protoreflect.StringKind, "")
	sharedLoad := (&intrav1.LoadFormDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, sharedLoad, "source_metadata", 1, protoreflect.MessageKind, "api.intra.v1.FormMeta")
	localeMetadata := requireMessageField(t, sharedLoad, "locale_metadata", 2, protoreflect.MessageKind, "api.intra.v1.FormMeta")
	if !localeMetadata.HasOptionalKeyword() {
		t.Errorf("%s.locale_metadata must be optional", sharedLoad.FullName())
	}
	requireMessageField(t, sharedLoad, "source_locale", 3, protoreflect.StringKind, "")
	requireMessageField(t, sharedLoad, "locale", 4, protoreflect.StringKind, "")
	requireMessageField(t, sharedLoad, "locale_exists", 5, protoreflect.BoolKind, "")
	requireMessageField(t, sharedLoad, "document_revision", 7, protoreflect.StringKind, "")
	requireOptionalField(t, sharedLoad, "target_revision", 8, protoreflect.StringKind)
	formMeta := (&intrav1.FormMeta{}).ProtoReflect().Descriptor()
	requireOptionalField(t, formMeta, "title", 1, protoreflect.StringKind)
	requireOptionalField(t, formMeta, "schema", 2, protoreflect.StringKind)
	for _, removed := range []protoreflect.Name{"yjs_state", "patch_mask", "og_generation_run_id", "expected_source_locale", "expected_current_edit_hash", "edit_hash", "source_revision_checkpoint"} {
		if sharedSave.Fields().ByName(removed) != nil || sharedLoad.Fields().ByName(removed) != nil {
			t.Fatalf("Form collaboration must not expose duplicate CAS field %s", removed)
		}
		if formMeta.Fields().ByName(removed) != nil {
			t.Fatalf("Form canonical payload must not expose removed field %s", removed)
		}
	}
}
