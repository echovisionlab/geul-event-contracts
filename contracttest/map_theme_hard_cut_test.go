package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMapThemeUsesRequiredVariantPairWithoutBuiltInState(t *testing.T) {
	for name, descriptor := range map[string]protoreflect.MessageDescriptor{
		"manage": (&managev1.MapTheme{}).ProtoReflect().Descriptor(),
		"open":   (&openv1.MapTheme{}).ProtoReflect().Descriptor(),
	} {
		light := descriptor.Fields().ByName("light_variant")
		dark := descriptor.Fields().ByName("dark_variant")
		if light == nil || dark == nil {
			t.Fatalf("%s MapTheme must expose singular light_variant and dark_variant", name)
		}
		if light.Cardinality() == protoreflect.Repeated || dark.Cardinality() == protoreflect.Repeated {
			t.Errorf("%s MapTheme variants must not be repeated", name)
		}
	}
	manageTheme := (&managev1.MapTheme{}).ProtoReflect().Descriptor()
	requireMessageField(t, manageTheme, "revision", 8, protoreflect.Int64Kind, "")

	create := (&managev1.CreateMapThemeRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, create, "light_variant", 3, protoreflect.MessageKind, "api.manage.v1.MapThemeVariantInput")
	requireMessageField(t, create, "dark_variant", 4, protoreflect.MessageKind, "api.manage.v1.MapThemeVariantInput")
}

func TestMapThemeDefaultAndCollaborationUseDedicatedCASContracts(t *testing.T) {
	list := (&managev1.ListMapThemesResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, list, "default_map_theme_id", 2, protoreflect.StringKind, "")
	publicSettings := (&managev1.PublicSettings{}).ProtoReflect().Descriptor()
	requireMessageField(t, publicSettings, "default_map_theme_id", 27, protoreflect.StringKind, "")

	manageFile := managev1.File_api_manage_v1_map_theme_proto
	manageService := manageFile.Services().ByName("MapThemeService")
	for _, method := range []protoreflect.Name{"ListMapThemes", "ResolveMapTheme"} {
		requireMethodTier(t, manageService, method, policyv1.AuthorizationRole_USER)
	}
	for _, method := range []protoreflect.Name{
		"GetMapTheme",
		"CreateMapTheme",
		"DeleteMapTheme",
		"CopyMapTheme",
		"SetDefaultMapTheme",
	} {
		requireMethodTier(t, manageService, method, policyv1.AuthorizationRole_ADMIN)
	}

	intraFile := intrav1.File_api_intra_v1_map_proto
	intraService := intraFile.Services().ByName("InternalMapService")
	for _, method := range []protoreflect.Name{"SaveMapThemeSnapshot", "LoadMapThemeSnapshot"} {
		if intraService.Methods().ByName(method) == nil {
			t.Errorf("api.intra.v1.InternalMapService.%s is missing", method)
		}
	}

	snapshot := (&intrav1.MapThemeDocumentSnapshot{}).ProtoReflect().Descriptor()
	requireMessageField(t, snapshot, "name", 1, protoreflect.StringKind, "")
	requireMessageField(t, snapshot, "settings", 2, protoreflect.MessageKind, "api.intra.v1.MapThemeDocumentSettings")
	requireMessageField(t, snapshot, "light_variant", 3, protoreflect.MessageKind, "api.intra.v1.MapThemeDocumentVariant")
	requireMessageField(t, snapshot, "dark_variant", 4, protoreflect.MessageKind, "api.intra.v1.MapThemeDocumentVariant")

	save := (&intrav1.SaveMapThemeSnapshotRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, save, "theme_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, save, "snapshot", 2, protoreflect.MessageKind, "api.intra.v1.MapThemeDocumentSnapshot")
	requireMessageField(t, save, "expected_revision", 3, protoreflect.Int64Kind, "")
	requireMessageField(t, save, "contributor_member_ids", 4, protoreflect.StringKind, "")
	requireMessageField(t, save, "locale", 5, protoreflect.StringKind, "")
	if field := save.Fields().ByName("source_revision_checkpoint"); field != nil {
		t.Fatal("SaveMapThemeSnapshotRequest.source_revision_checkpoint must be removed")
	}
	saveResponse := (&intrav1.SaveMapThemeSnapshotResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, saveResponse, "revision", 2, protoreflect.Int64Kind, "")
	requireMessageField(t, saveResponse, "locale", 3, protoreflect.StringKind, "")

	loadRequest := (&intrav1.LoadMapThemeSnapshotRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, loadRequest, "theme_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, loadRequest, "locale", 2, protoreflect.StringKind, "")
	load := (&intrav1.LoadMapThemeSnapshotResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, load, "snapshot", 1, protoreflect.MessageKind, "api.intra.v1.MapThemeDocumentSnapshot")
	requireMessageField(t, load, "revision", 2, protoreflect.Int64Kind, "")
	requireMessageField(t, load, "locale", 3, protoreflect.StringKind, "")
}

func TestPublicMapThemeBatchResolutionPreservesRequestedIdentity(t *testing.T) {
	openFile := openv1.File_api_open_v1_map_theme_proto
	service := openFile.Services().ByName("MapThemeService")
	if service.Methods().ByName("ResolveByIds") == nil {
		t.Fatal("api.open.v1.MapThemeService.ResolveByIds is missing")
	}

	request := (&openv1.ResolveMapThemesByIdsRequest{}).ProtoReflect().Descriptor()
	requestedIDs := requireMessageField(t, request, "requested_theme_ids", 1, protoreflect.StringKind, "")
	if requestedIDs.Cardinality() != protoreflect.Repeated {
		t.Error("ResolveMapThemesByIdsRequest.requested_theme_ids must be repeated")
	}

	result := (&openv1.ResolveMapThemeByIdResult{}).ProtoReflect().Descriptor()
	requireMessageField(t, result, "requested_theme_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, result, "theme", 2, protoreflect.MessageKind, "api.open.v1.MapTheme")

	response := (&openv1.ResolveMapThemesByIdsResponse{}).ProtoReflect().Descriptor()
	results := requireMessageField(t, response, "results", 1, protoreflect.MessageKind, "api.open.v1.ResolveMapThemeByIdResult")
	if results.Cardinality() != protoreflect.Repeated {
		t.Error("ResolveMapThemesByIdsResponse.results must be repeated")
	}
}
