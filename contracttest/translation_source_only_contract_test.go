package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationCapableRootsExposeInitialAndCurrentSourceLocale(t *testing.T) {
	creates := []struct {
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{(&managev1.CreatePostRequest{}).ProtoReflect().Descriptor(), 9},
		{(&managev1.CreatePageRequest{}).ProtoReflect().Descriptor(), 5},
		{(&managev1.CreateWorkRequest{}).ProtoReflect().Descriptor(), 13},
		{(&managev1.CreateProgramEventRequest{}).ProtoReflect().Descriptor(), 3},
		{(&managev1.CreateArtistRequest{}).ProtoReflect().Descriptor(), 9},
		{(&managev1.CreateLabelRequest{}).ProtoReflect().Descriptor(), 8},
		{(&managev1.CreateReleaseRequest{}).ProtoReflect().Descriptor(), 11},
		{(&managev1.CreateSeriesRequest{}).ProtoReflect().Descriptor(), 4},
		{(&managev1.CreateCampaignRequest{}).ProtoReflect().Descriptor(), 5},
		{(&managev1.CreateEmailTemplateRequest{}).ProtoReflect().Descriptor(), 5},
		{(&managev1.CreateEmailLayoutRequest{}).ProtoReflect().Descriptor(), 4},
		{(&managev1.CreateFormRequest{}).ProtoReflect().Descriptor(), 12},
		{(&managev1.CreateMenuRequest{}).ProtoReflect().Descriptor(), 3},
	}
	for _, contract := range creates {
		field := requireField(t, contract.descriptor, "source_locale", contract.number, protoreflect.StringKind)
		if field.HasOptionalKeyword() || field.IsList() {
			t.Errorf("%s.source_locale must be one required application scalar", contract.descriptor.FullName())
		}
	}

	reads := []struct {
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{(&managev1.Post{}).ProtoReflect().Descriptor(), 25},
		{(&managev1.Page{}).ProtoReflect().Descriptor(), 16},
		{(&managev1.PageSummary{}).ProtoReflect().Descriptor(), 9},
		{(&managev1.Work{}).ProtoReflect().Descriptor(), 24},
		{(&managev1.MyCreditedWork{}).ProtoReflect().Descriptor(), 12},
		{(&managev1.ProgramEvent{}).ProtoReflect().Descriptor(), 3},
		{(&managev1.ProgramEventSummary{}).ProtoReflect().Descriptor(), 3},
		{(&managev1.Artist{}).ProtoReflect().Descriptor(), 18},
		{(&managev1.Label{}).ProtoReflect().Descriptor(), 17},
		{(&managev1.Release{}).ProtoReflect().Descriptor(), 19},
		{(&managev1.Series{}).ProtoReflect().Descriptor(), 10},
		{(&managev1.SeriesSimple{}).ProtoReflect().Descriptor(), 4},
		{(&managev1.SeriesPost{}).ProtoReflect().Descriptor(), 7},
		{(&managev1.Campaign{}).ProtoReflect().Descriptor(), 16},
		{(&managev1.EmailTemplate{}).ProtoReflect().Descriptor(), 16},
		{(&managev1.EmailLayout{}).ProtoReflect().Descriptor(), 10},
		{(&managev1.Form{}).ProtoReflect().Descriptor(), 18},
		{(&managev1.FormSummary{}).ProtoReflect().Descriptor(), 8},
		{(&managev1.Menu{}).ProtoReflect().Descriptor(), 6},
		{(&managev1.Privacy{}).ProtoReflect().Descriptor(), 11},
		{(&managev1.Terms{}).ProtoReflect().Descriptor(), 11},
	}
	for _, contract := range reads {
		requireField(t, contract.descriptor, "source_locale", contract.number, protoreflect.StringKind)
	}

	for _, inherited := range []protoreflect.MessageDescriptor{
		(&managev1.CreatePrivacyVersionRequest{}).ProtoReflect().Descriptor(),
		(&managev1.CreateTermsVersionRequest{}).ProtoReflect().Descriptor(),
	} {
		requireNoField(t, inherited, "source_locale")
	}
}

func TestEmailLayoutMetadataUpdateCannotBypassDocumentCAS(t *testing.T) {
	request := (&managev1.UpdateEmailLayoutRequest{}).ProtoReflect().Descriptor()
	requireNoField(t, request, "html_content")
}

func TestBlockCollaborationExposesExactRequestedLocaleProjection(t *testing.T) {
	tests := []struct {
		load                   protoreflect.MessageDescriptor
		apply                  protoreflect.MessageDescriptor
		request                protoreflect.MessageDescriptor
		sourceMetadataNumber   protoreflect.FieldNumber
		localeNumber           protoreflect.FieldNumber
		localeExistsNumber     protoreflect.FieldNumber
		localeMetadataNumber   protoreflect.FieldNumber
		targetRevisionNumber   protoreflect.FieldNumber
		requestLocaleNumber    protoreflect.FieldNumber
		expectedTargetRevision protoreflect.FieldNumber
		documentType           protoreflect.FullName
	}{
		{(&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdatePostLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 6, 7, 8, 9, 10, 6, 7, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadPageBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyPageBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdatePageLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5, 6, 7, 8, 9, 7, 8, "api.content.v1.LocalizedPageDocument"},
		{(&intrav1.LoadWorkBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyWorkBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateWorkLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4, 5, 6, 7, 8, 7, 8, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadProgramEventBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyProgramEventBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateProgramEventLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4, 5, 6, 7, 8, 6, 7, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadArtistBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyArtistBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateArtistLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4, 5, 6, 7, 8, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadLabelBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyLabelBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateLabelLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4, 5, 6, 7, 8, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadReleaseBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyReleaseBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateReleaseLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4, 5, 6, 7, 8, 6, 7, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadCampaignDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyCampaignBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateCampaignLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 3, 4, 5, 6, 7, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadEmailTemplateDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyEmailTemplateBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateEmailTemplateLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 3, 4, 5, 6, 7, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadPrivacyBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyPrivacyBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdatePrivacyLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 3, 4, 5, 6, 7, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
		{(&intrav1.LoadTermsBlockDocumentResponse{}).ProtoReflect().Descriptor(), (&intrav1.ApplyTermsBlockBatchRequest{}).ProtoReflect().Descriptor(), (&intrav1.UpdateTermsLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 3, 4, 5, 6, 7, 5, 6, "api.content.v1.LocalizedRichTextDocument"},
	}
	for _, contract := range tests {
		document := requireField(t, contract.load, "document", 1, protoreflect.MessageKind)
		if document.Message().FullName() != contract.documentType {
			t.Errorf("%s.document type = %s, want %s", contract.load.FullName(), document.Message().FullName(), contract.documentType)
		}
		sourceMetadata := requireField(t, contract.load, "source_metadata", contract.sourceMetadataNumber, protoreflect.MessageKind)
		if sourceMetadata.IsList() {
			t.Errorf("%s.source_metadata must be singular", contract.load.FullName())
		}
		requireField(t, sourceMetadata.Message(), "locale", 1, protoreflect.StringKind)
		requireField(t, contract.load, "locale", contract.localeNumber, protoreflect.StringKind)
		requireField(t, contract.load, "locale_exists", contract.localeExistsNumber, protoreflect.BoolKind)
		localeMetadata := requireField(t, contract.load, "locale_metadata", contract.localeMetadataNumber, protoreflect.MessageKind)
		if sourceMetadata.Message().FullName() != localeMetadata.Message().FullName() || !localeMetadata.HasOptionalKeyword() {
			t.Errorf("%s.locale_metadata must be one optional exact-locale projection matching source_metadata", contract.load.FullName())
		}
		targetRevision := requireField(t, contract.load, "target_revision", contract.targetRevisionNumber, protoreflect.StringKind)
		if !targetRevision.HasOptionalKeyword() {
			t.Errorf("%s.target_revision must be optional", contract.load.FullName())
		}
		locale := requireField(t, contract.request, "locale", contract.requestLocaleNumber, protoreflect.StringKind)
		if locale.HasOptionalKeyword() || locale.IsList() {
			t.Errorf("%s.locale must be one required canonical room locale", contract.request.FullName())
		}
		expectedTargetRevision := requireField(t, contract.request, "expected_target_revision", contract.expectedTargetRevision, protoreflect.StringKind)
		if !expectedTargetRevision.HasOptionalKeyword() {
			t.Errorf("%s.expected_target_revision must be optional for source rooms", contract.request.FullName())
		}
		applyExpectedTargetRevision := requireField(t, contract.apply, "expected_target_revision", 4, protoreflect.StringKind)
		if !applyExpectedTargetRevision.HasOptionalKeyword() {
			t.Errorf("%s.expected_target_revision must be optional for source rooms", contract.apply.FullName())
		}
	}

	workLoad := (&intrav1.LoadWorkBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requireNoField(t, workLoad, "title")
	workSource := requireField(t, workLoad, "source_metadata", 4, protoreflect.MessageKind).Message()
	requireField(t, workSource, "summary", 2, protoreflect.StringKind)
	requireField(t, workSource, "title", 3, protoreflect.StringKind)
}

func TestPostCollaborationExposesExactRequestedLocaleOverlay(t *testing.T) {
	load := (&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor()
	requireField(t, load, "document", 1, protoreflect.MessageKind)
	sourceMetadata := requireField(t, load, "source_metadata", 6, protoreflect.MessageKind)
	localeMetadata := requireField(t, load, "locale_metadata", 9, protoreflect.MessageKind)
	if sourceMetadata.Message().FullName() != localeMetadata.Message().FullName() {
		t.Fatalf("Post source and requested locale metadata must use one typed shape")
	}
	requireField(t, sourceMetadata.Message(), "locale", 1, protoreflect.StringKind)
	requireField(t, load, "locale", 7, protoreflect.StringKind)
	requireField(t, load, "locale_exists", 8, protoreflect.BoolKind)
	targetRevision := requireField(t, load, "target_revision", 10, protoreflect.StringKind)
	if !targetRevision.HasOptionalKeyword() {
		t.Fatal("Post target_revision must be optional for source and missing-target rooms")
	}

	update := (&intrav1.UpdatePostLocaleMetadataRequest{}).ProtoReflect().Descriptor()
	requireField(t, update, "locale", 6, protoreflect.StringKind)
	expectedTargetRevision := requireField(t, update, "expected_target_revision", 7, protoreflect.StringKind)
	if !expectedTargetRevision.HasOptionalKeyword() {
		t.Fatal("Post expected_target_revision must be optional for source and missing-target rooms")
	}
}

func TestEmailLayoutCollaborationUsesOneExactLocaleDocumentContract(t *testing.T) {
	service := intrav1.File_api_intra_v1_email_layout_proto.Services().ByName("InternalEmailLayoutService")
	for _, method := range []protoreflect.Name{"SaveDocument", "LoadDocument"} {
		if service.Methods().ByName(method) == nil {
			t.Fatalf("Email Layout exact-locale RPC %s is missing", method)
		}
	}
	saveRequest := (&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor()
	requireField(t, saveRequest, "locale", 5, protoreflect.StringKind)
	requireField(t, saveRequest, "expected_document_revision", 6, protoreflect.StringKind)
	requireOptionalField(t, saveRequest, "expected_target_revision", 7, protoreflect.StringKind)
	saveResponse := (&intrav1.SaveEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor()
	requireField(t, saveResponse, "locale", 2, protoreflect.StringKind)
	requireField(t, saveResponse, "document_revision", 3, protoreflect.StringKind)
	requireOptionalField(t, saveResponse, "target_revision", 4, protoreflect.StringKind)
	requireField(t, (&intrav1.LoadEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor(), "locale", 2, protoreflect.StringKind)
	loadResponse := (&intrav1.LoadEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor()
	requireField(t, loadResponse, "source_locale", 3, protoreflect.StringKind)
	requireField(t, loadResponse, "locale", 4, protoreflect.StringKind)
	requireField(t, loadResponse, "document_revision", 5, protoreflect.StringKind)
	requireOptionalField(t, loadResponse, "target_revision", 6, protoreflect.StringKind)
}

func TestTranslationSettingsProtectedTermsAreRepeated(t *testing.T) {
	settings := (&managev1.TranslationSettings{}).ProtoReflect().Descriptor()
	protectedTerms := requireField(t, settings, "protected_terms", 3, protoreflect.StringKind)
	if !protectedTerms.IsList() {
		t.Fatal("TranslationSettings.protected_terms must be a repeated exact-spelling list")
	}
}

func TestTranslationHealthContainsObservedFactsOnly(t *testing.T) {
	tests := []struct {
		descriptor protoreflect.MessageDescriptor
		wantFields []struct {
			name   protoreflect.Name
			number protoreflect.FieldNumber
		}
	}{
		{
			descriptor: (&managev1.TranslationOverviewStats{}).ProtoReflect().Descriptor(),
			wantFields: []struct {
				name   protoreflect.Name
				number protoreflect.FieldNumber
			}{
				{"total_locales", 1},
				{"enabled_locales", 2},
				{"public_locales", 3},
				{"source_entities", 4},
				{"active_jobs", 5},
				{"existing_entries", 6},
			},
		},
		{
			descriptor: (&managev1.TranslationEntityHealth{}).ProtoReflect().Descriptor(),
			wantFields: []struct {
				name   protoreflect.Name
				number protoreflect.FieldNumber
			}{
				{"entity_type", 1},
				{"source_entities", 2},
				{"existing_entries", 3},
				{"active_jobs", 4},
				{"last_source_update_at", 5},
			},
		},
	}
	for _, test := range tests {
		if test.descriptor.Fields().Len() != len(test.wantFields) {
			t.Fatalf("%s fields = %d, want %d", test.descriptor.FullName(), test.descriptor.Fields().Len(), len(test.wantFields))
		}
		for index, want := range test.wantFields {
			field := test.descriptor.Fields().Get(index)
			if field.Name() != want.name || field.Number() != want.number {
				t.Fatalf("%s field %d = %s/%d, want %s/%d", test.descriptor.FullName(), index, field.Name(), field.Number(), want.name, want.number)
			}
		}
	}

	localeHealth := (&managev1.TranslationLocaleHealth{}).ProtoReflect().Descriptor()
	requireField(t, localeHealth, "existing_entries", 2, protoreflect.Int32Kind)
	requireField(t, localeHealth, "last_target_update_at", 4, protoreflect.MessageKind)
}

func TestSourceLocaleChangeAndRegenerationStayExplicit(t *testing.T) {
	setSource := (&managev1.SetEntitySourceLocaleRequest{}).ProtoReflect().Descriptor()
	requireField(t, setSource, "target", 1, protoreflect.MessageKind)
	requireField(t, setSource, "source_locale", 2, protoreflect.StringKind)
	requireField(t, setSource, "expected_document_revision", 3, protoreflect.StringKind)
	requireNoField(t, setSource, "contributor_member_ids")
	if setSource.Fields().Len() != 3 {
		t.Fatalf("SetEntitySourceLocaleRequest fields = %d, want target/source_locale/expected_document_revision only", setSource.Fields().Len())
	}

	regenerate := (&managev1.RegenerateEntityTranslationsRequest{}).ProtoReflect().Descriptor()
	locales := requireField(t, regenerate, "locales", 2, protoreflect.StringKind)
	if !locales.IsList() {
		t.Fatal("RegenerateEntityTranslationsRequest.locales must be a repeated explicit selection")
	}

	generateEvent := (&managev1.TranslationGenerateEvent{}).ProtoReflect().Descriptor()
	if generateEvent.Fields().Len() != 1 {
		t.Fatalf("TranslationGenerateEvent fields = %d, want job_id only", generateEvent.Fields().Len())
	}
	requireField(t, generateEvent, "job_id", 1, protoreflect.StringKind)
}
