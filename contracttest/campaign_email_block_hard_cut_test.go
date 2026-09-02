package contracttest_test

import (
	"testing"

	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestEmailProfileRejectsFileBlocks(t *testing.T) {
	document := &contentv1.RichTextDocument{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_EMAIL,
		SourceLocale:            "en",
		Base: &contentv1.RichTextBlockGraph{Nodes: []*contentv1.RichTextBlockNode{{
			Block: &contentv1.RichTextBlock{
				Id: "018f0d9c-1f54-7d51-8abc-a1b2c3d4e5f6",
				Value: &contentv1.RichTextBlock_File{File: &contentv1.FileBlock{Props: &contentv1.FileProps{
					Attachment: &contentv1.FileAttachment{State: &contentv1.FileAttachment_MissingAttachment{
						MissingAttachment: &contentv1.MissingAttachment{FormerFileId: "018f0d9c-1f54-7d51-8abc-a1b2c3d4e5f7"},
					}},
				}}},
			},
			Placement: &contentv1.ContentBlockPlacement{Index: 0},
		}}},
	}
	if err := contentv1.ValidateRichTextDocument(document, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
		t.Fatal("Email profile must reject File Blocks")
	}
}

func TestCampaignAndEmailTemplateCollaborationUseTypedBlockAggregates(t *testing.T) {
	tests := []struct {
		name            string
		service         protoreflect.ServiceDescriptor
		load            protoreflect.MessageDescriptor
		apply           protoreflect.MessageDescriptor
		metadata        protoreflect.MessageDescriptor
		resourceIDField protoreflect.Name
	}{
		{
			name:            "campaign",
			service:         intrav1.File_api_intra_v1_campaign_proto.Services().ByName("InternalCampaignService"),
			load:            (&intrav1.LoadCampaignDocumentRequest{}).ProtoReflect().Descriptor(),
			apply:           (&intrav1.ApplyCampaignBlockBatchRequest{}).ProtoReflect().Descriptor(),
			metadata:        (&intrav1.UpdateCampaignLocaleMetadataRequest{}).ProtoReflect().Descriptor(),
			resourceIDField: "campaign_id",
		},
		{
			name:            "email template",
			service:         intrav1.File_api_intra_v1_email_template_proto.Services().ByName("InternalEmailTemplateService"),
			load:            (&intrav1.LoadEmailTemplateDocumentRequest{}).ProtoReflect().Descriptor(),
			apply:           (&intrav1.ApplyEmailTemplateBlockBatchRequest{}).ProtoReflect().Descriptor(),
			metadata:        (&intrav1.UpdateEmailTemplateLocaleMetadataRequest{}).ProtoReflect().Descriptor(),
			resourceIDField: "email_template_id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.service == nil {
				t.Fatal("internal service descriptor is missing")
			}
			metadataMethod := protoreflect.Name("UpdateCampaignLocaleMetadata")
			if test.name == "email template" {
				metadataMethod = "UpdateEmailTemplateLocaleMetadata"
			}
			for _, method := range []protoreflect.Name{"ApplyBlockBatch", metadataMethod, "LoadDocument"} {
				if test.service.Methods().ByName(method) == nil {
					t.Errorf("service must expose %s", method)
				}
			}
			if test.service.Methods().ByName("CreateAuditCheckpoint") != nil {
				t.Error("Version-less documents must not expose an Audit checkpoint")
			}
			for _, removed := range []protoreflect.Name{
				"SaveDocument",
				"SaveTranslationDocument",
				"LoadTranslationDocument",
				"CreateVersionCheckpoint",
				"CheckpointSourceRevision",
			} {
				if test.service.Methods().ByName(removed) != nil {
					t.Errorf("service must not expose legacy method %s", removed)
				}
			}

			requireField(t, test.load, test.resourceIDField, 1, protoreflect.StringKind)
			principal := requireField(t, test.load, "principal", 2, protoreflect.MessageKind)
			if got, want := principal.Message().FullName(), protoreflect.FullName("api.intra.v1.CollaborationPrincipal"); got != want {
				t.Fatalf("%s.principal type = %s, want %s", test.load.FullName(), got, want)
			}

			requireField(t, test.apply, test.resourceIDField, 1, protoreflect.StringKind)
			batch := requireField(t, test.apply, "batch", 2, protoreflect.MessageKind)
			if got, want := batch.Message().FullName(), protoreflect.FullName("api.content.v1.RichTextBlockMutationBatch"); got != want {
				t.Fatalf("%s.batch type = %s, want %s", test.apply.FullName(), got, want)
			}
			requireCampaignEmailContributorIDsField(t, batch.Message(), 6)

			requireField(t, test.metadata, test.resourceIDField, 1, protoreflect.StringKind)
			requireField(t, test.metadata, "expected_revision", 2, protoreflect.StringKind)
			subject := requireField(t, test.metadata, "subject", 3, protoreflect.StringKind)
			if !subject.HasOptionalKeyword() {
				t.Error("metadata subject must preserve explicit presence")
			}
			requireCampaignEmailContributorIDsField(t, test.metadata, 4)
			requireField(t, test.metadata, "locale", 5, protoreflect.StringKind)
			expectedTargetRevision := requireField(t, test.metadata, "expected_target_revision", 6, protoreflect.StringKind)
			if !expectedTargetRevision.HasOptionalKeyword() {
				t.Error("target metadata CAS token must be optional for source rooms")
			}
			for _, removed := range []protoreflect.Name{
				"patch_mask",
				"expected_source_revision",
				"source_revision_checkpoint",
				"yjs_state",
				"content_json",
				"content_html",
			} {
				requireNoField(t, test.metadata, removed)
			}
		})
	}
}

func TestCampaignAndEmailTemplateManageContractsExposeTypedEditorDocuments(t *testing.T) {
	for _, descriptor := range []protoreflect.MessageDescriptor{
		(&managev1.Campaign{}).ProtoReflect().Descriptor(),
		(&managev1.EmailTemplate{}).ProtoReflect().Descriptor(),
	} {
		documentNumber := protoreflect.FieldNumber(14)
		if descriptor.Name() == "EmailTemplate" {
			documentNumber = 6
		}
		document := requireField(t, descriptor, "document", documentNumber, protoreflect.MessageKind)
		if got, want := document.Message().FullName(), protoreflect.FullName("api.content.v1.RichTextDocument"); got != want {
			t.Fatalf("%s.document type = %s, want %s", descriptor.FullName(), got, want)
		}
		requireField(t, descriptor, "document_revision", 15, protoreflect.StringKind)
		requireNoField(t, descriptor, "source_epoch")
		requireNoField(t, descriptor, "document_canonical_hash")
		for _, removed := range []protoreflect.Name{"content_html", "content_json", "content_text", "yjs_state", "edit_hash"} {
			requireNoField(t, descriptor, removed)
		}
	}

	campaignPreview := (&managev1.PreviewCampaignRequest{}).ProtoReflect().Descriptor()
	requireLocalizedRichTextDocumentField(t, campaignPreview, "document", 5)
	requireNoField(t, campaignPreview, "html_content")

	templatePreview := (&managev1.PreviewEmailTemplateRequest{}).ProtoReflect().Descriptor()
	requireLocalizedRichTextDocumentField(t, templatePreview, "document", 4)
	requireNoField(t, templatePreview, "content_html")
}

func requireCampaignEmailContributorIDsField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	number protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireField(t, descriptor, "contributor_member_ids", number, protoreflect.StringKind)
	if !field.IsList() {
		t.Fatalf("%s.contributor_member_ids must be repeated", descriptor.FullName())
	}
}

func requireLocalizedRichTextDocumentField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	name protoreflect.Name,
	number protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireField(t, descriptor, name, number, protoreflect.MessageKind)
	if got, want := field.Message().FullName(), protoreflect.FullName("api.content.v1.LocalizedRichTextDocument"); got != want {
		t.Fatalf("%s.%s type = %s, want %s", descriptor.FullName(), name, got, want)
	}
}
