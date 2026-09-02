package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
)

func TestResidentCollaborationLoadsDeclareExactLocaleValuePresence(t *testing.T) {
	contracts := []struct {
		path    string
		message string
		field   string
	}{
		{"api/intra/v1/post.proto", "LoadPostBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 11;"},
		{"api/intra/v1/page.proto", "LoadPageBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 10;"},
		{"api/intra/v1/work.proto", "LoadWorkBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 9;"},
		{"api/intra/v1/program_event.proto", "LoadProgramEventBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 9;"},
		{"api/intra/v1/artist.proto", "LoadArtistBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 9;"},
		{"api/intra/v1/label.proto", "LoadLabelBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 9;"},
		{"api/intra/v1/release.proto", "LoadReleaseBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 9;"},
		{"api/intra/v1/campaign.proto", "LoadCampaignDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 8;"},
		{"api/intra/v1/email_template.proto", "LoadEmailTemplateDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 8;"},
		{"api/intra/v1/privacy.proto", "LoadPrivacyBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 8;"},
		{"api/intra/v1/terms.proto", "LoadTermsBlockDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 8;"},
		{"api/intra/v1/form.proto", "LoadFormDocumentResponse", "repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 6;"},
	}

	for _, contract := range contracts {
		source := readSourceProto(t, contract.path)
		requireSourceContains(t, source, `import "api/manage/v1/ai.proto";`)
		declaration := sourceDeclaration(t, source, "message", contract.message)
		requireSourceContains(t, declaration,
			contract.field,
			"Exact persisted locale-owned leaf-terminal presence",
			"never inferred from materialized/source-fallback values",
			"canonical-sorted, duplicate-free",
			"target rooms report actual presence",
			"missing target reports an empty list",
			"relation_item, shared/topology/File",
			"identity-terminal (row_id",
			"Page unit_id",
			"composite object/list targets are forbidden",
		)
	}
}

func TestFormCollaborationSaveCarriesWholePostStatePresence(t *testing.T) {
	source := readSourceProto(t, "api/intra/v1/form.proto")
	declaration := sourceDeclaration(t, source, "message", "SaveFormDocumentRequest")
	requireSourceContains(t, declaration,
		"repeated api.manage.v1.AIDocumentFieldTarget present_locale_values = 5;",
		"Whole post-state declaration of locale-owned leaf terminals",
		"validates source saves against",
		"canonical source presence",
		"validates/preserves target saves as sparse",
		"target presence",
		"canonical-sorted, duplicate-free",
		"relation_item,",
		"shared/topology/File",
		"identity-terminal (row_id, cell_id, Page unit_id)",
		"composite object/list targets are forbidden",
	)
}

func TestLocaleValuePresenceReusesAIDocumentFieldTargetAndExcludesEmailLayout(t *testing.T) {
	presence := (&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor().Fields().ByName("present_locale_values")
	if presence == nil || presence.Message().FullName() != "api.manage.v1.AIDocumentFieldTarget" {
		t.Fatalf("resident presence field does not reuse AIDocumentFieldTarget: %v", presence)
	}

	emailLayoutLoad := (&intrav1.LoadEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor()
	if field := emailLayoutLoad.Fields().ByName("present_locale_values"); field != nil {
		t.Fatalf("%s unexpectedly exposes present_locale_values", emailLayoutLoad.FullName())
	}
	emailLayoutSave := (&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor()
	if field := emailLayoutSave.Fields().ByName("present_locale_values"); field != nil {
		t.Fatalf("%s unexpectedly exposes present_locale_values", emailLayoutSave.FullName())
	}
}
