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
		"contact_email": 18,
	} {
		field := open.Fields().ByName(name)
		if field == nil || field.Number() != number || field.Kind() != protoreflect.StringKind {
			t.Errorf("api.open.v1.SiteSettings.%s must be public string field %d", name, number)
		}
	}
}

func TestContactEmailRemovesLegacyFields(t *testing.T) {
	for _, contract := range []struct {
		descriptor       protoreflect.MessageDescriptor
		canonical, legal protoreflect.FieldNumber
	}{
		{(&managev1.PublicSettings{}).ProtoReflect().Descriptor(), 26, 5},
		{(&openv1.SiteSettings{}).ProtoReflect().Descriptor(), 18, 17},
	} {
		for name, number := range map[protoreflect.Name]protoreflect.FieldNumber{
			"contact_email": contract.canonical, "legal_email": contract.legal,
		} {
			field := contract.descriptor.Fields().ByName(name)
			if field == nil || field.Number() != number || field.Kind() != protoreflect.StringKind {
				t.Fatalf("%s.%s must retain string field %d", contract.descriptor.FullName(), name, number)
			}
		}
		for _, name := range []protoreflect.Name{"support_email", "privacy_email"} {
			if contract.descriptor.Fields().ByName(name) != nil || contract.descriptor.ReservedNames().Has(name) {
				t.Errorf("%s.%s must be removed without a reservation", contract.descriptor.FullName(), name)
			}
		}
	}
}

func TestSiteLogoVariantsReplaceLegacyLogoFields(t *testing.T) {
	for _, contract := range []struct {
		descriptor protoreflect.MessageDescriptor
		slots      map[protoreflect.Name]protoreflect.FieldNumber
		removed    []protoreflect.Name
	}{
		{
			(&managev1.PublicSettings{}).ProtoReflect().Descriptor(),
			map[protoreflect.Name]protoreflect.FieldNumber{
				"logo_full_light_asset": 17, "logo_full_dark_asset": 18, "logo_email_light_asset": 19,
				"logo_symbol_light_asset": 27, "logo_symbol_dark_asset": 28, "logo_email_dark_asset": 29,
			},
			[]protoreflect.Name{"logo_light_asset", "logo_dark_asset", "logo_email_asset"},
		},
		{
			(&openv1.SiteSettings{}).ProtoReflect().Descriptor(),
			map[protoreflect.Name]protoreflect.FieldNumber{
				"logo_full_light_asset": 12, "logo_full_dark_asset": 13,
				"logo_symbol_light_asset": 19, "logo_symbol_dark_asset": 20,
			},
			[]protoreflect.Name{"logo_light_asset", "logo_dark_asset"},
		},
	} {
		for name, number := range contract.slots {
			field := contract.descriptor.Fields().ByName(name)
			if field == nil || field.Number() != number || field.Kind() != protoreflect.MessageKind ||
				field.Message().FullName() != "api.common.v1.AssetRef" || !field.HasPresence() {
				t.Errorf("%s.%s must be optional AssetRef field %d", contract.descriptor.FullName(), name, number)
			}
		}
		for _, name := range contract.removed {
			if contract.descriptor.Fields().ByName(name) != nil || contract.descriptor.ReservedNames().Has(name) {
				t.Errorf("%s.%s must be removed without a reservation", contract.descriptor.FullName(), name)
			}
		}
	}
	if (&openv1.SiteSettings{}).ProtoReflect().Descriptor().Fields().ByName("logo_email_light_asset") != nil {
		t.Error("email logos are rendered by the API and must not be public manifest fields")
	}
}
