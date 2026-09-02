package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationXLIFFInterchangeUsesFileHandlesAndExplicitMode(t *testing.T) {
	mode := managev1.TranslationInterchangeMode(0).Descriptor()
	if got := mode.Values().Len(); got != 3 {
		t.Fatalf("TranslationInterchangeMode value count = %d, want 3", got)
	}
	for index, expected := range []protoreflect.Name{
		"TRANSLATION_INTERCHANGE_MODE_UNSPECIFIED",
		"TRANSLATION_INTERCHANGE_MODE_PATCH",
		"TRANSLATION_INTERCHANGE_MODE_REPLACE",
	} {
		if got := mode.Values().Get(index).Name(); got != expected {
			t.Fatalf("TranslationInterchangeMode[%d] = %s, want %s", index, got, expected)
		}
	}

	export := (&managev1.ExportEntityTranslationXLIFFRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, export, 4)
	requireMessageField(t, export, "target", 1, protoreflect.MessageKind, "api.manage.v1.TranslationTarget")
	requireMessageField(t, export, "target_locale", 2, protoreflect.StringKind, "")
	requireMessageField(t, export, "mode", 3, protoreflect.EnumKind, "api.manage.v1.TranslationInterchangeMode")
	units := requireMessageField(t, export, "unit_handles", 4, protoreflect.StringKind, "")
	if units.Cardinality() != protoreflect.Repeated {
		t.Fatal("ExportEntityTranslationXLIFFRequest.unit_handles must be repeated")
	}

	exported := (&managev1.ExportEntityTranslationXLIFFResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, exported, "artifact", 1, protoreflect.MessageKind, "api.common.v1.ExpiringMediaRef")
	targetRevision := requireMessageField(t, exported, "target_revision", 4, protoreflect.StringKind, "")
	if !targetRevision.HasPresence() {
		t.Fatal("ExportEntityTranslationXLIFFResponse.target_revision must preserve missing target presence")
	}

	importRequest := (&managev1.ImportEntityTranslationXLIFFRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, importRequest, 5)
	requireMessageField(t, importRequest, "file_id", 4, protoreflect.StringKind, "")
	expectedRevision := requireMessageField(t, importRequest, "expected_target_revision", 5, protoreflect.StringKind, "")
	if !expectedRevision.HasPresence() {
		t.Fatal("ImportEntityTranslationXLIFFRequest.expected_target_revision must preserve missing target CAS")
	}
	for index := 0; index < importRequest.Fields().Len(); index++ {
		if importRequest.Fields().Get(index).Kind() == protoreflect.BytesKind {
			t.Fatal("XLIFF import must not carry XML or base64 bytes")
		}
	}

	service := managev1.File_api_manage_v1_translation_proto.Services().ByName("TranslationService")
	for _, method := range []protoreflect.Name{
		"ExportEntityTranslationXLIFF",
		"ImportEntityTranslationXLIFF",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_USER)
	}
}
