package contracttest_test

import (
	"testing"

	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldExpectation struct {
	name     protoreflect.Name
	number   protoreflect.FieldNumber
	kind     protoreflect.Kind
	optional bool
	repeated bool
}

func TestBlockHotMutationResponsesAreMinimalAcks(t *testing.T) {
	filesAndMessages := []struct {
		file     protoreflect.FileDescriptor
		messages []protoreflect.Name
	}{
		{intrav1.File_api_intra_v1_page_proto, []protoreflect.Name{"ApplyPageBlockBatchResponse", "UpdatePageLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_work_proto, []protoreflect.Name{"ApplyWorkBlockBatchResponse", "UpdateWorkLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_program_event_proto, []protoreflect.Name{"ApplyProgramEventBlockBatchResponse", "UpdateProgramEventLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_artist_proto, []protoreflect.Name{"ApplyArtistBlockBatchResponse", "UpdateArtistLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_label_proto, []protoreflect.Name{"ApplyLabelBlockBatchResponse", "UpdateLabelLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_release_proto, []protoreflect.Name{"ApplyReleaseBlockBatchResponse", "UpdateReleaseLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_campaign_proto, []protoreflect.Name{"ApplyCampaignBlockBatchResponse", "UpdateCampaignLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_email_template_proto, []protoreflect.Name{"ApplyEmailTemplateBlockBatchResponse", "UpdateEmailTemplateLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_privacy_proto, []protoreflect.Name{"ApplyPrivacyBlockBatchResponse", "UpdatePrivacyLocaleMetadataResponse"}},
		{intrav1.File_api_intra_v1_terms_proto, []protoreflect.Name{"ApplyTermsBlockBatchResponse", "UpdateTermsLocaleMetadataResponse"}},
	}

	want := []fieldExpectation{
		{name: "document_revision", number: 1, kind: protoreflect.StringKind},
		{name: "changed", number: 2, kind: protoreflect.BoolKind},
		{name: "source_changed", number: 3, kind: protoreflect.BoolKind},
		{name: "changed_locales", number: 4, kind: protoreflect.StringKind, repeated: true},
		{name: "locale", number: 5, kind: protoreflect.StringKind},
		{name: "target_revision", number: 6, kind: protoreflect.StringKind, optional: true},
	}
	for _, group := range filesAndMessages {
		for _, name := range group.messages {
			message := group.file.Messages().ByName(name)
			if message == nil {
				t.Fatalf("%s is missing", name)
			}
			assertExactFields(t, message, want)
		}
	}

	for _, name := range []protoreflect.Name{"ApplyPostBlockBatchResponse", "UpdatePostLocaleMetadataResponse"} {
		message := intrav1.File_api_intra_v1_post_proto.Messages().ByName(name)
		if message == nil {
			t.Fatalf("%s is missing", name)
		}
		assertExactFields(t, message, want)
	}
}

func TestRichTextTableLocaleUsesDurableSemanticReferences(t *testing.T) {
	cellBase := (&contentv1.RichTextTableCellBase{}).ProtoReflect().Descriptor()
	assertExactFields(t, cellBase, []fieldExpectation{
		{name: "id", number: 1, kind: protoreflect.StringKind},
		{name: "header", number: 2, kind: protoreflect.BoolKind},
		{name: "props", number: 3, kind: protoreflect.MessageKind},
	})

	rowBase := (&contentv1.RichTextTableRowBase{}).ProtoReflect().Descriptor()
	assertExactFields(t, rowBase, []fieldExpectation{
		{name: "id", number: 1, kind: protoreflect.StringKind},
		{name: "cells", number: 2, kind: protoreflect.MessageKind, repeated: true},
	})

	cellLocale := (&contentv1.RichTextTableCellLocale{}).ProtoReflect().Descriptor()
	assertExactFields(t, cellLocale, []fieldExpectation{
		{name: "cell_id", number: 1, kind: protoreflect.StringKind},
		{name: "content", number: 2, kind: protoreflect.MessageKind, repeated: true},
	})

	rowLocale := (&contentv1.RichTextTableRowLocale{}).ProtoReflect().Descriptor()
	assertExactFields(t, rowLocale, []fieldExpectation{
		{name: "row_id", number: 1, kind: protoreflect.StringKind},
		{name: "cells", number: 2, kind: protoreflect.MessageKind, repeated: true},
	})
}

func assertExactFields(t *testing.T, message protoreflect.MessageDescriptor, expected []fieldExpectation) {
	t.Helper()
	fields := message.Fields()
	if fields.Len() != len(expected) {
		t.Fatalf("%s fields = %d, want %d", message.FullName(), fields.Len(), len(expected))
	}
	for _, want := range expected {
		field := fields.ByName(want.name)
		if field == nil {
			t.Fatalf("%s.%s is missing", message.FullName(), want.name)
		}
		if field.Number() != want.number || field.Kind() != want.kind || field.HasOptionalKeyword() != want.optional || field.IsList() != want.repeated {
			t.Fatalf("%s.%s descriptor = number %d kind %s optional %t repeated %t", message.FullName(), want.name, field.Number(), field.Kind(), field.HasOptionalKeyword(), field.IsList())
		}
	}
}

func TestTranslationServiceExposesReadGenerationAndJobControlOnly(t *testing.T) {
	service := managev1.File_api_manage_v1_translation_proto.Services().ByName("TranslationService")
	for _, name := range []protoreflect.Name{
		"ListEntityTranslations",
		"GetEntityTranslation",
		"SetEntitySourceLocale",
		"RegenerateEntityTranslations",
		"ListTranslationJobs",
		"CancelTranslationJob",
		"ExportEntityTranslationXLIFF",
		"ImportEntityTranslationXLIFF",
	} {
		if service.Methods().ByName(name) == nil {
			t.Fatalf("%s RPC is missing", name)
		}
	}
	for _, name := range []protoreflect.Name{
		"EnsureEntityTranslation",
		"UpdateEntityTranslation",
		"DeleteEntityTranslation",
		"DeleteBlockDocumentLocale",
		"RetryTranslationJob",
	} {
		if service.Methods().ByName(name) != nil {
			t.Fatalf("source-only TranslationService must not expose %s", name)
		}
	}
}

func TestBlockBackedFacetAndLifecycleMutationsReturnDomainAcks(t *testing.T) {
	tests := []struct {
		file    protoreflect.FileDescriptor
		service protoreflect.Name
		method  protoreflect.Name
		output  protoreflect.Name
	}{
		{managev1.File_api_manage_v1_post_proto, "PostService", "UpdatePost", "UpdatePostResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "PublishPost", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "UnpublishPost", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "ArchivePost", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "SchedulePost", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "CancelPostSchedule", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_post_proto, "PostService", "RepublishPost", "PostLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_page_proto, "PageService", "UpdatePage", "UpdatePageResponse"},
		{managev1.File_api_manage_v1_page_proto, "PageService", "PublishPage", "PageLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_page_proto, "PageService", "UnpublishPage", "PageLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_work_proto, "WorkService", "UpdateWork", "UpdateWorkResponse"},
		{managev1.File_api_manage_v1_work_proto, "WorkService", "PublishWork", "WorkLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_work_proto, "WorkService", "UnpublishWork", "WorkLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "UpdateProgramEvent", "UpdateProgramEventResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "AddProgramEventMedia", "AddProgramEventMediaResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "DeleteProgramEventMedia", "DeleteProgramEventMediaResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "ReorderProgramEventMedia", "ReorderProgramEventMediaResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "ReorderProgramEventCredits", "ReorderProgramEventCreditsResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "PublishProgramEvent", "ProgramEventLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_program_event_proto, "ProgramEventService", "ArchiveProgramEvent", "ProgramEventLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_artist_proto, "ArtistService", "PublishArtist", "ArtistLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_artist_proto, "ArtistService", "UnpublishArtist", "ArtistLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_label_proto, "LabelService", "PublishLabel", "LabelLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_label_proto, "LabelService", "UnpublishLabel", "LabelLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_release_proto, "ReleaseService", "UpdateRelease", "UpdateReleaseResponse"},
		{managev1.File_api_manage_v1_release_proto, "ReleaseService", "PublishRelease", "ReleaseLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_release_proto, "ReleaseService", "UnpublishRelease", "ReleaseLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_campaign_proto, "CampaignService", "UpdateCampaignName", "UpdateCampaignNameResponse"},
		{managev1.File_api_manage_v1_campaign_proto, "CampaignService", "UpdateCampaignConfiguration", "UpdateCampaignConfigurationResponse"},
		{managev1.File_api_manage_v1_campaign_proto, "CampaignService", "ScheduleCampaign", "ScheduleCampaignResponse"},
		{managev1.File_api_manage_v1_campaign_proto, "CampaignService", "CancelCampaign", "CancelCampaignResponse"},
		{managev1.File_api_manage_v1_campaign_proto, "CampaignService", "SendCampaignNow", "SendCampaignNowResponse"},
		{managev1.File_api_manage_v1_email_template_proto, "EmailTemplateService", "UpdateEmailTemplate", "UpdateEmailTemplateResponse"},
		{managev1.File_api_manage_v1_privacy_proto, "PrivacyService", "SchedulePrivacy", "PrivacyLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_privacy_proto, "PrivacyService", "CancelPrivacySchedule", "PrivacyLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_privacy_proto, "PrivacyService", "ActivatePrivacyNow", "PrivacyLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_terms_proto, "TermsService", "ScheduleTerms", "TermsLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_terms_proto, "TermsService", "CancelTermsSchedule", "TermsLifecycleMutationResponse"},
		{managev1.File_api_manage_v1_terms_proto, "TermsService", "ActivateTermsNow", "TermsLifecycleMutationResponse"},
	}

	for _, test := range tests {
		service := test.file.Services().ByName(test.service)
		method := service.Methods().ByName(test.method)
		if method == nil || method.Output().Name() != test.output {
			t.Fatalf("%s.%s output = %v, want %s", test.service, test.method, method.Output().Name(), test.output)
		}
	}
}

func TestUpdateReleaseDateUsesExplicitSetClearOneof(t *testing.T) {
	request := (&managev1.UpdateReleaseRequest{}).ProtoReflect().Descriptor()
	setDate := request.Fields().ByName("set_release_date")
	clearDate := request.Fields().ByName("clear_release_date")
	if setDate == nil || clearDate == nil || setDate.Number() != 5 || clearDate.Number() != 6 {
		t.Fatal("UpdateReleaseRequest explicit set/clear date fields are missing")
	}
	if setDate.ContainingOneof() == nil || setDate.ContainingOneof() != clearDate.ContainingOneof() || setDate.ContainingOneof().Name() != "release_date_change" {
		t.Fatal("Release date set/clear fields do not share release_date_change oneof")
	}
}
