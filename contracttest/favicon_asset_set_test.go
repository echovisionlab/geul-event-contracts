package contracttest_test

import (
	"testing"

	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFaviconAssetSetShape(t *testing.T) {
	descriptor := (&commonv1.FaviconAssetSet{}).ProtoReflect().Descriptor()
	fields := []struct {
		name     protoreflect.Name
		number   protoreflect.FieldNumber
		optional bool
	}{
		{name: "icon_ico", number: 1},
		{name: "icon_png_16", number: 2},
		{name: "icon_png_32", number: 3},
		{name: "icon_png_48", number: 4},
		{name: "apple_touch_icon_180", number: 5},
		{name: "manifest_icon_192", number: 6},
		{name: "manifest_icon_512", number: 7},
		{name: "icon_svg", number: 8, optional: true},
	}

	for _, expected := range fields {
		field := descriptor.Fields().ByName(expected.name)
		if field == nil {
			t.Fatalf("%s is missing %s", descriptor.FullName(), expected.name)
		}
		if field.Number() != expected.number {
			t.Errorf("%s.%s number = %d, want %d", descriptor.FullName(), expected.name, field.Number(), expected.number)
		}
		if field.Kind() != protoreflect.MessageKind || field.Message().FullName() != "api.common.v1.AssetRef" {
			t.Errorf("%s.%s must be an AssetRef", descriptor.FullName(), expected.name)
		}
		if field.HasOptionalKeyword() != expected.optional {
			t.Errorf("%s.%s optional keyword = %t, want %t", descriptor.FullName(), expected.name, field.HasOptionalKeyword(), expected.optional)
		}
	}
}

func TestFaviconAssetSetIsProjectedToSiteSettings(t *testing.T) {
	tests := []struct {
		name       string
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{name: "manage", descriptor: (&managev1.PublicSettings{}).ProtoReflect().Descriptor(), number: 26},
		{name: "open", descriptor: (&openv1.SiteSettings{}).ProtoReflect().Descriptor(), number: 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := test.descriptor.Fields().ByName("favicon_asset_set")
			if field == nil {
				t.Fatalf("%s is missing favicon_asset_set", test.descriptor.FullName())
			}
			if field.Number() != test.number || field.Kind() != protoreflect.MessageKind ||
				field.Message().FullName() != "api.common.v1.FaviconAssetSet" || !field.HasPresence() {
				t.Fatalf("%s.favicon_asset_set has an invalid contract", test.descriptor.FullName())
			}
		})
	}
}

func TestFaviconDerivativeTypesAreStable(t *testing.T) {
	wants := map[managev1.FileDerivativeType]int32{
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_ICO:             5,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_PNG_16:          6,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_PNG_32:          7,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_PNG_48:          8,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_APPLE_TOUCH_180: 9,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_MANIFEST_192:    10,
		managev1.FileDerivativeType_FILE_DERIVATIVE_TYPE_FAVICON_MANIFEST_512:    11,
	}

	for derivativeType, value := range wants {
		if got := int32(derivativeType); got != value {
			t.Errorf("%s = %d, want %d", derivativeType, got, value)
		}
	}
}
