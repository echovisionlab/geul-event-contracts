package contracttest_test

import (
	"testing"

	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMemberSummaryIsExactBoundedProjection(t *testing.T) {
	summary := (&commonv1.MemberSummary{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, summary, 4)
	requireMessageField(t, summary, "id", 1, protoreflect.StringKind, "")
	nickname := requireMessageField(t, summary, "nickname", 2, protoreflect.StringKind, "")
	if nickname.HasPresence() {
		t.Error("MemberSummary.nickname must be required; pre-onboarding uses the Member UUID placeholder")
	}
	avatar := requireMessageField(t, summary, "avatar_asset", 3, protoreflect.MessageKind, "api.common.v1.AssetRef")
	if !avatar.HasPresence() {
		t.Error("MemberSummary.avatar_asset must preserve absence")
	}
	requireMessageField(t, summary, "deleted", 4, protoreflect.BoolKind, "")
	for _, forbidden := range []protoreflect.Name{
		"display_name",
		"email",
		"email_verified",
		"identity_id",
		"provider",
		"role",
		"status",
		"ban_reason",
		"onboarded",
	} {
		if summary.Fields().ByName(forbidden) != nil {
			t.Errorf("MemberSummary.%s must not mix account authority", forbidden)
		}
	}
}

func TestPublicAuthorSummaryUsesANewMemberSemanticType(t *testing.T) {
	response := (&openv1.ListAuthorsResponse{}).ProtoReflect().Descriptor()
	authors := requireMessageField(t, response, "authors", 1, protoreflect.MessageKind, "api.open.v1.PublicAuthorSummary")
	if authors.Cardinality() != protoreflect.Repeated {
		t.Error("ListAuthorsResponse.authors must be repeated")
	}

	author := (&openv1.PublicAuthorSummary{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, author, 3)
	requireMessageField(t, author, "member", 1, protoreflect.MessageKind, "api.common.v1.MemberSummary")
	requireMessageField(t, author, "post_count", 2, protoreflect.Int32Kind, "")
	bio := requireMessageField(t, author, "bio", 3, protoreflect.StringKind, "")
	if !bio.HasPresence() {
		t.Error("PublicAuthorSummary.bio must preserve absence")
	}
}

func TestMemberAndAccountServicesOwnDisjointMutations(t *testing.T) {
	member := managev1.File_api_manage_v1_member_proto.Services().ByName("MemberService")
	account := managev1.File_api_manage_v1_account_proto.Services().ByName("AccountService")
	if member == nil || account == nil {
		t.Fatal("MemberService and AccountService must both exist")
	}

	for _, method := range []protoreflect.Name{
		"GetCurrentSession",
		"GetMyProfile",
		"GetMySections",
		"GetMySettings",
		"CheckNicknameAvailability",
		"CompleteMyOnboarding",
		"UpdateMyProfile",
		"UpdateMyPreferences",
		"SetMyAvatar",
		"ListMembersAdmin",
		"SearchMembers",
		"SetMemberTags",
	} {
		if member.Methods().ByName(method) == nil {
			t.Errorf("MemberService.%s is missing", method)
		}
	}

	for _, forbidden := range []protoreflect.Name{
		"GetMySecurity",
		"RequestEmailChange",
		"SetAccountRole",
		"BanAccount",
		"DeleteAccount",
	} {
		if member.Methods().ByName(forbidden) != nil {
			t.Errorf("MemberService.%s must remain AccountService-owned", forbidden)
		}
	}

	for _, method := range []protoreflect.Name{
		"GetMySecurity",
		"SetMyCanonicalEmail",
		"RevokeMySession",
		"RequestEmailChange",
		"RequestAccountDeletion",
		"SetAccountRole",
		"BanAccount",
		"DeleteAccount",
	} {
		if account.Methods().ByName(method) == nil {
			t.Errorf("AccountService.%s is missing", method)
		}
	}
	for _, forbidden := range []protoreflect.Name{"RemoveMyEmailCodeCredential", "RequestEmailAddition"} {
		if account.Methods().ByName(forbidden) != nil {
			t.Errorf("AccountService.%s must not expose a second Email Code identifier lifecycle", forbidden)
		}
	}
	for _, forbidden := range []protoreflect.Name{"UpdateMyProfile", "UpdateMyPreferences", "SetMyAvatar"} {
		if account.Methods().ByName(forbidden) != nil {
			t.Errorf("AccountService.%s must remain MemberService-owned", forbidden)
		}
	}
	for _, method := range []protoreflect.Name{"SetMyNewsletterSubscription", "UnsubscribeAccountFromNewsletter"} {
		if account.Methods().ByName(method) == nil {
			t.Errorf("AccountService.%s is missing", method)
		}
		if member.Methods().ByName(method) != nil {
			t.Errorf("MemberService.%s must not own Identity-scoped newsletter mutation", method)
		}
	}
	requireMethodTier(t, member, "GetCurrentSession", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, member, "GetMyProfile", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, member, "GetMySections", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, member, "CheckNicknameAvailability", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, member, "CompleteMyOnboarding", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, account, "GetMySecurity", policyv1.AuthorizationRole_USER)
	requireMethodTier(t, account, "SetAccountRole", policyv1.AuthorizationRole_ADMIN)

	openMember := openv1.File_api_open_v1_member_proto.Services().ByName("MemberService")
	openAccount := openv1.File_api_open_v1_account_proto.Services().ByName("AccountService")
	if openMember.Methods().ByName("GetPublicMember") == nil || openMember.Methods().ByName("ListAuthors") == nil {
		t.Error("public MemberService must own public profile and author reads")
	}
	for _, method := range []protoreflect.Name{"ConfirmAccountDeletion", "CancelAccountDeletion", "RequestAccountRecovery", "ConfirmAccountRecovery"} {
		if openAccount.Methods().ByName(method) == nil {
			t.Errorf("public AccountService.%s is missing", method)
		}
		if openMember.Methods().ByName(method) != nil {
			t.Errorf("public MemberService.%s must not own account lifecycle", method)
		}
	}
}

func TestMemberNicknameHardCutAndOnboardingContract(t *testing.T) {
	session := (&managev1.GetCurrentSessionResponse{}).ProtoReflect().Descriptor()
	suggestion := requireMessageField(t, session, "nickname_suggestion", 4, protoreflect.StringKind, "")
	if !suggestion.HasPresence() {
		t.Error("GetCurrentSessionResponse.nickname_suggestion must preserve absence")
	}
	requireMessageField(t, session, "onboarded", 3, protoreflect.BoolKind, "")

	checkRequest := (&managev1.CheckNicknameAvailabilityRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, checkRequest, 1)
	requireMessageField(t, checkRequest, "nickname", 1, protoreflect.StringKind, "")
	checkResponse := (&managev1.CheckNicknameAvailabilityResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, checkResponse, 1)
	requireMessageField(t, checkResponse, "available", 1, protoreflect.BoolKind, "")

	completeRequest := (&managev1.CompleteMyOnboardingRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, completeRequest, 1)
	requireMessageField(t, completeRequest, "nickname", 1, protoreflect.StringKind, "")
	completeResponse := (&managev1.CompleteMyOnboardingResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, completeResponse, 2)
	requireMessageField(t, completeResponse, "member", 1, protoreflect.MessageKind, "api.common.v1.MemberSummary")
	requireMessageField(t, completeResponse, "onboarded", 2, protoreflect.BoolKind, "")

	myUpdate := (&managev1.UpdateMyProfileRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, myUpdate, "nickname", 1, protoreflect.StringKind, "")
	if myUpdate.Fields().ByName("onboarded") != nil {
		t.Error("UpdateMyProfileRequest must not mutate onboarding state")
	}
	adminUpdate := (&managev1.UpdateMemberProfileRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, adminUpdate, "nickname", 2, protoreflect.StringKind, "")

	for _, descriptor := range map[string]protoreflect.MessageDescriptor{
		"VersionContributor": (&managev1.VersionContributor{}).ProtoReflect().Descriptor(),
		"SeriesManager":      (&managev1.SeriesManager{}).ProtoReflect().Descriptor(),
	} {
		requireMessageField(t, descriptor, "nickname", 2, protoreflect.StringKind, "")
	}
}

func TestAccountSecurityExposesCanonicalEmailWithoutDeliveryAlias(t *testing.T) {
	security := (&managev1.AccountSecurity{}).ProtoReflect().Descriptor()
	requireMessageField(t, security, "email_candidates", 2, protoreflect.MessageKind, "api.manage.v1.AccountEmailCandidate")
	requireMessageField(t, security, "canonical_email", 6, protoreflect.StringKind, "")

	candidate := (&managev1.AccountEmailCandidate{}).ProtoReflect().Descriptor()
	requireMessageField(t, candidate, "current", 3, protoreflect.BoolKind, "")
}

func TestDeletionCommandsReserveIdentitySemanticUserID(t *testing.T) {
	identity := (&managev1.UserDeleteIdentityCommand{}).ProtoReflect().Descriptor()
	memberID := requireMessageField(t, identity, "member_id", 6, protoreflect.StringKind, "")
	identityID := requireMessageField(t, identity, "identity_id", 7, protoreflect.StringKind, "")
	if memberID.HasPresence() || identityID.HasPresence() {
		t.Error("identity deletion requires both member_id and identity_id")
	}

	avatar := (&managev1.UserDeleteAvatarCommand{}).ProtoReflect().Descriptor()
	requireMessageField(t, avatar, "member_id", 2, protoreflect.StringKind, "")
}

func TestUserIDAttributionUsesNewMemberFields(t *testing.T) {
	tests := []struct {
		descriptor protoreflect.MessageDescriptor
		memberName protoreflect.Name
		memberNo   protoreflect.FieldNumber
	}{
		{(&managev1.MetadataGenerationJob{}).ProtoReflect().Descriptor(), "requester_member_id", 3},
		{(&managev1.CommentWithAuthor{}).ProtoReflect().Descriptor(), "member_id", 8},
		{(&managev1.CommentNode{}).ProtoReflect().Descriptor(), "member_id", 11},
		{(&managev1.FormSubmission{}).ProtoReflect().Descriptor(), "member_id", 3},
		{(&managev1.SeriesManager{}).ProtoReflect().Descriptor(), "member_id", 1},
		{(&managev1.AddSeriesManagerRequest{}).ProtoReflect().Descriptor(), "member_id", 2},
		{(&managev1.RemoveSeriesManagerRequest{}).ProtoReflect().Descriptor(), "member_id", 2},
		{(&managev1.AddWorkCreditRequest{}).ProtoReflect().Descriptor(), "member_id", 4},
		{(&managev1.ProgramEventCredit{}).ProtoReflect().Descriptor(), "member_id", 8},
		{(&managev1.AddProgramEventCreditRequest{}).ProtoReflect().Descriptor(), "member_id", 3},
		{(&managev1.ReleaseCreditInput{}).ProtoReflect().Descriptor(), "member_id", 3},
		{(&managev1.TrackCredit{}).ProtoReflect().Descriptor(), "member_id", 6},
		{(&managev1.TrackCreditInput{}).ProtoReflect().Descriptor(), "member_id", 3},
		{(&openv1.ReleaseCredit{}).ProtoReflect().Descriptor(), "member_id", 6},
	}
	for _, contract := range tests {
		requireMessageField(t, contract.descriptor, contract.memberName, contract.memberNo, protoreflect.StringKind, "")
	}
}

func TestResourceMemberNamingContract(t *testing.T) {
	postFile := managev1.File_api_manage_v1_post_proto
	postService := postFile.Services().ByName("PostService")
	for _, name := range []protoreflect.Name{
		"ListPostParticipants",
		"AddPostAuthor",
		"RemovePostAuthor",
		"AddPostCollaborator",
		"RemovePostCollaborator",
	} {
		if postService.Methods().ByName(name) == nil {
			t.Errorf("PostService.%s is missing", name)
		}
	}

	seriesFile := managev1.File_api_manage_v1_series_proto
	seriesService := seriesFile.Services().ByName("SeriesService")
	for _, name := range []protoreflect.Name{"ListSeriesManagers", "AddSeriesManager", "RemoveSeriesManager"} {
		if seriesService.Methods().ByName(name) == nil {
			t.Errorf("SeriesService.%s is missing", name)
		}
	}
}

func TestInternalCollaborationAuthorizationUsesCanonicalTypedPrincipal(t *testing.T) {
	request := (&intrav1.AuthorizeCollaborationRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 3)
	for _, name := range []protoreflect.Name{"session_id", "issuer", "subject", "identity_id", "member_id", "resource_type", "resource_id", "object", "relation", "tuple"} {
		if request.Fields().ByName(name) != nil {
			t.Errorf("AuthorizeCollaborationRequest.%s must not be a raw top-level authorization field", name)
		}
	}
	requireMessageField(t, request, "principal", 1, protoreflect.MessageKind, "api.intra.v1.CollaborationPrincipal")
	requireMessageField(t, request, "resource", 2, protoreflect.MessageKind, "api.intra.v1.CollaborationResource")
	requireMessageField(t, request, "permission", 3, protoreflect.EnumKind, "")

	principal := (&intrav1.CollaborationPrincipal{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, principal, 1)
	requireMessageField(t, principal, "session_id", 1, protoreflect.StringKind, "")
	for _, name := range []protoreflect.Name{"issuer", "subject", "identity_id", "account_identity_id", "member_id"} {
		if principal.Fields().ByName(name) != nil {
			t.Errorf("CollaborationPrincipal.%s must not remain active", name)
		}
	}

	resource := (&intrav1.CollaborationResource{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, resource, 3)
	requireMessageField(t, resource, "type", 1, protoreflect.EnumKind, "")
	requireMessageField(t, resource, "id", 2, protoreflect.StringKind, "")
	locale := requireMessageField(t, resource, "locale", 3, protoreflect.StringKind, "")
	if locale.HasPresence() || locale.Cardinality() == protoreflect.Repeated {
		t.Error("CollaborationResource.locale must be one canonical required scalar")
	}
	for _, name := range []protoreflect.Name{"object", "relation", "subject", "tuple"} {
		if resource.Fields().ByName(name) != nil {
			t.Errorf("CollaborationResource.%s must not remain active", name)
		}
	}

	permission := request.Fields().ByName("permission").Enum()
	if permission.Values().Len() != 4 {
		t.Fatalf("CollaborationPermission must expose only unspecified, edit, view, and manage, got %d values", permission.Values().Len())
	}
	for _, name := range []protoreflect.Name{
		"COLLABORATION_PERMISSION_UNSPECIFIED",
		"COLLABORATION_PERMISSION_EDIT",
		"COLLABORATION_PERMISSION_VIEW",
		"COLLABORATION_PERMISSION_MANAGE",
	} {
		if permission.Values().ByName(name) == nil {
			t.Errorf("CollaborationPermission.%s is missing", name)
		}
	}

	response := (&intrav1.AuthorizeCollaborationResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, response, 4)
	requireMessageField(t, response, "authorized", 1, protoreflect.BoolKind, "")
	member := requireMessageField(t, response, "member", 2, protoreflect.MessageKind, "api.common.v1.MemberSummary")
	if !member.HasPresence() {
		t.Error("AuthorizeCollaborationResponse.member must preserve denial absence")
	}
	requireMessageField(t, response, "denial_reason", 3, protoreflect.EnumKind, "")
	actualLocale := requireMessageField(t, response, "locale", 4, protoreflect.StringKind, "")
	if actualLocale.HasPresence() || actualLocale.Cardinality() == protoreflect.Repeated {
		t.Error("AuthorizeCollaborationResponse.locale must be one canonical required scalar")
	}
	denialReason := response.Fields().ByName("denial_reason").Enum()
	for _, name := range []protoreflect.Name{
		"COLLABORATION_AUTHORIZATION_DENIAL_REASON_UNSPECIFIED",
		"COLLABORATION_AUTHORIZATION_DENIAL_REASON_SESSION_INVALID",
		"COLLABORATION_AUTHORIZATION_DENIAL_REASON_PERMISSION_DENIED",
	} {
		if denialReason.Values().ByName(name) == nil {
			t.Errorf("CollaborationAuthorizationDenialReason.%s is missing", name)
		}
	}

	service := intrav1.File_api_intra_v1_collaboration_proto.Services().ByName("InternalCollaborationAuthorizationService")
	if service == nil || service.Methods().ByName("AuthorizeCollaboration") == nil {
		t.Fatal("InternalCollaborationAuthorizationService.AuthorizeCollaboration is missing")
	}
}

func TestNewsletterAndCampaignRecipientsUseMemberAttribution(t *testing.T) {
	newsletter := (&managev1.NewsletterSubscriptionRecipient{}).ProtoReflect().Descriptor()
	requireMessageField(t, newsletter, "identity_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, newsletter, "member_id", 2, protoreflect.StringKind, "")

	campaign := (&managev1.CampaignRecipient{}).ProtoReflect().Descriptor()
	requireMessageField(t, campaign, "member_id", 6, protoreflect.StringKind, "")
}

func TestAdminMemberKeepsBoundedMemberAndAccountDetailOwnership(t *testing.T) {
	admin := (&managev1.AdminMember{}).ProtoReflect().Descriptor()
	requireMessageField(t, admin, "member", 1, protoreflect.MessageKind, "api.manage.v1.MemberProfile")
	requireMessageField(t, admin, "account", 2, protoreflect.MessageKind, "api.manage.v1.AccountSummary")
	requireMessageField(t, admin, "tag_ids", 3, protoreflect.StringKind, "")
	details := requireMessageField(t, admin, "account_details", 4, protoreflect.MessageKind, "api.manage.v1.AccountAdminDetails")
	if !details.HasPresence() {
		t.Error("AdminMember.account_details must be omittable from list rows")
	}
	requireMessageField(t, admin, "newsletter_subscription", 5, protoreflect.MessageKind, "api.manage.v1.NewsletterSubscriptionState")
	requireMessageField(t, admin, "onboarded", 6, protoreflect.BoolKind, "")

	accountDetails := (&managev1.AccountAdminDetails{}).ProtoReflect().Descriptor()
	requireMessageField(t, accountDetails, "providers", 1, protoreflect.MessageKind, "api.manage.v1.AccountProvider")
	requireMessageField(t, accountDetails, "email_candidates", 2, protoreflect.MessageKind, "api.manage.v1.AccountEmailCandidate")
	if accountDetails.Fields().ByName("sessions") != nil || accountDetails.Fields().ByName("credentials") != nil {
		t.Error("AccountAdminDetails must not expose sessions or credentials")
	}
}

func TestTranslationAttributionUsesRequesterMemberOnly(t *testing.T) {
	job := (&managev1.TranslationJob{}).ProtoReflect().Descriptor()
	requester := requireMessageField(t, job, "requested_by_member_id", 12, protoreflect.StringKind, "")
	if requester.HasPresence() {
		t.Fatal("TranslationJob.requested_by_member_id is an in-flight invariant, not optional state")
	}

	entry := (&managev1.TranslationEntry{}).ProtoReflect().Descriptor()
	if entry.Fields().ByName("edited_by_member_id") != nil {
		t.Fatal("read-only machine projection must not expose manual editor attribution")
	}
}
