package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCampaignSharedDocumentDoesNotCarryConfiguration(t *testing.T) {
	response := (&intrav1.LoadCampaignDocumentResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, response, 8)
	requireMessageField(t, response, "document", 1, protoreflect.MessageKind, "api.content.v1.LocalizedRichTextDocument")
	requireMessageField(t, response, "document_revision", 2, protoreflect.StringKind, "")
	requireMessageField(t, response, "source_metadata", 3, protoreflect.MessageKind, "api.intra.v1.CampaignLocaleMetadata")
	requireMessageField(t, response, "locale", 4, protoreflect.StringKind, "")
	requireMessageField(t, response, "locale_exists", 5, protoreflect.BoolKind, "")
	requireMessageField(t, response, "locale_metadata", 6, protoreflect.MessageKind, "api.intra.v1.CampaignLocaleMetadata")
	requireMessageField(t, response, "target_revision", 7, protoreflect.StringKind, "")
	requireMessageField(t, response, "present_locale_values", 8, protoreflect.MessageKind, "api.manage.v1.AIDocumentFieldTarget")
	if response.Fields().ByName("source_epoch") != nil || response.Fields().ByName("snapshot_digest") != nil {
		t.Fatal("Campaign collaboration load must not duplicate source identity or snapshot digest")
	}

	apply := (&intrav1.ApplyCampaignBlockBatchRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, apply, 5)
	requireMessageField(t, apply, "campaign_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, apply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")
	requireMessageField(t, apply, "locale", 3, protoreflect.StringKind, "")
	requireMessageField(t, apply, "expected_target_revision", 4, protoreflect.StringKind, "")
	requireMessageField(t, apply, "affected_locale_values", 5, protoreflect.MessageKind, "api.manage.v1.AIDocumentFieldTarget")
	for _, field := range []protoreflect.Name{"recipient_scope", "target_mode", "segment_id", "layout_id"} {
		if apply.Fields().ByName(field) != nil || response.Fields().ByName(field) != nil {
			t.Fatalf("Campaign collaboration document must not carry %s", field)
		}
	}
}

func TestCampaignTargetModeIsClosedAndTyped(t *testing.T) {
	targetMode := managev1.CampaignTargetMode(0).Descriptor()
	if got := targetMode.Values().Len(); got != 3 {
		t.Fatalf("api.manage.v1.CampaignTargetMode value count = %d, want 3", got)
	}
	for index, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{name: "CAMPAIGN_TARGET_MODE_UNSPECIFIED", number: 0},
		{name: "CAMPAIGN_TARGET_MODE_ALL", number: 1},
		{name: "CAMPAIGN_TARGET_MODE_SEGMENT", number: 2},
	} {
		value := targetMode.Values().Get(index)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Errorf(
				"api.manage.v1.CampaignTargetMode value %d = %s/%d, want %s/%d",
				index,
				value.Name(),
				value.Number(),
				expected.name,
				expected.number,
			)
		}
	}

	campaign := (&managev1.Campaign{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, campaign, 16)
	field := requireMessageField(
		t,
		campaign,
		"target_mode",
		13,
		protoreflect.EnumKind,
		"api.manage.v1.CampaignTargetMode",
	)
	if field.HasPresence() {
		t.Error("api.manage.v1.Campaign.target_mode must be a required typed enum")
	}
}

func TestCampaignRecipientScopeIsPositiveAndTyped(t *testing.T) {
	recipientScope := managev1.CampaignRecipientScope(0).Descriptor()
	if got := recipientScope.Values().Len(); got != 2 {
		t.Fatalf("api.manage.v1.CampaignRecipientScope value count = %d, want 2", got)
	}
	for index, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{name: "CAMPAIGN_RECIPIENT_SCOPE_SUBSCRIBED_USERS", number: 0},
		{name: "CAMPAIGN_RECIPIENT_SCOPE_ALL_MATCHING_USERS", number: 1},
	} {
		value := recipientScope.Values().Get(index)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Errorf(
				"api.manage.v1.CampaignRecipientScope value %d = %s/%d, want %s/%d",
				index,
				value.Name(),
				value.Number(),
				expected.name,
				expected.number,
			)
		}
	}

	for _, expected := range []struct {
		message protoreflect.MessageDescriptor
		name    protoreflect.Name
		number  protoreflect.FieldNumber
	}{
		{message: (&managev1.Campaign{}).ProtoReflect().Descriptor(), name: "recipient_scope", number: 11},
		{message: (&managev1.ScheduleCampaignRequest{}).ProtoReflect().Descriptor(), name: "recipient_scope", number: 3},
		{message: (&managev1.SendCampaignNowRequest{}).ProtoReflect().Descriptor(), name: "recipient_scope", number: 2},
	} {
		field := requireMessageField(
			t,
			expected.message,
			expected.name,
			expected.number,
			protoreflect.EnumKind,
			"api.manage.v1.CampaignRecipientScope",
		)
		if field.HasPresence() {
			t.Errorf("%s.%s must be a required typed enum", expected.message.FullName(), field.Name())
		}
	}

	recipient := (&managev1.CampaignRecipient{}).ProtoReflect().Descriptor()
	requireMessageField(t, recipient, "member_id", 6, protoreflect.StringKind, "")
}

func TestCampaignSubjectIsSourceOwned(t *testing.T) {
	saveTranslation := (&intrav1.UpdateCampaignLocaleMetadataRequest{}).ProtoReflect().Descriptor()
	requireMessageField(
		t,
		saveTranslation,
		"subject",
		3,
		protoreflect.StringKind,
		"",
	)

	loadTranslation := (&intrav1.CampaignLocaleMetadata{}).ProtoReflect().Descriptor()
	requireMessageField(
		t,
		loadTranslation,
		"subject",
		2,
		protoreflect.StringKind,
		"",
	)
}

func TestAdminSegmentExposesReferenceCountsAndArchiveState(t *testing.T) {
	segment := (&managev1.Segment{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, segment, 12)
	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{name: "campaign_count", number: 9},
		{name: "delivery_run_count", number: 10},
		{name: "download_policy_reference_count", number: 11},
	} {
		field := requireMessageField(
			t,
			segment,
			expected.name,
			expected.number,
			protoreflect.Int32Kind,
			"",
		)
		if field.HasPresence() {
			t.Errorf("api.manage.v1.Segment.%s must be a non-optional count", expected.name)
		}
	}
	archivedAt := requireMessageField(
		t,
		segment,
		"archived_at",
		12,
		protoreflect.MessageKind,
		"google.protobuf.Timestamp",
	)
	if !archivedAt.HasPresence() {
		t.Error("api.manage.v1.Segment.archived_at must preserve active versus archived presence")
	}

	if segment.ReservedNames().Len() != 0 || segment.ReservedRanges().Len() != 0 {
		t.Error("api.manage.v1.Segment must not retain a compatibility reservation")
	}

	summary := (&managev1.AudienceSegmentSummary{}).ProtoReflect().Descriptor()
	for _, name := range []protoreflect.Name{
		"campaign_count",
		"delivery_run_count",
		"download_policy_reference_count",
	} {
		if summary.Fields().ByName(name) != nil {
			t.Errorf("api.manage.v1.AudienceSegmentSummary must not expose admin %s usage", name)
		}
	}
}

func TestAudienceSegmentArchiveLifecycleContract(t *testing.T) {
	list := (&managev1.ListSegmentsAdminRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, list, 4)
	requireMessageField(t, list, "pagination", 1, protoreflect.MessageKind, "api.common.v1.PaginationRequest")
	requireMessageField(t, list, "filters", 2, protoreflect.MessageKind, "api.common.v1.FilterSpec")
	requireMessageField(t, list, "sorts", 3, protoreflect.MessageKind, "api.common.v1.SortSpec")
	requireMessageField(t, list, "include_archived", 4, protoreflect.BoolKind, "")

	archive := (&managev1.ArchiveSegmentRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, archive, 1)
	requireMessageField(t, archive, "id", 1, protoreflect.StringKind, "")

	restore := (&managev1.RestoreSegmentRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, restore, 1)
	requireMessageField(t, restore, "id", 1, protoreflect.StringKind, "")

	file := managev1.File_api_manage_v1_audience_proto
	service := file.Services().ByName("AudienceService")
	if service == nil {
		t.Fatal("api.manage.v1.AudienceService descriptor is missing")
	}
	for _, expected := range []struct {
		method  protoreflect.Name
		request protoreflect.FullName
	}{
		{method: "ArchiveSegment", request: "api.manage.v1.ArchiveSegmentRequest"},
		{method: "RestoreSegment", request: "api.manage.v1.RestoreSegmentRequest"},
	} {
		method := service.Methods().ByName(expected.method)
		if method == nil {
			t.Fatalf("%s.%s descriptor is missing", service.FullName(), expected.method)
		}
		if got := method.Input().FullName(); got != expected.request {
			t.Errorf("%s.%s input = %s, want %s", service.FullName(), expected.method, got, expected.request)
		}
		if got := method.Output().FullName(); got != "api.manage.v1.Segment" {
			t.Errorf("%s.%s output = %s, want api.manage.v1.Segment", service.FullName(), expected.method, got)
		}
		requireMethodTier(t, service, expected.method, policyv1.AuthorizationRole_ADMIN)
	}
}

func TestAudienceSegmentConfigUsesCanonicalTypedFilters(t *testing.T) {
	config := (&managev1.SegmentConfig{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, config, 5)

	segmentType := managev1.SegmentType(0).Descriptor()
	for index, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.EnumNumber
	}{
		{name: "SEGMENT_TYPE_UNSPECIFIED", number: 0},
		{name: "SEGMENT_TYPE_ALL_MEMBERS", number: 1},
		{name: "SEGMENT_TYPE_MEMBER_TAGS", number: 2},
		{name: "SEGMENT_TYPE_MEMBERS_BY_FILTER", number: 3},
	} {
		value := segmentType.Values().Get(index)
		if value.Name() != expected.name || value.Number() != expected.number {
			t.Errorf("api.manage.v1.SegmentType value %d = %s/%d, want %s/%d", index, value.Name(), value.Number(), expected.name, expected.number)
		}
	}

	roles := requireMessageField(
		t,
		config,
		"account_roles",
		2,
		protoreflect.EnumKind,
		"api.policy.v1.AuthorizationRole",
	)
	if !roles.IsList() {
		t.Error("api.manage.v1.SegmentConfig.account_roles must be repeated")
	}

	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{name: "created_after", number: 3},
		{name: "created_before", number: 4},
	} {
		field := requireMessageField(
			t,
			config,
			expected.name,
			expected.number,
			protoreflect.MessageKind,
			"google.protobuf.Timestamp",
		)
		if !field.HasOptionalKeyword() {
			t.Errorf("api.manage.v1.SegmentConfig.%s must remain optional", expected.name)
		}
	}
	memberTags := requireMessageField(t, config, "member_tag_ids", 1, protoreflect.StringKind, "")
	if !memberTags.IsList() {
		t.Error("api.manage.v1.SegmentConfig.member_tag_ids must be repeated")
	}
	excluded := requireMessageField(t, config, "exclude_member_ids", 5, protoreflect.StringKind, "")
	if !excluded.IsList() {
		t.Error("api.manage.v1.SegmentConfig.exclude_member_ids must be repeated")
	}
}

func TestAdminEmailResourcesExposeDeletionBlockerCounts(t *testing.T) {
	template := (&managev1.EmailTemplate{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, template, 16)
	requireNonOptionalCountField(t, template, "delivery_run_count", 14)

	layout := (&managev1.EmailLayout{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, layout, 10)
	requireNonOptionalCountField(t, layout, "campaign_count", 7)
	requireNonOptionalCountField(t, layout, "template_count", 8)
	requireNonOptionalCountField(t, layout, "delivery_run_count", 9)
}

func TestAllSettingsSeparatesPersistedAndRuntimeProjections(t *testing.T) {
	settings := (&managev1.AllSettings{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, settings, 3)
	requireMessageField(t, settings, "runtime", 3, protoreflect.MessageKind, "api.manage.v1.RuntimeSettings")
}

func requireNonOptionalCountField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireMessageField(t, message, name, number, protoreflect.Int32Kind, "")
	if field.HasPresence() {
		t.Errorf("%s.%s must be a non-optional count", message.FullName(), name)
	}
}

func requireOptionalField(
	t *testing.T,
	message protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
	kind protoreflect.Kind,
) {
	t.Helper()
	field := requireMessageField(t, message, name, number, kind, "")
	if !field.HasOptionalKeyword() {
		t.Errorf("%s.%s must be optional", message.FullName(), name)
	}
}
