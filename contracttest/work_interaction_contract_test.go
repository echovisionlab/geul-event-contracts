package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestWorkStatusAndMutationContract(t *testing.T) {
	get := (&openv1.GetWorkRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, get, "share_password", 3, protoreflect.StringKind)
	getResponse := (&openv1.GetWorkResponse{}).ProtoReflect().Descriptor()
	blockMedia := requireMessageField(t, getResponse, "block_media", 2, protoreflect.MessageKind, "api.content.v1.ContentBlockMediaItem")
	if blockMedia.Cardinality() != protoreflect.Repeated {
		t.Fatal("api.open.v1.GetWorkResponse.block_media must preserve every Block selector")
	}

	update := (&managev1.UpdateWorkRequest{}).ProtoReflect().Descriptor()
	if field := update.Fields().ByName("status"); field != nil {
		t.Fatal("api.manage.v1.UpdateWorkRequest.status must not bypass PublishWork and UnpublishWork")
	}
	workMeta := (&intrav1.UpdateWorkLocaleMetadataRequest{}).ProtoReflect().Descriptor()
	if field := workMeta.Fields().ByName("status"); field != nil {
		t.Fatal("api.intra.v1.UpdateWorkLocaleMetadataRequest.status must not bypass Work lifecycle RPCs")
	}

	if managev1.WorkStatus_WORK_STATUS_ARCHIVED.Number() != protoreflect.EnumNumber(3) {
		t.Fatal("manage archived Work read state must retain wire value 3")
	}
	if openv1.WorkStatus_WORK_STATUS_ARCHIVED.Number() != protoreflect.EnumNumber(3) {
		t.Fatal("open archived Work read state must retain wire value 3")
	}

	service := managev1.File_api_manage_v1_work_proto.Services().ByName("WorkService")
	if service == nil {
		t.Fatal("api.manage.v1.WorkService is missing")
	}
	for _, method := range []string{
		"ListWorksAdmin",
		"CreateWork",
		"DeleteWork",
		"PublishWork",
		"UnpublishWork",
		"GetWork",
		"UpdateWork",
		"SetWorkFeaturedImage",
		"DeleteWorkFeaturedImage",
		"GetWorkCredits",
		"AddWorkCredit",
		"UpdateWorkCredit",
		"DeleteWorkCredit",
		"CreateWorkCreditGroup",
		"UpdateWorkCreditGroup",
		"DeleteWorkCreditGroup",
		"ListWorkVersions",
		"RestoreWorkVersion",
		"CheckWorkSlugAvailable",
	} {
		requireMethodTier(t, service, protoreflect.Name(method), policyv1.AuthorizationRole_ADMIN)
	}
	requireMethodTier(t, service, "ListMyCreditedWorks", policyv1.AuthorizationRole_USER)
}
