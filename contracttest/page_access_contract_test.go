package contracttest_test

import (
	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"testing"
)

func TestPageAccessPolicyUsesStableAdditiveFields(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		message proto.Message
		field   string
		tag     protoreflect.FieldNumber
	}{
		{&managev1.Page{}, "access_policy", 17}, {&managev1.PageSummary{}, "access_policy", 10},
		{&managev1.CreatePageRequest{}, "access_policy", 6}, {&managev1.UpdatePageRequest{}, "access_policy", 4},
		{&managev1.UpdatePageResponse{}, "access_policy", 6}, {&openv1.Page{}, "access_policy", 16},
	} {
		descriptor := test.message.ProtoReflect().Descriptor()
		field := requireMessageField(t, descriptor, protoreflect.Name(test.field), test.tag, protoreflect.MessageKind, "api.common.v1.PageAccessPolicy")
		if !field.HasPresence() {
			t.Errorf("%s.%s must preserve absence", descriptor.FullName(), test.field)
		}
		// Every pre-existing field still occupies its original consecutive tag.
		for tag := protoreflect.FieldNumber(1); tag < test.tag; tag++ {
			if descriptor.Fields().ByNumber(tag) == nil {
				t.Errorf("%s lost existing tag %d", descriptor.FullName(), tag)
			}
		}
	}
	response := (&openv1.GetPageResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, response, "page", 1, protoreflect.MessageKind, "api.open.v1.Page")
	requireMessageField(t, response, "block_media", 2, protoreflect.MessageKind, "api.content.v1.ContentBlockMediaItem")
	reason := requireMessageField(t, response, "access_reason", 3, protoreflect.EnumKind, "")
	if reason.Enum().FullName() != "api.common.v1.PageAccessReason" {
		t.Fatal("wrong access reason enum")
	}
}

func TestPageAccessPolicyWireRoundTripAndPatchPresence(t *testing.T) {
	t.Parallel()
	policy := &commonv1.PageAccessPolicy{Mode: commonv1.PageAccessMode_PAGE_ACCESS_MODE_CONDITIONS, AllowedRoles: []policyv1.AuthorizationRole{policyv1.AuthorizationRole_USER, policyv1.AuthorizationRole_AUTHOR, policyv1.AuthorizationRole_ADMIN}, UserTagIds: []string{"tag-1", "tag-2"}, NewsletterSubscriber: true, Match: commonv1.PageAccessMatch_PAGE_ACCESS_MATCH_ALL}
	for _, original := range []*managev1.UpdatePageRequest{{Id: "page-1"}, {Id: "page-1", AccessPolicy: policy}} {
		wire, err := proto.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		decoded := &managev1.UpdatePageRequest{}
		if err := proto.Unmarshal(wire, decoded); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(original, decoded) {
			t.Fatal("Page policy wire roundtrip differs")
		}
	}
	denied := &openv1.GetPageResponse{AccessReason: commonv1.PageAccessReason_PAGE_ACCESS_REASON_AUTHENTICATION_REQUIRED}
	wire, err := proto.Marshal(denied)
	if err != nil {
		t.Fatal(err)
	}
	decoded := &openv1.GetPageResponse{}
	if err := proto.Unmarshal(wire, decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Page != nil || len(decoded.BlockMedia) != 0 || !proto.Equal(denied, decoded) {
		t.Fatal("denied Page response roundtrip differs")
	}
}
