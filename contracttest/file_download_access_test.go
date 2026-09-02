package contracttest_test

import (
	"strings"
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestFileDownloadAudienceContract(t *testing.T) {
	wants := map[managev1.FileDownloadAudience]int32{
		managev1.FileDownloadAudience_FILE_DOWNLOAD_AUDIENCE_UNSPECIFIED:   0,
		managev1.FileDownloadAudience_FILE_DOWNLOAD_AUDIENCE_DISABLED:      1,
		managev1.FileDownloadAudience_FILE_DOWNLOAD_AUDIENCE_PUBLIC:        2,
		managev1.FileDownloadAudience_FILE_DOWNLOAD_AUDIENCE_AUTHENTICATED: 3,
		managev1.FileDownloadAudience_FILE_DOWNLOAD_AUDIENCE_RESTRICTED:    4,
	}
	for audience, want := range wants {
		if got := int32(audience); got != want {
			t.Errorf("%s = %d, want %d", audience, got, want)
		}
	}
}

func TestManageFileDownloadPolicyTargetsExactRelation(t *testing.T) {
	service := managev1.File_api_manage_v1_file_proto.Services().ByName("FileService")
	if service == nil {
		t.Fatal("api.manage.v1.FileService descriptor is missing")
	}

	for _, name := range []protoreflect.Name{
		"GetFileDownloadPolicy",
		"UpdateFileDownloadPolicy",
	} {
		requireMethodTier(t, service, name, policyv1.AuthorizationRole_AUTHOR)
	}

	getPolicy := (&managev1.GetFileDownloadPolicyRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, getPolicy, 4)
	requireMessageField(t, getPolicy, "entity_type", 1, protoreflect.EnumKind, "api.manage.v1.TranscodeEntityType")
	requireMessageField(t, getPolicy, "entity_id", 2, protoreflect.StringKind, "")
	getBlockID := requireMessageField(t, getPolicy, "block_id", 3, protoreflect.StringKind, "")
	if !getBlockID.HasOptionalKeyword() {
		t.Error("GetFileDownloadPolicyRequest.block_id must be optional for Track targets")
	}
	getReferencePath := requireMessageField(t, getPolicy, "reference_path", 4, protoreflect.StringKind, "")
	if !getReferencePath.HasOptionalKeyword() {
		t.Error("GetFileDownloadPolicyRequest.reference_path must be optional for Track targets")
	}
	if getPolicy.Fields().ByName("file_id") != nil {
		t.Error("GetFileDownloadPolicyRequest must derive the current File from the exact relation")
	}

	updatePolicy := (&managev1.UpdateFileDownloadPolicyRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, updatePolicy, 7)
	requireMessageField(t, updatePolicy, "entity_type", 1, protoreflect.EnumKind, "api.manage.v1.TranscodeEntityType")
	requireMessageField(t, updatePolicy, "entity_id", 2, protoreflect.StringKind, "")
	updateBlockID := requireMessageField(t, updatePolicy, "block_id", 3, protoreflect.StringKind, "")
	if !updateBlockID.HasOptionalKeyword() {
		t.Error("UpdateFileDownloadPolicyRequest.block_id must be optional for Track targets")
	}
	updateReferencePath := requireMessageField(t, updatePolicy, "reference_path", 4, protoreflect.StringKind, "")
	if !updateReferencePath.HasOptionalKeyword() {
		t.Error("UpdateFileDownloadPolicyRequest.reference_path must be optional for Track targets")
	}
	expectedFileID := requireMessageField(t, updatePolicy, "expected_file_id", 5, protoreflect.StringKind, "")
	if expectedFileID.HasOptionalKeyword() {
		t.Error("UpdateFileDownloadPolicyRequest.expected_file_id must be a required CAS input")
	}
	requireMessageField(t, updatePolicy, "audience", 6, protoreflect.EnumKind, "api.manage.v1.FileDownloadAudience")
	segmentIDs := requireMessageField(t, updatePolicy, "audience_segment_ids", 7, protoreflect.StringKind, "")
	if !segmentIDs.IsList() {
		t.Error("UpdateFileDownloadPolicyRequest.audience_segment_ids must be repeated")
	}
	if updatePolicy.Fields().ByName("file_id") != nil {
		t.Error("UpdateFileDownloadPolicyRequest must use expected_file_id only as a CAS guard")
	}

	policy := (&managev1.FileDownloadPolicy{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, policy, 7)
	requireMessageField(t, policy, "entity_type", 1, protoreflect.EnumKind, "api.manage.v1.TranscodeEntityType")
	requireMessageField(t, policy, "entity_id", 2, protoreflect.StringKind, "")
	policyBlockID := requireMessageField(t, policy, "block_id", 3, protoreflect.StringKind, "")
	if !policyBlockID.HasOptionalKeyword() {
		t.Error("FileDownloadPolicy.block_id must be optional for Track targets")
	}
	policyReferencePath := requireMessageField(t, policy, "reference_path", 4, protoreflect.StringKind, "")
	if !policyReferencePath.HasOptionalKeyword() {
		t.Error("FileDownloadPolicy.reference_path must be optional for Track targets")
	}
	requireMessageField(t, policy, "file_id", 5, protoreflect.StringKind, "")
	requireMessageField(t, policy, "audience", 6, protoreflect.EnumKind, "api.manage.v1.FileDownloadAudience")
	segments := requireMessageField(
		t,
		policy,
		"audience_segments",
		7,
		protoreflect.MessageKind,
		"api.manage.v1.AudienceSegmentSummary",
	)
	if !segments.IsList() {
		t.Error("FileDownloadPolicy.audience_segments must be repeated")
	}

	for _, message := range []protoreflect.MessageDescriptor{getPolicy, updatePolicy, policy} {
		if message.ReservedNames().Len() != 0 || message.ReservedRanges().Len() != 0 {
			t.Errorf("%s must not retain a reserved compatibility surface", message.FullName())
		}
	}

	fileManagerFile := (&managev1.FileManagerFile{}).ProtoReflect().Descriptor()
	if fileManagerFile.Fields().ByName("download_audience") != nil {
		t.Error("FileManagerFile must not expose relation-owned download policy")
	}
	requireMessageField(t, fileManagerFile, "delivery", 11, protoreflect.MessageKind, "api.common.v1.MediaDelivery")
	requireMessageField(t, fileManagerFile, "usage_count", 12, protoreflect.Int32Kind, "")
	if fileManagerFile.ReservedNames().Len() != 0 || fileManagerFile.ReservedRanges().Len() != 0 {
		t.Error("FileManagerFile must not reserve the removed global policy field")
	}
}

func TestAudienceSegmentAuthenticatedAccessContract(t *testing.T) {
	summary := (&managev1.AudienceSegmentSummary{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, summary, 4)
	requireMessageField(t, summary, "id", 1, protoreflect.StringKind, "")
	requireMessageField(t, summary, "name", 2, protoreflect.StringKind, "")
	description := requireMessageField(t, summary, "description", 3, protoreflect.StringKind, "")
	if !description.HasOptionalKeyword() {
		t.Error("AudienceSegmentSummary.description must be optional")
	}
	requireMessageField(t, summary, "segment_type", 4, protoreflect.EnumKind, "api.manage.v1.SegmentType")

	request := (&managev1.ListSegmentsForAuthenticatedAccessRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 3)
	requireMessageField(t, request, "pagination", 1, protoreflect.MessageKind, "api.common.v1.PaginationRequest")
	filters := requireMessageField(
		t,
		request,
		"filters",
		2,
		protoreflect.MessageKind,
		"api.common.v1.FilterSpec",
	)
	if !filters.IsList() {
		t.Error("ListSegmentsForAuthenticatedAccessRequest.filters must be repeated")
	}
	sorts := requireMessageField(
		t,
		request,
		"sorts",
		3,
		protoreflect.MessageKind,
		"api.common.v1.SortSpec",
	)
	if !sorts.IsList() {
		t.Error("ListSegmentsForAuthenticatedAccessRequest.sorts must be repeated")
	}

	response := (&managev1.ListSegmentsForAuthenticatedAccessResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, response, 2)
	responseSegments := requireMessageField(
		t,
		response,
		"segments",
		1,
		protoreflect.MessageKind,
		"api.manage.v1.AudienceSegmentSummary",
	)
	if !responseSegments.IsList() {
		t.Error("ListSegmentsForAuthenticatedAccessResponse.segments must be repeated")
	}
	requireMessageField(t, response, "pagination", 2, protoreflect.MessageKind, "api.common.v1.PaginationResponse")

	service := managev1.File_api_manage_v1_audience_proto.Services().ByName("AudienceService")
	if service == nil {
		t.Fatal("api.manage.v1.AudienceService descriptor is missing")
	}
	requireMethodTier(
		t,
		service,
		"ListSegmentsForAuthenticatedAccess",
		policyv1.AuthorizationRole_AUTHOR,
	)
}

func TestSetMemberTagsContract(t *testing.T) {
	request := (&managev1.SetMemberTagsRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 2)
	requireMessageField(t, request, "member_id", 1, protoreflect.StringKind, "")
	tagIDs := requireMessageField(t, request, "tag_ids", 2, protoreflect.StringKind, "")
	if !tagIDs.IsList() {
		t.Error("SetMemberTagsRequest.tag_ids must be repeated")
	}

	service := managev1.File_api_manage_v1_member_proto.Services().ByName("MemberService")
	if service == nil {
		t.Fatal("api.manage.v1.MemberService descriptor is missing")
	}
	requireMethodTier(t, service, "SetMemberTags", policyv1.AuthorizationRole_ADMIN)
	method := service.Methods().ByName("SetMemberTags")
	if got := method.Output().FullName(); got != "api.manage.v1.AdminMember" {
		t.Errorf("SetMemberTags output = %s, want api.manage.v1.AdminMember", got)
	}
}

func TestFileDownloadManagementContractsDoNotExposeIdentityMetadata(t *testing.T) {
	for _, message := range []protoreflect.MessageDescriptor{
		(&managev1.AudienceSegmentSummary{}).ProtoReflect().Descriptor(),
		(&managev1.FileDownloadPolicy{}).ProtoReflect().Descriptor(),
		(&openv1.AuthorizeDownloadResponse{}).ProtoReflect().Descriptor(),
		(&openv1.FileDownloadAccess{}).ProtoReflect().Descriptor(),
	} {
		requireNoIdentityDisplayMetadata(t, message)
	}
}

func TestFileDownloadPolicyUsesOnlyAudienceSegmentSummary(t *testing.T) {
	policy := (&managev1.FileDownloadPolicy{}).ProtoReflect().Descriptor()
	segments := requireMessageField(
		t,
		policy,
		"audience_segments",
		7,
		protoreflect.MessageKind,
		"api.manage.v1.AudienceSegmentSummary",
	)
	if !segments.IsList() {
		t.Error("FileDownloadPolicy.audience_segments must be repeated")
	}
}

func TestOpenAuthorizeDownloadIsScopedAndMinimal(t *testing.T) {
	if got := int32(openv1.PublicMediaEntityType_PUBLIC_MEDIA_ENTITY_TYPE_RELEASE); got != 5 {
		t.Fatalf("PUBLIC_MEDIA_ENTITY_TYPE_RELEASE = %d, want 5", got)
	}

	service := openv1.File_api_open_v1_file_proto.Services().ByName("FileService")
	if service == nil || service.Methods().ByName("AuthorizeDownload") == nil {
		t.Fatal("api.open.v1.FileService.AuthorizeDownload descriptor is missing")
	}

	request := (&openv1.AuthorizeDownloadRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, request, "entity_type", 1, protoreflect.EnumKind, "api.open.v1.PublicMediaEntityType")
	requireMessageField(t, request, "entity_id", 2, protoreflect.StringKind, "")
	contentBlock := requireMessageField(t, request, "content_block", 3, protoreflect.MessageKind, "api.content.v1.ContentBlockMediaSelector")
	trackID := requireMessageField(t, request, "track_id", 4, protoreflect.StringKind, "")
	if request.Fields().Len() != 4 {
		t.Fatalf("AuthorizeDownloadRequest must expose exactly owner + relation target, got %d fields", request.Fields().Len())
	}
	if request.Fields().ByName("file_id") != nil || request.Fields().ByName("share_token") != nil {
		t.Error("AuthorizeDownloadRequest must derive File from the exact relation without ShareLink proof")
	}
	if request.Oneofs().Len() != 1 || request.Oneofs().Get(0).Name() != "relation_target" {
		t.Fatal("AuthorizeDownloadRequest must expose one exact relation_target")
	}
	for _, field := range []protoreflect.FieldDescriptor{contentBlock, trackID} {
		if field.ContainingOneof() == nil || field.ContainingOneof().Name() != "relation_target" {
			t.Errorf("AuthorizeDownloadRequest.%s must belong to relation_target", field.Name())
		}
	}
	if request.ReservedNames().Len() != 0 || request.ReservedRanges().Len() != 0 {
		t.Error("AuthorizeDownloadRequest must not retain a compatibility reservation")
	}

	response := (&openv1.AuthorizeDownloadResponse{}).ProtoReflect().Descriptor()
	if response.Fields().Len() != 2 {
		t.Fatalf("AuthorizeDownloadResponse must expose exactly access + download, got %d fields", response.Fields().Len())
	}
	requireMessageField(t, response, "access", 1, protoreflect.MessageKind, "api.open.v1.FileDownloadAccess")
	download := requireMessageField(t, response, "download", 2, protoreflect.MessageKind, "api.common.v1.ExpiringMediaRef")
	if !download.HasOptionalKeyword() {
		t.Error("AuthorizeDownloadResponse.download must be optional")
	}
	for index := 0; index < response.Fields().Len(); index++ {
		fieldName := string(response.Fields().Get(index).Name())
		if strings.Contains(fieldName, "group") || strings.Contains(fieldName, "segment") {
			t.Errorf("public authorization response must not expose audience metadata")
		}
	}
}

func TestPublicBulkMediaDeliveryIsHardCut(t *testing.T) {
	file := openv1.File_api_open_v1_file_proto
	service := file.Services().ByName("FileService")
	if service == nil {
		t.Fatal("api.open.v1.FileService descriptor is missing")
	}
	if service.Methods().ByName("GetBulkMediaDeliveries") != nil {
		t.Error("api.open.v1.FileService.GetBulkMediaDeliveries must be removed")
	}
	if service.Methods().Len() != 1 || service.Methods().ByName("AuthorizeDownload") == nil {
		t.Error("api.open.v1.FileService must expose only exact-relation AuthorizeDownload")
	}
	for _, name := range []protoreflect.Name{
		"PublicMediaFileRequest",
		"GetBulkMediaDeliveriesRequest",
		"GetBulkMediaDeliveriesResponse",
	} {
		if file.Messages().ByName(name) != nil {
			t.Errorf("api.open.v1.%s must be removed with the transitional Bulk RPC", name)
		}
	}

	effective := (&openv1.FileDownloadAccess{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, effective, 2)
	requireMessageField(t, effective, "availability", 1, protoreflect.EnumKind, "api.open.v1.FileDownloadAvailability")
	requireMessageField(t, effective, "action", 2, protoreflect.EnumKind, "api.open.v1.FileDownloadAction")
	if effective.ReservedNames().Len() != 0 || effective.ReservedRanges().Len() != 0 {
		t.Error("FileDownloadAccess must not retain a compatibility reservation")
	}
}

func TestReleaseTrackUsesEffectiveDownloadState(t *testing.T) {
	releaseTrack := (&openv1.ReleaseTrack{}).ProtoReflect().Descriptor()
	requireMessageField(t, releaseTrack, "download_availability", 8, protoreflect.EnumKind, "api.open.v1.FileDownloadAvailability")
	requireMessageField(t, releaseTrack, "download_action", 9, protoreflect.EnumKind, "api.open.v1.FileDownloadAction")
}

func requireMethodTier(
	t *testing.T,
	service protoreflect.ServiceDescriptor,
	name protoreflect.Name,
	want policyv1.AuthorizationRole,
) {
	t.Helper()
	method := service.Methods().ByName(name)
	if method == nil {
		t.Fatalf("%s.%s descriptor is missing", service.FullName(), name)
	}
	options, ok := method.Options().(*descriptorpb.MethodOptions)
	if !ok {
		t.Fatalf("%s.%s has invalid method options", service.FullName(), name)
	}
	value := proto.GetExtension(options, policyv1.E_Access)
	access, ok := value.(*policyv1.AccessPolicy)
	if !ok {
		t.Fatalf("%s.%s is missing access policy metadata", service.FullName(), name)
	}
	if access.GetRole() != want {
		t.Errorf("%s.%s role = %s, want %s", service.FullName(), name, access.GetRole(), want)
	}
}

func requireMessageField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
	kind protoreflect.Kind,
	typeName protoreflect.FullName,
) protoreflect.FieldDescriptor {
	t.Helper()
	field := message.Fields().ByName(name)
	if field == nil {
		t.Fatalf("%s.%s is missing", message.FullName(), name)
	}
	if field.Number() != number {
		t.Errorf("%s.%s number = %d, want %d", message.FullName(), name, field.Number(), number)
	}
	if field.Kind() != kind {
		t.Errorf("%s.%s kind = %s, want %s", message.FullName(), name, field.Kind(), kind)
	}
	if typeName != "" {
		var got protoreflect.FullName
		switch kind {
		case protoreflect.EnumKind:
			got = field.Enum().FullName()
		case protoreflect.MessageKind:
			got = field.Message().FullName()
		}
		if got != typeName {
			t.Errorf("%s.%s type = %s, want %s", message.FullName(), name, got, typeName)
		}
	}
	return field
}

func requireMessageFieldCount(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	want int,
) {
	t.Helper()
	if got := message.Fields().Len(); got != want {
		t.Errorf("%s must expose exactly %d fields, got %d", message.FullName(), want, got)
	}
}

func requireNoIdentityDisplayMetadata(
	t *testing.T,
	message protoreflect.MessageDescriptor,
) {
	t.Helper()
	for _, name := range []protoreflect.Name{"identity_name", "identity_email"} {
		if message.Fields().ByName(name) != nil {
			t.Errorf("%s must not expose %s", message.FullName(), name)
		}
	}
}
