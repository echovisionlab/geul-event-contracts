package contracttest_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestCollaborationLocaleSourceHardCut(t *testing.T) {
	contracts := []struct {
		path     string
		messages map[string]string
	}{
		{"api/intra/v1/collaboration.proto", map[string]string{
			"CollaborationResource":          "string locale = 3;",
			"AuthorizeCollaborationResponse": "string locale = 4;",
		}},
		{"api/intra/v1/post.proto", versionBlockLocaleSourceFields("PostBlock", "PostBlock", "PostVersion", 7, 4)},
		{"api/intra/v1/page.proto", versionBlockLocaleSourceFields("PageBlock", "PageBlock", "PageVersion", 6, 4)},
		{"api/intra/v1/work.proto", versionBlockLocaleSourceFields("WorkBlock", "WorkBlock", "WorkVersion", 5, 4)},
		{"api/intra/v1/program_event.proto", blockLocaleSourceFields("ProgramEventBlock", "ProgramEventBlock", 5)},
		{"api/intra/v1/artist.proto", blockLocaleSourceFields("ArtistBlock", "ArtistBlock", 5)},
		{"api/intra/v1/label.proto", blockLocaleSourceFields("LabelBlock", "LabelBlock", 5)},
		{"api/intra/v1/release.proto", blockLocaleSourceFields("ReleaseBlock", "ReleaseBlock", 5)},
		{"api/intra/v1/campaign.proto", blockLocaleSourceFields("Campaign", "CampaignBlock", 4)},
		{"api/intra/v1/email_template.proto", blockLocaleSourceFields("EmailTemplate", "EmailTemplateBlock", 4)},
		{"api/intra/v1/privacy.proto", blockLocaleSourceFields("PrivacyBlock", "PrivacyBlock", 4)},
		{"api/intra/v1/terms.proto", blockLocaleSourceFields("TermsBlock", "TermsBlock", 4)},
	}

	for _, contract := range contracts {
		source := readSourceProto(t, contract.path)
		for message, field := range contract.messages {
			declaration := sourceDeclaration(t, source, "message", message)
			requireSourceContains(t, declaration, field)
			requireSourceExcludes(t, declaration, "optional string locale", "repeated string locale")
		}
	}
}

func TestCanonicalCollaborationDocumentsUseDocumentAndTargetRevisionCAS(t *testing.T) {
	form := readSourceProto(t, "api/intra/v1/form.proto")
	formSaveRequest := sourceDeclaration(t, form, "message", "SaveFormDocumentRequest")
	requireSourceContains(t, formSaveRequest,
		"optional FormMeta meta = 2;",
		"string locale = 4;",
		"repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 5;",
		"string expected_document_revision = 6;",
		"optional string expected_target_revision = 7;",
	)
	requireSourceContains(t, sourceDeclaration(t, form, "message", "LoadFormDocumentRequest"), "string locale = 2;")
	requireSourceContains(t, sourceDeclaration(t, form, "message", "LoadFormDocumentResponse"),
		"FormMeta source_metadata = 1;",
		"optional FormMeta locale_metadata = 2;",
		"string source_locale = 3;",
		"string locale = 4;",
		"bool locale_exists = 5;",
		"repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 6;",
		"string document_revision = 7;",
		"optional string target_revision = 8;",
	)
	requireSourceContains(t, sourceDeclaration(t, form, "message", "SaveFormDocumentResponse"),
		"string locale = 2;",
		"string document_revision = 3;",
		"optional string target_revision = 4;",
	)
	requireSourceExcludes(t, formSaveRequest, "yjs_state", "og_generation_run_id")

	emailLayout := readSourceProto(t, "api/intra/v1/email_layout.proto")
	requireSourceContains(t, sourceDeclaration(t, emailLayout, "message", "SaveEmailLayoutDocumentRequest"),
		"string locale = 5;",
		"string expected_document_revision = 6;",
		"optional string expected_target_revision = 7;",
	)
	requireSourceContains(t, sourceDeclaration(t, emailLayout, "message", "LoadEmailLayoutDocumentRequest"), "string locale = 2;")
	requireSourceContains(t, sourceDeclaration(t, emailLayout, "message", "LoadEmailLayoutDocumentResponse"),
		"string source_locale = 3;",
		"string locale = 4;",
		"string document_revision = 5;",
		"optional string target_revision = 6;",
	)
	requireSourceContains(t, sourceDeclaration(t, emailLayout, "message", "SaveEmailLayoutDocumentResponse"),
		"string locale = 2;",
		"string document_revision = 3;",
		"optional string target_revision = 4;",
	)
	requireSourceExcludes(t, emailLayout, "yjs_state", "edit_hash", "expected_current_edit_hash", "expected_source_locale")

	mapTheme := readSourceProto(t, "api/intra/v1/map.proto")
	for _, message := range []string{"SaveMapThemeSnapshotRequest", "LoadMapThemeSnapshotRequest"} {
		declaration := sourceDeclaration(t, mapTheme, "message", message)
		requireSourceContains(t, declaration, "locale", "\"und\"")
	}
	requireSourceContains(t, sourceDeclaration(t, mapTheme, "message", "SaveMapThemeSnapshotResponse"), "string locale = 3;")
	requireSourceContains(t, sourceDeclaration(t, mapTheme, "message", "LoadMapThemeSnapshotResponse"), "string locale = 3;")
}

func TestCollaborationMetadataLocaleSourceHardCut(t *testing.T) {
	contracts := []struct {
		path          string
		request       string
		requestField  string
		response      string
		responseField string
	}{
		{"api/intra/v1/post.proto", "UpdatePostLocaleMetadataRequest", "string locale = 6;", "UpdatePostLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/post.proto", "UpdatePostDocumentMetadataRequest", "string locale = 6;", "UpdatePostDocumentMetadataResponse", "string locale = 4;"},
		{"api/intra/v1/page.proto", "UpdatePageLocaleMetadataRequest", "string locale = 7;", "UpdatePageLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/page.proto", "UpdatePageDocumentMetadataRequest", "string locale = 5;", "UpdatePageDocumentMetadataResponse", "string locale = 4;"},
		{"api/intra/v1/work.proto", "UpdateWorkLocaleMetadataRequest", "string locale = 7;", "UpdateWorkLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/artist.proto", "UpdateArtistLocaleMetadataRequest", "string locale = 5;", "UpdateArtistLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/artist.proto", "UpdateArtistDocumentMetadataRequest", "string locale = 5;", "UpdateArtistDocumentMetadataResponse", "string locale = 4;"},
		{"api/intra/v1/label.proto", "UpdateLabelLocaleMetadataRequest", "string locale = 5;", "UpdateLabelLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/label.proto", "UpdateLabelDocumentMetadataRequest", "string locale = 5;", "UpdateLabelDocumentMetadataResponse", "string locale = 4;"},
		{"api/intra/v1/release.proto", "UpdateReleaseLocaleMetadataRequest", "string locale = 6;", "UpdateReleaseLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/program_event.proto", "UpdateProgramEventLocaleMetadataRequest", "string locale = 6;", "UpdateProgramEventLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/campaign.proto", "UpdateCampaignLocaleMetadataRequest", "string locale = 5;", "UpdateCampaignLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/email_template.proto", "UpdateEmailTemplateLocaleMetadataRequest", "string locale = 5;", "UpdateEmailTemplateLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/privacy.proto", "UpdatePrivacyLocaleMetadataRequest", "string locale = 5;", "UpdatePrivacyLocaleMetadataResponse", "string locale = 5;"},
		{"api/intra/v1/terms.proto", "UpdateTermsLocaleMetadataRequest", "string locale = 5;", "UpdateTermsLocaleMetadataResponse", "string locale = 5;"},
	}

	for _, contract := range contracts {
		source := readSourceProto(t, contract.path)
		request := sourceDeclaration(t, source, "message", contract.request)
		requireSourceContains(t, request, contract.requestField)
		requireSourceExcludes(t, request, "optional string locale", "repeated string locale")
		response := sourceDeclaration(t, source, "message", contract.response)
		requireSourceContains(t, response, contract.responseField)
		requireSourceExcludes(t, response, "optional string locale", "repeated string locale")
	}
}

func TestRichTextTableSourceUsesDurableSemanticIDs(t *testing.T) {
	source := readRepositorySource(t, "scripts/generated/generate-block-catalog.mjs")
	requireSourceContains(t, sourceDeclaration(t, source, "message", "RichTextTableCellBase"),
		"string id = 1;",
		"bool header = 2;",
		"RichTextTableCellProps props = 3;",
	)
	requireSourceContains(t, sourceDeclaration(t, source, "message", "RichTextTableRowBase"),
		"string id = 1;",
		"repeated RichTextTableCellBase cells = 2;",
	)
	requireSourceContains(t, sourceDeclaration(t, source, "message", "RichTextTableCellLocale"),
		"string cell_id = 1;",
		"repeated RichTextInline content = 2;",
	)
	requireSourceContains(t, sourceDeclaration(t, source, "message", "RichTextTableRowLocale"),
		"string row_id = 1;",
		"repeated RichTextTableCellLocale cells = 2;",
	)
}

func blockLocaleSourceFields(load, apply string, loadResponseLocale int) map[string]string {
	return map[string]string{
		"Load" + load + "DocumentRequest":  "string locale = 3;",
		"Load" + load + "DocumentResponse": sourceLocaleField(loadResponseLocale),
		"Apply" + apply + "BatchRequest":   "string locale = 3;",
		"Apply" + apply + "BatchResponse":  "string locale = 5;",
	}
}

func versionBlockLocaleSourceFields(load, apply, checkpoint string, loadResponseLocale, checkpointResponseLocale int) map[string]string {
	fields := blockLocaleSourceFields(load, apply, loadResponseLocale)
	fields["Create"+checkpoint+"CheckpointRequest"] = "string locale = 4;"
	fields["Create"+checkpoint+"CheckpointResponse"] = sourceLocaleField(checkpointResponseLocale)
	return fields
}

func sourceLocaleField(number int) string {
	return "string locale = " + strconv.Itoa(number) + ";"
}

func readRepositorySource(t *testing.T, relativePath string) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source contract test path")
	}
	contents, err := os.ReadFile(filepath.Join(filepath.Dir(sourceFile), "..", filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatalf("read repository source %s: %v", relativePath, err)
	}
	return string(contents)
}
