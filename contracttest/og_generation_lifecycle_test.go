package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestOgGenerationLifecycleServiceCutover(t *testing.T) {
	admin := managev1.File_api_manage_v1_admin_proto.Services().ByName("AdminService")
	if admin == nil {
		t.Fatal("api.manage.v1.AdminService descriptor is missing")
	}
	for _, name := range []protoreflect.Name{
		"RegenerateOgImage",
		"RegenerateAllOgImages",
		"GetOgGeneration",
		"GetLatestOgGeneration",
		"GetOgGenerationRun",
	} {
		if admin.Methods().ByName(name) == nil {
			t.Errorf("AdminService must expose %s", name)
		}
	}
	internal := intrav1.File_api_intra_v1_og_proto.Services().ByName("InternalOgService")
	if internal == nil {
		t.Fatal("api.intra.v1.InternalOgService descriptor is missing")
	}
	for _, name := range []protoreflect.Name{
		"ClaimOgGeneration",
		"CompleteOgGeneration",
		"FailOgGeneration",
	} {
		if internal.Methods().ByName(name) == nil {
			t.Errorf("InternalOgService must expose %s", name)
		}
	}
}

func TestOgFailuresAreTerminalAndRegenerationUsesANewIdentity(t *testing.T) {
	admin := managev1.File_api_manage_v1_admin_proto.Services().ByName("AdminService")
	if admin.Methods().ByName("RegenerateOgImage") == nil {
		t.Fatal("AdminService must expose RegenerateOgImage as the explicit recovery action")
	}

	internal := intrav1.File_api_intra_v1_og_proto.Services().ByName("InternalOgService")
	if got := internal.Methods().Len(); got != 3 {
		t.Fatalf("InternalOgService methods = %d, want 3 terminal lifecycle methods", got)
	}
	if internal.Methods().ByName("DeferOgGeneration") != nil {
		t.Fatal("InternalOgService must not expose DeferOgGeneration")
	}

	claimResult := intrav1.File_api_intra_v1_og_proto.Enums().ByName("OgGenerationClaimResult")
	if got := claimResult.Values().Len(); got != 3 {
		t.Fatalf("OgGenerationClaimResult values = %d, want 3 without retry-later", got)
	}
	if claimResult.Values().ByName("OG_GENERATION_CLAIM_RESULT_RETRY_LATER") != nil {
		t.Fatal("OgGenerationClaimResult must not expose RETRY_LATER")
	}
}

func TestOgLifecycleOmitsDispatchLedgerState(t *testing.T) {
	status := managev1.OgGenerationStatus(0).Descriptor()
	wantStatusNames := []protoreflect.Name{
		"OG_GENERATION_STATUS_UNSPECIFIED",
		"OG_GENERATION_STATUS_QUEUED",
		"OG_GENERATION_STATUS_PROCESSING",
		"OG_GENERATION_STATUS_READY",
		"OG_GENERATION_STATUS_FAILED",
		"OG_GENERATION_STATUS_SUPERSEDED",
		"OG_GENERATION_STATUS_CANCELLED",
	}
	if status.Values().Len() != len(wantStatusNames) {
		t.Fatalf("OgGenerationStatus values = %d, want %d", status.Values().Len(), len(wantStatusNames))
	}
	for index, wantName := range wantStatusNames {
		value := status.Values().Get(index)
		if value.Name() != wantName || value.Number() != protoreflect.EnumNumber(index) {
			t.Fatalf("OgGenerationStatus[%d] = %s/%d, want %s/%d", index, value.Name(), value.Number(), wantName, index)
		}
	}
	if status.Values().ByName("OG_GENERATION_STATUS_DISPATCHED") != nil {
		t.Fatal("OgGenerationStatus must not expose transport dispatch state")
	}

	run := (&managev1.OgGenerationRun{}).ProtoReflect().Descriptor()
	wantRunFields := []protoreflect.Name{
		"run_id",
		"status",
		"generation_count",
		"queued_count",
		"processing_count",
		"ready_count",
		"failed_count",
		"superseded_count",
		"cancelled_count",
		"failures",
		"created_at",
		"updated_at",
		"completed_at",
	}
	if run.Fields().Len() != len(wantRunFields) {
		t.Fatalf("OgGenerationRun fields = %d, want %d", run.Fields().Len(), len(wantRunFields))
	}
	for index, wantName := range wantRunFields {
		field := run.Fields().Get(index)
		wantNumber := protoreflect.FieldNumber(index + 1)
		if field.Name() != wantName || field.Number() != wantNumber {
			t.Fatalf("OgGenerationRun field %d = %s/%d, want %s/%d", index, field.Name(), field.Number(), wantName, wantNumber)
		}
	}
	if run.Fields().ByName("dispatched_count") != nil {
		t.Fatal("OgGenerationRun must not expose transport dispatch count")
	}
}

func TestOgGenerationTargetsAreExplicit(t *testing.T) {
	selection := managev1.File_api_manage_v1_admin_proto.Messages().ByName("OgTargetSelection")
	if selection == nil {
		t.Fatal("OgTargetSelection descriptor is missing")
	}
	assertOneofFields(t, selection, "target",
		expectedField{"primary", 1, protoreflect.MessageKind, true},
		expectedField{"locale", 2, protoreflect.StringKind, true},
		expectedField{"all_locales", 3, protoreflect.MessageKind, true},
	)

	target := managev1.File_api_manage_v1_events_proto.Messages().ByName("OgGenerationTarget")
	if target == nil {
		t.Fatal("OgGenerationTarget descriptor is missing")
	}
	assertOneofFields(t, target, "scope",
		expectedField{"entity_type", 1, protoreflect.EnumKind, false},
		expectedField{"entity_id", 2, protoreflect.StringKind, false},
		expectedField{"entity", 3, protoreflect.MessageKind, true},
		expectedField{"locale", 4, protoreflect.MessageKind, true},
	)
}

func TestCanonicalFormSaveResponseOmitsRunIdentity(t *testing.T) {
	message := intrav1.File_api_intra_v1_form_proto.Messages().ByName("SaveFormDocumentResponse")
	if message == nil {
		t.Fatal("SaveFormDocumentResponse descriptor is missing")
	}
	if message.Fields().ByName("og_generation_run_id") != nil {
		t.Fatal("canonical Form save response must not expose og_generation_run_id")
	}
}

func TestBlockSourceMetadataMutationOmitsHotPathRunIdentity(t *testing.T) {
	response := (&intrav1.UpdatePageLocaleMetadataResponse{}).ProtoReflect().Descriptor()
	if response.Fields().ByName("og_generation_run_id") != nil {
		t.Fatal("Block hot mutation response must not expose og_generation_run_id")
	}
	if response.Fields().ByName("document_revision") == nil || response.Fields().ByName("source_changed") == nil {
		t.Fatal("Block hot mutation response must expose the minimal mutation ACK")
	}
}

func TestOgAffectingPublicMutationsExposeRunIdentity(t *testing.T) {
	tests := []struct {
		file    protoreflect.FileDescriptor
		message protoreflect.Name
	}{
		{managev1.File_api_manage_v1_artist_proto, "SetArtistImageResponse"},
		{managev1.File_api_manage_v1_form_proto, "SetFormFeaturedImageResponse"},
		{managev1.File_api_manage_v1_label_proto, "SetLabelImageResponse"},
		{managev1.File_api_manage_v1_page_proto, "SetPageFeaturedImageResponse"},
		{managev1.File_api_manage_v1_post_proto, "SetPostFeaturedImageResponse"},
		{managev1.File_api_manage_v1_release_proto, "SetReleaseArtworkResponse"},
		{managev1.File_api_manage_v1_series_proto, "SetSeriesFeaturedImageResponse"},
		{managev1.File_api_manage_v1_site_setting_proto, "SetSettingResponse"},
		{managev1.File_api_manage_v1_site_setting_proto, "SetManySettingsResponse"},
		{managev1.File_api_manage_v1_work_proto, "SetWorkFeaturedImageResponse"},
	}
	for _, test := range tests {
		message := test.file.Messages().ByName(test.message)
		if message == nil {
			t.Errorf("%s descriptor is missing", test.message)
			continue
		}
		if message.Fields().ByName("og_generation_run_id") == nil {
			t.Errorf("%s must expose og_generation_run_id", test.message)
		}
	}
}

func TestOgAffectingAssetDeletesUseFocusedResponse(t *testing.T) {
	response := managev1.File_api_manage_v1_common_proto.Messages().ByName("OgAssetDeleteResponse")
	if response == nil {
		t.Fatal("OgAssetDeleteResponse descriptor is missing")
	}
	if response.Fields().Len() != 2 {
		t.Errorf("OgAssetDeleteResponse must have exactly 2 fields, got %d", response.Fields().Len())
	}
	success := response.Fields().ByName("success")
	if success == nil || success.Number() != 1 || success.Kind() != protoreflect.BoolKind {
		t.Error("OgAssetDeleteResponse.success must be bool field 1")
	}
	assertOptionalStringField(t, response, "og_generation_run_id", 2)

	generic := managev1.File_api_manage_v1_common_proto.Messages().ByName("DeleteResponse")
	if generic == nil {
		t.Fatal("DeleteResponse descriptor is missing")
	}
	if generic.Fields().Len() != 1 {
		t.Errorf("generic DeleteResponse must remain unchanged with exactly 1 field, got %d", generic.Fields().Len())
	}
	if field := generic.Fields().ByName("success"); field == nil || field.Number() != 1 || field.Kind() != protoreflect.BoolKind {
		t.Error("generic DeleteResponse.success must remain bool field 1")
	}

	expected := map[protoreflect.FullName]bool{
		"api.manage.v1.ArtistService.DeleteArtistImage":         false,
		"api.manage.v1.FormService.DeleteFormFeaturedImage":     false,
		"api.manage.v1.LabelService.DeleteLabelImage":           false,
		"api.manage.v1.PageService.DeletePageFeaturedImage":     false,
		"api.manage.v1.PostService.DeletePostFeaturedImage":     false,
		"api.manage.v1.ReleaseService.DeleteReleaseArtwork":     false,
		"api.manage.v1.SeriesService.DeleteSeriesFeaturedImage": false,
		"api.manage.v1.WorkService.DeleteWorkFeaturedImage":     false,
	}
	var actual int
	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		services := file.Services()
		for serviceIndex := 0; serviceIndex < services.Len(); serviceIndex++ {
			service := services.Get(serviceIndex)
			methods := service.Methods()
			for methodIndex := 0; methodIndex < methods.Len(); methodIndex++ {
				method := methods.Get(methodIndex)
				if method.Output().FullName() != response.FullName() {
					continue
				}
				actual++
				fullName := service.FullName().Append(method.Name())
				if _, ok := expected[fullName]; !ok {
					t.Errorf("unexpected RPC %s uses OgAssetDeleteResponse", fullName)
					continue
				}
				expected[fullName] = true
			}
		}
		return true
	})
	if actual != len(expected) {
		t.Errorf("exactly %d RPCs must use OgAssetDeleteResponse, got %d", len(expected), actual)
	}
	for fullName, found := range expected {
		if !found {
			t.Errorf("RPC %s must use OgAssetDeleteResponse", fullName)
		}
	}
}

func TestOgQueuePayloadOnlyCarriesGenerationIdentity(t *testing.T) {
	job := managev1.File_api_manage_v1_events_proto.Messages().ByName("OgGenerationJob")
	if job == nil {
		t.Fatal("OgGenerationJob descriptor is missing")
	}
	if job.Fields().Len() != 1 || job.Fields().ByName("generation_id") == nil {
		t.Error("OgGenerationJob must carry only generation_id")
	}
}

type expectedField struct {
	name    protoreflect.Name
	number  protoreflect.FieldNumber
	kind    protoreflect.Kind
	inOneof bool
}

func assertOneofFields(t *testing.T, message protoreflect.MessageDescriptor, oneofName protoreflect.Name, expected ...expectedField) {
	t.Helper()
	if message.Oneofs().Len() != 1 {
		t.Errorf("%s must have exactly 1 oneof, got %d", message.Name(), message.Oneofs().Len())
	}
	oneof := message.Oneofs().ByName(oneofName)
	if oneof == nil {
		t.Fatalf("%s.%s oneof is missing", message.Name(), oneofName)
	}
	if message.Fields().Len() != len(expected) {
		t.Errorf("%s must have exactly %d fields, got %d", message.Name(), len(expected), message.Fields().Len())
	}
	expectedOneofFields := 0
	for _, want := range expected {
		if want.inOneof {
			expectedOneofFields++
		}
		field := message.Fields().ByName(want.name)
		if field == nil {
			t.Errorf("%s must include %s", message.Name(), want.name)
			continue
		}
		if field.Number() != want.number {
			t.Errorf("%s.%s must use field number %d, got %d", message.Name(), want.name, want.number, field.Number())
		}
		if field.Kind() != want.kind {
			t.Errorf("%s.%s must use kind %s, got %s", message.Name(), want.name, want.kind, field.Kind())
		}
		containingOneof := field.ContainingOneof()
		if want.inOneof {
			if containingOneof != oneof {
				t.Errorf("%s.%s must belong to oneof %s", message.Name(), want.name, oneofName)
			}
		} else if containingOneof != nil {
			t.Errorf("%s.%s must not belong to a oneof", message.Name(), want.name)
		}
	}
	if oneof.Fields().Len() != expectedOneofFields {
		t.Errorf("%s.%s must have exactly %d fields, got %d", message.Name(), oneofName, expectedOneofFields, oneof.Fields().Len())
	}
}

func assertOptionalStringField(t *testing.T, message protoreflect.MessageDescriptor, fieldName protoreflect.Name, fieldNumber protoreflect.FieldNumber) {
	t.Helper()
	field := message.Fields().ByName(fieldName)
	if field == nil {
		t.Errorf("%s must expose %s", message.Name(), fieldName)
		return
	}
	if field.Number() != fieldNumber {
		t.Errorf("%s.%s must use field number %d, got %d", message.Name(), fieldName, fieldNumber, field.Number())
	}
	if field.Kind() != protoreflect.StringKind {
		t.Errorf("%s.%s must be a string", message.Name(), fieldName)
	}
	if !field.HasOptionalKeyword() {
		t.Errorf("%s.%s must be optional", message.Name(), fieldName)
	}
}
