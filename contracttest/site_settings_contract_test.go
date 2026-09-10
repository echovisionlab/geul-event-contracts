package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSiteOriginIsRuntimeOnlyAndRemainsPubliclyProjected(t *testing.T) {
	public := (&managev1.PublicSettings{}).ProtoReflect().Descriptor()
	if field := public.Fields().ByName("site_url"); field != nil {
		t.Fatal("api.manage.v1.PublicSettings.site_url must not remain persisted mutable state")
	}
	if field := public.Fields().ByName("site_origin"); field != nil {
		t.Fatal("api.manage.v1.PublicSettings.site_origin must not become persisted mutable state")
	}

	runtime := (&managev1.RuntimeSettings{}).ProtoReflect().Descriptor()
	runtimeOrigin := runtime.Fields().ByName("site_origin")
	if runtimeOrigin == nil || runtimeOrigin.Number() != 1 || runtimeOrigin.Kind() != protoreflect.StringKind {
		t.Fatal("api.manage.v1.RuntimeSettings.site_origin must be read-only string field 1")
	}

	open := (&openv1.SiteSettings{}).ProtoReflect().Descriptor()
	openOrigin := open.Fields().ByName("site_origin")
	if openOrigin == nil || openOrigin.Number() != 11 || openOrigin.Kind() != protoreflect.StringKind {
		t.Fatal("api.open.v1.SiteSettings.site_origin must remain public string field 11")
	}
}

func TestLoaderMutationUsesDedicatedRelationRPCs(t *testing.T) {
	service := managev1.File_api_manage_v1_site_setting_proto.Services().ByName("SiteSettingService")
	if service == nil {
		t.Fatal("api.manage.v1.SiteSettingService is missing")
	}
	for _, methodName := range []protoreflect.Name{"AddSiteLoaderAsset", "RemoveSiteLoaderAsset"} {
		if method := service.Methods().ByName(methodName); method == nil {
			t.Errorf("api.manage.v1.SiteSettingService.%s is missing", methodName)
		}
	}

	public := (&managev1.PublicSettings{}).ProtoReflect().Descriptor()
	if field := public.Fields().ByName("loader_file_id"); field != nil {
		t.Error("api.manage.v1.PublicSettings.loader_file_id must not exist")
	}
}

func TestPublicSiteSettingsProjectContactEmails(t *testing.T) {
	open := (&openv1.SiteSettings{}).ProtoReflect().Descriptor()
	for name, number := range map[protoreflect.Name]protoreflect.FieldNumber{
		"legal_email":   17,
		"support_email": 18,
		"privacy_email": 19,
	} {
		field := open.Fields().ByName(name)
		if field == nil || field.Number() != number || field.Kind() != protoreflect.StringKind {
			t.Errorf("api.open.v1.SiteSettings.%s must be public string field %d", name, number)
		}
	}
}
