package contracttest_test

import (
	"testing"

	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCommonDocumentLayoutShape(t *testing.T) {
	layout := (&commonv1.DocumentLayout{}).ProtoReflect().Descriptor()

	requireEnumField(
		t,
		layout,
		"content_height",
		1,
		"api.common.v1.DocumentContentHeight",
		[]protoreflect.Name{
			"DOCUMENT_CONTENT_HEIGHT_UNSPECIFIED",
			"DOCUMENT_CONTENT_HEIGHT_CONTENT",
			"DOCUMENT_CONTENT_HEIGHT_VIEWPORT",
		},
	)
	requireEnumField(
		t,
		layout,
		"page_chrome",
		2,
		"api.common.v1.DocumentRegionPlacement",
		[]protoreflect.Name{
			"DOCUMENT_REGION_PLACEMENT_UNSPECIFIED",
			"DOCUMENT_REGION_PLACEMENT_FLOW",
			"DOCUMENT_REGION_PLACEMENT_PINNED",
		},
	)
	requireEnumField(
		t,
		layout,
		"footer",
		3,
		"api.common.v1.DocumentRegionPlacement",
		[]protoreflect.Name{
			"DOCUMENT_REGION_PLACEMENT_UNSPECIFIED",
			"DOCUMENT_REGION_PLACEMENT_FLOW",
			"DOCUMENT_REGION_PLACEMENT_PINNED",
		},
	)
}

func TestDocumentLayoutIsProjectedOnCurrentDocumentDTOsAndOwningMutations(t *testing.T) {
	tests := []struct {
		name        string
		descriptor  protoreflect.MessageDescriptor
		fieldNumber protoreflect.FieldNumber
	}{
		{name: "manage Page", descriptor: (&managev1.Page{}).ProtoReflect().Descriptor(), fieldNumber: 13},
		{name: "public Page", descriptor: (&openv1.Page{}).ProtoReflect().Descriptor(), fieldNumber: 14},
		{name: "manage Post", descriptor: (&managev1.Post{}).ProtoReflect().Descriptor(), fieldNumber: 19},
		{name: "public Post", descriptor: (&openv1.Post{}).ProtoReflect().Descriptor(), fieldNumber: 20},
		{name: "intra Page document metadata update", descriptor: (&intrav1.UpdatePageDocumentMetadataRequest{}).ProtoReflect().Descriptor(), fieldNumber: 3},
		{name: "manage Post update", descriptor: (&managev1.UpdatePostRequest{}).ProtoReflect().Descriptor(), fieldNumber: 5},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireDocumentLayoutField(t, test.descriptor, test.fieldNumber)
		})
	}
}

func TestPageLayoutHasOneCollaborativeMutationAuthority(t *testing.T) {
	intraUpdate := (&intrav1.UpdatePageDocumentMetadataRequest{}).ProtoReflect().Descriptor()
	requireField(t, intraUpdate, "expected_revision", 2, protoreflect.StringKind)
	requireDocumentLayoutField(t, intraUpdate, 3)
	requireNoField(t, (&managev1.UpdatePageRequest{}).ProtoReflect().Descriptor(), "document_layout")
}

func requireDocumentLayoutField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	fieldNumber protoreflect.FieldNumber,
) {
	t.Helper()

	field := descriptor.Fields().ByName("document_layout")
	if field == nil {
		t.Fatalf("%s is missing document_layout", descriptor.FullName())
	}
	if field.Number() != fieldNumber {
		t.Fatalf("%s.document_layout number = %d, want %d", descriptor.FullName(), field.Number(), fieldNumber)
	}
	if field.Kind() != protoreflect.MessageKind {
		t.Fatalf("%s.document_layout kind = %s, want message", descriptor.FullName(), field.Kind())
	}
	if got, want := field.Message().FullName(), protoreflect.FullName("api.common.v1.DocumentLayout"); got != want {
		t.Fatalf("%s.document_layout type = %s, want %s", descriptor.FullName(), got, want)
	}
}

func requireEnumField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	fieldName protoreflect.Name,
	fieldNumber protoreflect.FieldNumber,
	enumName protoreflect.FullName,
	valueNames []protoreflect.Name,
) {
	t.Helper()

	field := descriptor.Fields().ByName(fieldName)
	if field == nil {
		t.Fatalf("%s is missing %s", descriptor.FullName(), fieldName)
	}
	if field.Number() != fieldNumber {
		t.Fatalf("%s.%s number = %d, want %d", descriptor.FullName(), fieldName, field.Number(), fieldNumber)
	}
	if field.Kind() != protoreflect.EnumKind {
		t.Fatalf("%s.%s kind = %s, want enum", descriptor.FullName(), fieldName, field.Kind())
	}
	if got := field.Enum().FullName(); got != enumName {
		t.Fatalf("%s.%s type = %s, want %s", descriptor.FullName(), fieldName, got, enumName)
	}
	if field.Enum().Values().Len() != len(valueNames) {
		t.Fatalf("%s values = %d, want %d", enumName, field.Enum().Values().Len(), len(valueNames))
	}
	for number, valueName := range valueNames {
		value := field.Enum().Values().ByNumber(protoreflect.EnumNumber(number))
		if value == nil || value.Name() != valueName {
			t.Fatalf("%s value %d = %v, want %s", enumName, number, value, valueName)
		}
	}
}

func requireField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	fieldName protoreflect.Name,
	fieldNumber protoreflect.FieldNumber,
	kind protoreflect.Kind,
) protoreflect.FieldDescriptor {
	t.Helper()

	field := descriptor.Fields().ByName(fieldName)
	if field == nil {
		t.Fatalf("%s is missing %s", descriptor.FullName(), fieldName)
	}
	if field.Number() != fieldNumber {
		t.Fatalf("%s.%s number = %d, want %d", descriptor.FullName(), fieldName, field.Number(), fieldNumber)
	}
	if field.Kind() != kind {
		t.Fatalf("%s.%s kind = %s, want %s", descriptor.FullName(), fieldName, field.Kind(), kind)
	}

	return field
}

func requireNoField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	fieldName protoreflect.Name,
) {
	t.Helper()

	if field := descriptor.Fields().ByName(fieldName); field != nil {
		t.Fatalf("%s unexpectedly has %s field %d", descriptor.FullName(), fieldName, field.Number())
	}
}
