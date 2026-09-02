package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestPostStatusAndScheduleContract(t *testing.T) {
	for name, enum := range map[string]protoreflect.EnumDescriptor{
		"manage": (&managev1.Post{}).ProtoReflect().Descriptor().Fields().ByName("status").Enum(),
		"open":   (&openv1.Post{}).ProtoReflect().Descriptor().Fields().ByName("status").Enum(),
	} {
		value := enum.Values().ByName("POST_STATUS_SCHEDULED")
		if value == nil || value.Number() != 4 {
			t.Errorf("%s PostStatus.SCHEDULED must remain value 4", name)
		}
	}

	managePost := (&managev1.Post{}).ProtoReflect().Descriptor()
	authors := requireMessageField(t, managePost, "author_members", 12, protoreflect.MessageKind, "api.common.v1.MemberSummary")
	if authors.Cardinality() != protoreflect.Repeated {
		t.Error("Post.author_members must be repeated")
	}
	requireOptionalField(t, managePost, "scheduled_at", 20, protoreflect.MessageKind)
	requireOptionalField(t, managePost, "scheduled_time_zone", 21, protoreflect.StringKind)
	actions := requireMessageField(t, managePost, "allowed_actions", 22, protoreflect.EnumKind, "api.manage.v1.PostAction")
	if actions.Cardinality() != protoreflect.Repeated {
		t.Error("Post.allowed_actions must be repeated")
	}

	openPost := (&openv1.Post{}).ProtoReflect().Descriptor()
	requireOptionalField(t, openPost, "scheduled_at", 21, protoreflect.MessageKind)
	requireOptionalField(t, openPost, "scheduled_time_zone", 22, protoreflect.StringKind)
	requireOptionalField(t, openPost, "featured_image_delivery", 23, protoreflect.MessageKind)

	openSummary := (&openv1.PostSummary{}).ProtoReflect().Descriptor()
	requireOptionalField(t, openSummary, "featured_image_delivery", 10, protoreflect.MessageKind)

	requireOptionalField(t, managePost, "featured_image_delivery", 23, protoreflect.MessageKind)

	setFeatured := (&managev1.SetPostFeaturedImageResponse{}).ProtoReflect().Descriptor()
	requireMessageField(
		t,
		setFeatured,
		"image_delivery",
		2,
		protoreflect.MessageKind,
		"api.common.v1.MediaDelivery",
	)
}

func TestPostActionsKeepAuthorMutationAuthorityDistinct(t *testing.T) {
	actions := managev1.File_api_manage_v1_post_proto.Enums().ByName("PostAction")
	for name, number := range map[protoreflect.Name]protoreflect.EnumNumber{
		"POST_ACTION_ADD_AUTHOR":           9,
		"POST_ACTION_REMOVE_AUTHOR":        10,
		"POST_ACTION_MANAGE_COLLABORATORS": 11,
	} {
		value := actions.Values().ByName(name)
		if value == nil || value.Number() != number {
			t.Errorf("PostAction.%s must remain value %d", name, number)
		}
	}

	participant := (&managev1.PostParticipant{}).ProtoReflect().Descriptor()
	requireMessageField(t, participant, "member", 1, protoreflect.MessageKind, "api.common.v1.MemberSummary")
	requireMessageField(t, participant, "role", 2, protoreflect.EnumKind, "api.manage.v1.PostParticipantRole")
	requireMessageField(t, participant, "has_effective_authority", 4, protoreflect.BoolKind, "")

}

func TestPostRPCGatewayTiersPreserveExactAPIAuthority(t *testing.T) {
	service := managev1.File_api_manage_v1_post_proto.Services().ByName("PostService")
	for _, method := range []protoreflect.Name{
		"UpdatePost",
		"SetPostFeaturedImage",
		"DeletePostFeaturedImage",
		"ListPostVersions",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_USER)
	}
	requireMethodTier(t, service, "RemovePostAuthor", policyv1.AuthorizationRole_ADMIN)
	requireMethodTier(t, service, "RepublishPost", policyv1.AuthorizationRole_ADMIN)
	for _, method := range []protoreflect.Name{
		"PublishPost",
		"SchedulePost",
		"CancelPostSchedule",
		"UnpublishPost",
		"ArchivePost",
		"AddPostAuthor",
		"AddPostCollaborator",
		"RemovePostCollaborator",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_AUTHOR)
	}
}

func TestPostCollaborationCarriesOnlyTransientCategoryAndTagSelection(t *testing.T) {
	load := (&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, load, "category_ids", 3, protoreflect.StringKind, "")
	requireMessageField(t, load, "tag_ids", 4, protoreflect.StringKind, "")

	metadata := (&intrav1.UpdatePostDocumentMetadataRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, metadata, "category_ids", 2, protoreflect.MessageKind, "api.intra.v1.PostStringIdList")
	requireMessageField(t, metadata, "tag_ids", 3, protoreflect.MessageKind, "api.intra.v1.PostStringIdList")
	requireMessageField(t, metadata, "expected_revision", 4, protoreflect.StringKind, "")
	requireMessageField(t, metadata, "contributor_member_ids", 5, protoreflect.StringKind, "")

	apply := (&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, apply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")

	update := (&managev1.UpdatePostRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, update, "document_layout", 5, protoreflect.MessageKind, "api.common.v1.DocumentLayout")
}

func TestPostCollaborationCarriesExactLocaleAndTargetCAS(t *testing.T) {
	load := (&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, load, "document", 1, protoreflect.MessageKind, "api.content.v1.LocalizedRichTextDocument")
	requireMessageField(t, load, "locale", 7, protoreflect.StringKind, "")
	requireMessageField(t, load, "locale_exists", 8, protoreflect.BoolKind, "")
	localeMetadata := requireMessageField(t, load, "locale_metadata", 9, protoreflect.MessageKind, "api.intra.v1.PostLocaleMetadata")
	if !localeMetadata.HasPresence() {
		t.Error("Post locale_metadata must preserve missing-target presence")
	}
	requireOptionalField(t, load, "target_revision", 10, protoreflect.StringKind)

	apply := (&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, apply, "expected_target_revision", 4, protoreflect.StringKind)
	requireOptionalField(t, (&intrav1.ApplyPostBlockBatchResponse{}).ProtoReflect().Descriptor(), "target_revision", 6, protoreflect.StringKind)

	metadata := (&intrav1.UpdatePostLocaleMetadataRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, metadata, "expected_target_revision", 7, protoreflect.StringKind)
	requireOptionalField(t, (&intrav1.UpdatePostLocaleMetadataResponse{}).ProtoReflect().Descriptor(), "target_revision", 6, protoreflect.StringKind)

	service := intrav1.File_api_intra_v1_post_proto.Services().ByName("InternalPostService")
	if service.Methods().ByName("UpdatePostLocaleMetadata") == nil {
		t.Fatal("InternalPostService.UpdatePostLocaleMetadata is missing")
	}
	if service.Methods().ByName("UpdatePostSourceMetadata") != nil {
		t.Fatal("InternalPostService.UpdatePostSourceMetadata must be hard-cut")
	}
}

func TestPostTargetCASDoesNotUseAggregateRevision(t *testing.T) {
	source := readSourceProto(t, "api/intra/v1/post.proto")
	load := sourceDeclaration(t, source, "message", "LoadPostBlockDocumentResponse")
	requireSourceContains(t, load,
		"Persisted content_document.revision",
		"shared/source graph CAS",
		"Target-only writes never advance it",
		"derived from document_revision plus that exact locale row's updated_at",
	)
	requireSourceExcludes(t, load, "source_epoch", "snapshot_digest")
	apply := sourceDeclaration(t, source, "message", "ApplyPostBlockBatchRequest")
	requireSourceContains(t, apply,
		"persisted content_document.revision",
		"hard-cut to observe only the shared/source graph",
		"optional string expected_target_revision = 4;",
		"Required for every non-source save",
		"missing target",
		"explicitly create and source-seed it through",
	)
	metadata := sourceDeclaration(t, source, "message", "UpdatePostLocaleMetadataRequest")
	requireSourceContains(t, metadata,
		"Required for every non-source save",
		"rejected until AIDocument explicit create commits",
		"optional string expected_target_revision = 7;",
	)
	ack := sourceDeclaration(t, source, "message", "ApplyPostBlockBatchResponse")
	requireSourceContains(t, ack,
		"remains equal to the request token for a",
		"unrelated target locales never conflict",
	)
}

func TestCollaborationMutationRejectionsAreTyped(t *testing.T) {
	values := intrav1.CollaborationMutationRejectionReason_name
	want := map[int32]string{
		0: "COLLABORATION_MUTATION_REJECTION_REASON_UNSPECIFIED",
		1: "COLLABORATION_MUTATION_REJECTION_REASON_NON_SOURCE_STRUCTURE_FORBIDDEN",
		2: "COLLABORATION_MUTATION_REJECTION_REASON_ROOM_LOCALE_MISMATCH",
		3: "COLLABORATION_MUTATION_REJECTION_REASON_NON_SOURCE_SHARED_FIELD_FORBIDDEN",
		4: "COLLABORATION_MUTATION_REJECTION_REASON_NON_SOURCE_FILE_RELATION_FORBIDDEN",
		5: "COLLABORATION_MUTATION_REJECTION_REASON_NON_SOURCE_DOCUMENT_METADATA_FORBIDDEN",
	}
	if len(values) != len(want) {
		t.Fatalf("CollaborationMutationRejectionReason values = %v", values)
	}
	for number, name := range want {
		if values[number] != name {
			t.Errorf("CollaborationMutationRejectionReason %d = %q, want %q", number, values[number], name)
		}
	}
	detail := (&intrav1.CollaborationMutationRejectionDetail{}).ProtoReflect().Descriptor()
	requireMessageField(t, detail, "reason", 1, protoreflect.EnumKind, "api.intra.v1.CollaborationMutationRejectionReason")
	if got := intrav1.CollaborationConflictReason_name[6]; got != "COLLABORATION_CONFLICT_REASON_TARGET_REVISION_CHANGED" {
		t.Errorf("CollaborationConflictReason 6 = %q", got)
	}
	for _, removed := range []int32{2, 3, 4} {
		if _, exists := intrav1.CollaborationConflictReason_name[removed]; exists {
			t.Errorf("legacy CollaborationConflictReason %d must be reserved", removed)
		}
	}
}

func TestPostSeriesRelationMutatesOnlyThroughSeriesService(t *testing.T) {
	post := (&managev1.Post{}).ProtoReflect().Descriptor()
	requireOptionalField(t, post, "series", 15, protoreflect.MessageKind)
	requireOptionalField(t, post, "series_order", 16, protoreflect.Int32Kind)

	service := managev1.File_api_manage_v1_series_proto.Services().ByName("SeriesService")
	for _, method := range []protoreflect.Name{
		"AssignPostToSeries",
		"UnassignPostFromSeries",
		"ReorderSeriesPosts",
	} {
		if service.Methods().ByName(method) == nil {
			t.Errorf("SeriesService.%s must remain the dedicated Post relation mutation contract", method)
		}
	}
}

func TestShareLinkCarriesBoundedExpiryAndPasswordProof(t *testing.T) {
	item := (&managev1.ShareLinkItem{}).ProtoReflect().Descriptor()
	expiresAt := requireMessageField(t, item, "expires_at", 7, protoreflect.MessageKind, "google.protobuf.Timestamp")
	if expiresAt.HasOptionalKeyword() {
		t.Error("ShareLinkItem.expires_at must always carry the effective bounded expiry")
	}
	requireMessageField(t, item, "has_password", 9, protoreflect.BoolKind, "")

	create := (&managev1.CreateShareLinkRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, create, "expires_at", 4, protoreflect.MessageKind)
	requireOptionalField(t, create, "password", 5, protoreflect.StringKind)

	validate := (&openv1.ValidateShareLinkRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, validate, "password", 2, protoreflect.StringKind)
	response := (&openv1.ValidateShareLinkResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, response, "password_required", 5, protoreflect.BoolKind, "")

	postGet := (&openv1.GetPostRequest{}).ProtoReflect().Descriptor()
	requireOptionalField(t, postGet, "share_password", 3, protoreflect.StringKind)
	managePostGet := (&managev1.GetPostRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, managePostGet, 1)
	requireMessageField(t, managePostGet, "id", 1, protoreflect.StringKind, "")
	for _, name := range []protoreflect.Name{"share_token", "share_password"} {
		if managePostGet.Fields().ByName(name) != nil {
			t.Errorf("manage GetPostRequest must not accept public ShareLink credential %s", name)
		}
	}
	if managePostGet.ReservedNames().Len() != 0 || managePostGet.ReservedRanges().Len() != 0 {
		t.Error("manage GetPostRequest must not retain a compatibility reservation")
	}
	postGetResponse := (&openv1.GetPostResponse{}).ProtoReflect().Descriptor()
	blockMedia := requireMessageField(
		t,
		postGetResponse,
		"block_media",
		2,
		protoreflect.MessageKind,
		"api.content.v1.ContentBlockMediaItem",
	)
	if blockMedia.Cardinality() != protoreflect.Repeated {
		t.Error("GetPostResponse.block_media must preserve every Block selector")
	}
}
