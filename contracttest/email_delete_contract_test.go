package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestEmailTemplateDeletionContract(t *testing.T) {
	template := (&managev1.EmailTemplate{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, template, 16)

	list := (&managev1.ListEmailTemplatesAdminRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, list, 3)
	requireMessageField(t, list, "pagination", 1, protoreflect.MessageKind, "api.common.v1.PaginationRequest")
	requireMessageField(t, list, "filters", 2, protoreflect.MessageKind, "api.common.v1.FilterSpec")
	requireMessageField(t, list, "sorts", 3, protoreflect.MessageKind, "api.common.v1.SortSpec")

	request := (&managev1.DeleteEmailTemplateRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 1)
	requireMessageField(t, request, "id", 1, protoreflect.StringKind, "")

	file := managev1.File_api_manage_v1_email_template_proto
	requireEmailDeleteService(
		t,
		file.Services().ByName("EmailTemplateService"),
		"DeleteEmailTemplate",
		"api.manage.v1.DeleteEmailTemplateRequest",
	)

}

func TestEmailLayoutDeletionContract(t *testing.T) {
	layout := (&managev1.EmailLayout{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, layout, 10)

	list := (&managev1.ListEmailLayoutsAdminRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, list, 3)
	requireMessageField(t, list, "pagination", 1, protoreflect.MessageKind, "api.common.v1.PaginationRequest")
	requireMessageField(t, list, "filters", 2, protoreflect.MessageKind, "api.common.v1.FilterSpec")
	requireMessageField(t, list, "sorts", 3, protoreflect.MessageKind, "api.common.v1.SortSpec")

	request := (&managev1.DeleteEmailLayoutRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 1)
	requireMessageField(t, request, "id", 1, protoreflect.StringKind, "")

	file := managev1.File_api_manage_v1_email_layout_proto
	requireEmailDeleteService(
		t,
		file.Services().ByName("EmailLayoutService"),
		"DeleteEmailLayout",
		"api.manage.v1.DeleteEmailLayoutRequest",
	)
}

func requireEmailDeleteService(
	t *testing.T,
	service protoreflect.ServiceDescriptor,
	methodName protoreflect.Name,
	request protoreflect.FullName,
) {
	t.Helper()
	if service == nil {
		t.Fatal("email service descriptor is missing")
	}
	method := service.Methods().ByName(methodName)
	if method == nil {
		t.Fatalf("%s.%s descriptor is missing", service.FullName(), methodName)
	}
	if got := method.Input().FullName(); got != request {
		t.Errorf("%s.%s input = %s, want %s", service.FullName(), methodName, got, request)
	}
	if got := method.Output().FullName(); got != "api.manage.v1.DeleteResponse" {
		t.Errorf("%s.%s output = %s, want api.manage.v1.DeleteResponse", service.FullName(), methodName, got)
	}
	requireMethodTier(t, service, methodName, policyv1.AuthorizationRole_ADMIN)
}
