package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestOpenNewsletterServiceIsTokenOnlyUnsubscribe(t *testing.T) {
	service := openv1.File_api_open_v1_newsletter_proto.Services().ByName("NewsletterService")
	if service == nil {
		t.Fatal("api.open.v1.NewsletterService descriptor is missing")
	}
	if got := service.Methods().Len(); got != 1 {
		t.Fatalf("api.open.v1.NewsletterService method count = %d, want 1", got)
	}
	method := service.Methods().ByName("Unsubscribe")
	if method == nil {
		t.Fatal("api.open.v1.NewsletterService.Unsubscribe descriptor is missing")
	}
	request := (&openv1.UnsubscribeNewsletterRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 1)
	token := requireMessageField(t, request, "token", 1, protoreflect.StringKind, "")
	if token.HasPresence() {
		t.Error("UnsubscribeNewsletterRequest.token must be required by the application contract")
	}
}

func TestMemberAggregatesIdentityOwnedNewsletterSubscriptionState(t *testing.T) {
	state := (&managev1.NewsletterSubscriptionState{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, state, 2)
	subscribed := requireMessageField(t, state, "subscribed", 1, protoreflect.BoolKind, "")
	if subscribed.HasPresence() {
		t.Error("NewsletterSubscriptionState.subscribed must be a required current-state boolean")
	}
	subscribedAt := requireMessageField(
		t,
		state,
		"subscribed_at",
		2,
		protoreflect.MessageKind,
		"google.protobuf.Timestamp",
	)
	if !subscribedAt.HasPresence() {
		t.Error("NewsletterSubscriptionState.subscribed_at must preserve absence while unsubscribed")
	}
	for _, forbidden := range []protoreflect.Name{"email", "identity_id", "member_id"} {
		if state.Fields().ByName(forbidden) != nil {
			t.Errorf("NewsletterSubscriptionState.%s must remain outside subscription state", forbidden)
		}
	}

	settings := (&managev1.GetMySettingsResponse{}).ProtoReflect().Descriptor()
	subscription := requireMessageField(
		t,
		settings,
		"newsletter_subscription",
		4,
		protoreflect.MessageKind,
		"api.manage.v1.NewsletterSubscriptionState",
	)
	if !subscription.HasPresence() {
		t.Error("GetMySettingsResponse.newsletter_subscription must be a present screen aggregate")
	}
}

func TestAccountServiceOwnsNewsletterSubscriptionMutations(t *testing.T) {
	memberService := managev1.File_api_manage_v1_member_proto.Services().ByName("MemberService")
	accountService := managev1.File_api_manage_v1_account_proto.Services().ByName("AccountService")
	if memberService == nil || accountService == nil {
		t.Fatal("MemberService and AccountService descriptors are required")
	}
	requireMethodTier(
		t,
		accountService,
		"SetMyNewsletterSubscription",
		policyv1.AuthorizationRole_USER,
	)
	requireMethodTier(
		t,
		accountService,
		"UnsubscribeAccountFromNewsletter",
		policyv1.AuthorizationRole_ADMIN,
	)

	preferences := (&managev1.UpdateMyPreferencesRequest{}).ProtoReflect().Descriptor()
	if preferences.Fields().ByName("newsletter_subscription") != nil {
		t.Error("UpdateMyPreferencesRequest must not mutate Identity-owned newsletter state")
	}

	self := (&managev1.SetMyNewsletterSubscriptionRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, self, 1)
	selfSubscribed := requireMessageField(t, self, "subscribed", 1, protoreflect.BoolKind, "")
	if selfSubscribed.HasPresence() {
		t.Error("SetMyNewsletterSubscriptionRequest.subscribed must be explicit")
	}
	for _, forbidden := range []protoreflect.Name{"email", "identity_id", "member_id"} {
		if self.Fields().ByName(forbidden) != nil {
			t.Errorf("SetMyNewsletterSubscriptionRequest.%s must derive current identity from the session", forbidden)
		}
	}

	admin := (&managev1.UnsubscribeAccountFromNewsletterRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, admin, 1)
	requireMessageField(t, admin, "member_id", 1, protoreflect.StringKind, "")
	if admin.Fields().ByName("subscribed") != nil {
		t.Error("admin newsletter mutation must not expose opt-in")
	}

	response := (&managev1.NewsletterSubscriptionMutationResponse{}).ProtoReflect().Descriptor()
	requireMessageField(
		t,
		response,
		"newsletter_subscription",
		1,
		protoreflect.MessageKind,
		"api.manage.v1.NewsletterSubscriptionState",
	)
}
