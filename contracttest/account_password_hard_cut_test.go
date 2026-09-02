package contracttest_test

import (
	"testing"

	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestAccountAndMemberServicesRemainSplit(t *testing.T) {
	if managev1.File_api_manage_v1_member_proto.Services().ByName("MemberService") == nil {
		t.Fatal("api.manage.v1.MemberService descriptor is missing")
	}
	if managev1.File_api_manage_v1_account_proto.Services().ByName("AccountService") == nil {
		t.Fatal("api.manage.v1.AccountService descriptor is missing")
	}
	if openv1.File_api_open_v1_member_proto.Services().ByName("MemberService") == nil {
		t.Fatal("api.open.v1.MemberService descriptor is missing")
	}
	if openv1.File_api_open_v1_account_proto.Services().ByName("AccountService") == nil {
		t.Fatal("api.open.v1.AccountService descriptor is missing")
	}
}

func TestEmailVerificationRemainsAnAccountProofProjection(t *testing.T) {
	email := (&managev1.CanonicalEmailSummary{}).ProtoReflect().Descriptor()
	requireMessageField(t, email, "verified", 2, protoreflect.BoolKind, "")

	member := (&commonv1.MemberSummary{}).ProtoReflect().Descriptor()
	for _, name := range []protoreflect.Name{"email", "email_verified", "role", "status"} {
		if member.Fields().ByName(name) != nil {
			t.Errorf("api.common.v1.MemberSummary.%s must not mix account authority", name)
		}
	}
}

func TestAccountDeletionAndRecoveryContractsRemainOnAccountServices(t *testing.T) {
	manageService := managev1.File_api_manage_v1_account_proto.Services().ByName("AccountService")
	if manageService == nil || manageService.Methods().ByName("RequestAccountDeletion") == nil {
		t.Fatal("api.manage.v1.AccountService.RequestAccountDeletion must remain")
	}

	openService := openv1.File_api_open_v1_account_proto.Services().ByName("AccountService")
	for _, name := range []protoreflect.Name{
		"ConfirmAccountDeletion",
		"CancelAccountDeletion",
		"RequestAccountRecovery",
		"ConfirmAccountRecovery",
	} {
		if openService.Methods().ByName(name) == nil {
			t.Errorf("api.open.v1.AccountService.%s must remain", name)
		}
	}

}
