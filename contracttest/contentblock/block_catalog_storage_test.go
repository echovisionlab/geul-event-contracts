package contentblock_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const blockCatalogTestID = "00000000-0000-4000-8000-000000000001"
const blockCatalogColumnID = "00000000-0000-4000-8000-000000000002"
const blockCatalogSectionID = "00000000-0000-4000-8000-000000000004"
const blockCatalogFileID = "00000000-0000-4000-8000-000000000005"
const blockCatalogMissingFileID = "00000000-0000-4000-8000-000000000006"
const blockCatalogPreservedFileID = "00000000-0000-4000-8000-000000000007"
const blockCatalogUnitID = "00000000-0000-4000-8000-000000000008"

func TestStorageRoundTripAndSourceHash(t *testing.T) {
	t.Parallel()

	document := paragraphDocument("original")
	rows, err := contentv1.FlattenRichTextDocumentStorage(
		document,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten original: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	ko := proto.Clone(document.LocaleOverlays[0]).(*contentv1.RichTextLocaleOverlay)
	ko.Locale = "ko"
	ko.Blocks[0].GetParagraph().Content[0].GetText().Text = "한국어"
	document.LocaleOverlays = append(document.LocaleOverlays, ko)
	rows, err = contentv1.FlattenRichTextDocumentStorage(
		document,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten multilingual document: %v", err)
	}
	localized, err := contentv1.MaterializeLocalizedRichTextDocumentStorage(
		contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		"ko",
		rows,
	)
	if err != nil {
		t.Fatalf("materialize requested locale: %v", err)
	}
	if localized.LocaleOverlay.GetLocale() != "ko" || localized.LocaleOverlay.Blocks[0].GetParagraph().Content[0].GetText().Text != "한국어" {
		t.Fatalf("materialized wrong locale overlay: %#v", localized.LocaleOverlay)
	}

	rebuilt, err := contentv1.MaterializeRichTextDocumentStorage(
		contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		"en",
		rows,
	)
	if err != nil {
		t.Fatalf("materialize: %v", err)
	}
	rebuiltRows, err := contentv1.FlattenRichTextDocumentStorage(
		rebuilt,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten rebuilt: %v", err)
	}
	originalHash, err := contentv1.ContentStorageCanonicalHash("post", rows)
	if err != nil {
		t.Fatalf("hash original: %v", err)
	}
	rebuiltHash, err := contentv1.ContentStorageCanonicalHash("post", rebuiltRows)
	if err != nil {
		t.Fatalf("hash rebuilt: %v", err)
	}
	if originalHash != rebuiltHash {
		t.Fatalf("round-trip hash differs: %s != %s", originalHash, rebuiltHash)
	}

	originalSourceHash, err := contentv1.ContentStorageTranslationSourceHash("post", rows, "en")
	if err != nil {
		t.Fatalf("source hash original: %v", err)
	}
	changed := proto.Clone(document).(*contentv1.RichTextDocument)
	changed.LocaleOverlays[0].Blocks[0].GetParagraph().Content[0].GetText().Text = "changed"
	changedRows, err := contentv1.FlattenRichTextDocumentStorage(
		changed,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten changed: %v", err)
	}
	changedSourceHash, err := contentv1.ContentStorageTranslationSourceHash("post", changedRows, "en")
	if err != nil {
		t.Fatalf("source hash changed: %v", err)
	}
	if originalSourceHash == changedSourceHash {
		t.Fatal("source-localized text edit did not change translation source hash")
	}
}

func TestStorageKindUsesCatalogKebabCaseForHyphenatedProtoCase(t *testing.T) {
	t.Parallel()

	document := codeBlockDocument("const answer = 42")
	rows, err := contentv1.FlattenRichTextDocumentStorage(
		document,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten code Block: %v", err)
	}
	if len(rows) != 1 || rows[0].Kind != "code-block" {
		t.Fatalf("storage kind = %q, want catalog kind code-block", rows[0].Kind)
	}

	rebuilt, err := contentv1.MaterializeRichTextDocumentStorage(
		contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		"en",
		rows,
	)
	if err != nil {
		t.Fatalf("materialize code Block: %v", err)
	}
	if got := rebuilt.LocaleOverlays[0].Blocks[0].GetCodeBlock().GetContent(); got != "const answer = 42" {
		t.Fatalf("materialized code = %q", got)
	}
}

func TestProtobufJSONFieldNameUsesShow3dBuildings(t *testing.T) {
	t.Parallel()

	normalized, err := contentv1.NormalizeContentStorageBlock(
		"post",
		"map",
		[]byte(`{"map":{"props":{"show3dBuildings":true}}}`),
		[]byte(`{"map":{"props":{}}}`),
	)
	if err != nil {
		t.Fatalf("normalize protobuf JSON name: %v", err)
	}
	if string(normalized.SharedData) != `{"map":{"props":{"show3dBuildings":true}}}` {
		t.Fatalf("normalized shared JSON = %s", normalized.SharedData)
	}
	if _, err := contentv1.NormalizeContentStorageBlock(
		"post",
		"map",
		[]byte(`{"map":{"props":{"show3DBuildings":true}}}`),
		[]byte(`{"map":{"props":{}}}`),
	); err == nil {
		t.Fatal("legacy show3DBuildings JSON name was accepted")
	}
}

func TestStorageFileReferencesCarryExactAttachmentSelector(t *testing.T) {
	t.Parallel()

	expectedMissing := map[string]contentv1.MissingAttachmentMediaKind{
		"shader.stages.0.channels.1.file":                    contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE,
		"immersive_scene:" + blockCatalogUnitID + ":texture": contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE,
	}
	seenMissing := map[string]bool{}
	activeCount := 0
	for _, row := range []contentv1.ContentStorageRow{shaderStorageRow(t), immersiveStorageRow(t)} {
		source := row.Locales[0]
		normalized, err := contentv1.NormalizeContentStorageBlock("page", row.Kind, row.SharedData, source.LocalizedData)
		if err != nil {
			t.Fatalf("normalize %s storage Block: %v", row.Kind, err)
		}
		for _, reference := range normalized.FileReferences {
			if reference.Missing {
				want, ok := expectedMissing[reference.ReferencePath]
				if !ok || reference.MissingAttachmentMediaKind != want {
					t.Fatalf("missing selector %q kind = %s, want %s", reference.ReferencePath, reference.MissingAttachmentMediaKind, want)
				}
				seenMissing[reference.ReferencePath] = true
				continue
			}
			activeCount++
			if reference.MissingAttachmentMediaKind != contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_UNSPECIFIED {
				t.Fatalf("active selector %q has missing kind %s", reference.ReferencePath, reference.MissingAttachmentMediaKind)
			}
		}
	}
	if len(seenMissing) != len(expectedMissing) || activeCount != 3 {
		t.Fatalf("selectors = %d missing/%d active, want %d/3", len(seenMissing), activeCount, len(expectedMissing))
	}

	if _, err := contentv1.NormalizeContentStorageBlock(
		"post",
		"file",
		[]byte(`{"file":{"props":{"attachment":{"missingAttachment":{"formerFileId":"`+blockCatalogFileID+`"}}}}}`),
		[]byte(`{"file":{"props":{}}}`),
	); err == nil {
		t.Fatal("missing attachment without media kind was accepted")
	}
}

func TestNormalizeContentStorageSharedOwnsCanonicalValidationAndReferences(t *testing.T) {
	t.Parallel()

	upperFileID := strings.ToUpper(blockCatalogFileID)
	activeJSON := []byte(`{"file":{"props":{"attachment":{"activeFileId":"` + upperFileID + `"},"previewWidth":100}}}`)
	shared, err := contentv1.NormalizeContentStorageShared(
		"post",
		"file",
		activeJSON,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("normalize active shared storage: %v", err)
	}
	if got, want := string(shared.SharedData), `{"file":{"props":{"attachment":{"activeFileId":"`+blockCatalogFileID+`"}}}}`; got != want {
		t.Fatalf("canonical shared JSON = %s, want %s", got, want)
	}
	if len(shared.FileReferences) != 1 || shared.FileReferences[0].Missing || shared.FileReferences[0].FileID != blockCatalogFileID || shared.FileReferences[0].MissingAttachmentMediaKind != contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_UNSPECIFIED {
		t.Fatalf("active shared references = %#v", shared.FileReferences)
	}
	full, err := contentv1.NormalizeContentStorageBlock(
		"post",
		"file",
		activeJSON,
		[]byte(`{"file":{"props":{}}}`),
	)
	if err != nil {
		t.Fatalf("normalize full storage Block: %v", err)
	}
	if !reflect.DeepEqual(shared.FileReferences, full.FileReferences) {
		t.Fatalf("shared refs = %#v, full refs = %#v", shared.FileReferences, full.FileReferences)
	}

	missingJSON := []byte(`{"file":{"props":{"attachment":{"missingAttachment":{"formerFileId":"` + blockCatalogMissingFileID + `","mediaKind":"MISSING_ATTACHMENT_MEDIA_KIND_VIDEO"}}}}}`)
	if _, err := contentv1.NormalizeContentStorageShared("post", "file", missingJSON, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
		t.Fatal("WRITE shared normalizer accepted missing attachment")
	}
	missing, err := contentv1.NormalizeContentStorageShared("post", "file", missingJSON, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_RESTORE_SNAPSHOT)
	if err != nil {
		t.Fatalf("RESTORE shared normalizer rejected missing attachment: %v", err)
	}
	if len(missing.FileReferences) != 1 || !missing.FileReferences[0].Missing || missing.FileReferences[0].MissingAttachmentMediaKind != contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_VIDEO {
		t.Fatalf("missing shared references = %#v", missing.FileReferences)
	}

	failures := []struct {
		name, profile, kind, shared string
	}{
		{"unknown kind", "post", "unknown", `{"paragraph":{"props":{}}}`},
		{"forbidden profile kind", "compact", "map", `{"map":{"props":{}}}`},
		{"runtime field", "post", "paragraph", `{"paragraph":{"props":{"url":"https://runtime.invalid"}}}`},
		{"invalid shared field", "post", "file", `{"file":{"props":{}}}`},
	}
	for _, failure := range failures {
		if _, err := contentv1.NormalizeContentStorageShared(failure.profile, failure.kind, []byte(failure.shared), contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
			t.Fatalf("%s was accepted", failure.name)
		}
	}
}

func TestMutationFlattenerReusesSharedNormalizer(t *testing.T) {
	t.Parallel()

	batch := &contentv1.RichTextBlockMutationBatch{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		ExpectedRevision:        blockCatalogColumnID,
		ContributorMemberIds:    []string{blockCatalogUnitID},
		BaseMutations: []*contentv1.RichTextBlockMutation{{
			Operation: &contentv1.RichTextBlockMutation_Upsert{Upsert: &contentv1.UpsertRichTextBlock{Node: &contentv1.RichTextBlockNode{
				Block:     &contentv1.RichTextBlock{Id: blockCatalogTestID, Value: &contentv1.RichTextBlock_File{File: &contentv1.FileBlock{Props: &contentv1.FileProps{Attachment: missingAttachment(blockCatalogMissingFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_FILE)}}}},
				Placement: &contentv1.ContentBlockPlacement{},
			}}},
		}},
	}
	if _, err := contentv1.FlattenRichTextMutationBatchStorage(batch, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
		t.Fatal("WRITE mutation flattener bypassed shared missing-attachment validation")
	}
	storage, err := contentv1.FlattenRichTextMutationBatchStorage(batch, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_RESTORE_SNAPSHOT)
	if err != nil {
		t.Fatalf("RESTORE mutation flattener rejected valid shared payload: %v", err)
	}
	shared, err := contentv1.NormalizeContentStorageShared("post", "file", storage.BaseUpserts[0].SharedData, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_RESTORE_SNAPSHOT)
	if err != nil || string(shared.SharedData) != string(storage.BaseUpserts[0].SharedData) {
		t.Fatalf("flattener shared output did not reuse canonical seam: %v", err)
	}
}

func TestRewriteContentStorageMissingAttachmentsIsRecursiveAndImmutable(t *testing.T) {
	t.Parallel()

	rows := []contentv1.ContentStorageRow{
		shaderStorageRow(t),
		immersiveStorageRow(t),
	}
	original := cloneStorageRows(rows)
	rewritten, err := contentv1.RewriteContentStorageMissingAttachments(
		"page",
		"en",
		rows,
		map[string]contentv1.MissingAttachmentMediaKind{
			blockCatalogFileID: contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_VIDEO,
		},
	)
	if err != nil {
		t.Fatalf("rewrite missing attachments: %v", err)
	}
	if !reflect.DeepEqual(rows, original) {
		t.Fatal("rewrite mutated the original storage rows")
	}

	shader := &contentv1.RichTextBlockData{}
	if err := protojson.Unmarshal(rewritten[0].SharedData, shader); err != nil {
		t.Fatalf("decode rewritten shader: %v", err)
	}
	channels := shader.GetShader().GetProps().GetStages()[0].GetChannels()
	assertMissingAttachment(t, channels[0].GetFile(), blockCatalogFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_VIDEO)
	assertMissingAttachment(t, channels[1].GetFile(), blockCatalogMissingFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE)
	if got := channels[2].GetFile().GetActiveFileId(); got != blockCatalogPreservedFileID {
		t.Fatalf("unmapped active File = %q, want preserved", got)
	}

	immersive := &contentv1.PageSectionData{}
	if err := protojson.Unmarshal(rewritten[1].SharedData, immersive); err != nil {
		t.Fatalf("decode rewritten immersive section: %v", err)
	}
	props := immersive.GetImmersiveScene().GetUnits()[0].GetProps()
	assertMissingAttachment(t, props.GetMeshFile(), blockCatalogFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_VIDEO)
	assertMissingAttachment(t, props.GetTextureFile(), blockCatalogMissingFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE)
}

func TestSystemMutationFlattenersOwnEmptyContributorBoundary(t *testing.T) {
	t.Parallel()

	rich := &contentv1.RichTextBlockMutationBatch{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		ExpectedRevision:        blockCatalogColumnID,
		BaseMutations: []*contentv1.RichTextBlockMutation{{
			Operation: &contentv1.RichTextBlockMutation_Delete{Delete: &contentv1.DeleteRichTextBlock{BlockId: blockCatalogTestID}},
		}},
	}
	if _, err := contentv1.FlattenRichTextMutationBatchStorage(rich, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
		t.Fatal("public Rich Text flattener accepted empty contributors")
	}
	if _, err := contentv1.FlattenRichTextSystemMutationBatchStorage(rich, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err != nil {
		t.Fatalf("system Rich Text flattener rejected empty contributors: %v", err)
	}

	page := &contentv1.PageSectionMutationBatch{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		ExpectedRevision:        blockCatalogColumnID,
		BaseMutations: []*contentv1.PageSectionMutation{{
			Operation: &contentv1.PageSectionMutation_Delete{Delete: &contentv1.DeletePageSection{SectionId: blockCatalogSectionID}},
		}},
	}
	if _, err := contentv1.FlattenPageMutationBatchStorage(page, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err == nil {
		t.Fatal("public Page flattener accepted empty contributors")
	}
	if _, err := contentv1.FlattenPageSystemMutationBatchStorage(page, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE); err != nil {
		t.Fatalf("system Page flattener rejected empty contributors: %v", err)
	}
}

func TestLocaleMutationStorageCarriesExpectedKindAndValidatesProfile(t *testing.T) {
	t.Parallel()

	batch := &contentv1.RichTextBlockMutationBatch{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		ExpectedRevision:        blockCatalogColumnID,
		ContributorMemberIds:    []string{blockCatalogUnitID},
		LocaleMutationGroups: []*contentv1.RichTextLocaleMutationGroup{{
			Locale: "ko",
			Mutations: []*contentv1.RichTextBlockLocaleMutation{{
				Operation: &contentv1.RichTextBlockLocaleMutation_Upsert{Upsert: &contentv1.UpsertRichTextBlockLocale{
					Block: &contentv1.RichTextBlockLocale{
						BlockId: blockCatalogTestID,
						Value: &contentv1.RichTextBlockLocale_CodeBlock{CodeBlock: &contentv1.CodeBlockBlockLocale{
							Props:   &contentv1.CodeBlockLocaleProps{},
							Content: "const answer = 42",
						}},
					},
				}},
			}},
		}},
	}

	storage, err := contentv1.FlattenRichTextMutationBatchStorage(
		batch,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten locale mutation: %v", err)
	}
	upsert := storage.LocaleGroups[0].Upserts[0]
	if upsert.ExpectedKind != "code-block" {
		t.Fatalf("ExpectedKind = %q, want code-block", upsert.ExpectedKind)
	}
	if _, err := contentv1.NormalizeContentStorageLocale(
		"compact",
		upsert.ExpectedKind,
		upsert.LocalizedData,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err == nil {
		t.Fatal("compact profile accepted forbidden code-block locale payload")
	}
}

func TestLocaleMutationStoragePreservesExplicitDefaultPresence(t *testing.T) {
	t.Parallel()

	empty := ""
	pageBatch := func(caption *string) *contentv1.PageSectionMutationBatch {
		return &contentv1.PageSectionMutationBatch{
			BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
			ExpectedRevision:        blockCatalogColumnID,
			ContributorMemberIds:    []string{blockCatalogUnitID},
			LocaleMutationGroups: []*contentv1.PageLocaleMutationGroup{{
				Locale: "ko",
				Mutations: []*contentv1.PageSectionLocaleMutation{{
					Operation: &contentv1.PageSectionLocaleMutation_Upsert{Upsert: &contentv1.UpsertPageSectionLocale{
						Section: &contentv1.PageSectionLocale{
							SectionId: blockCatalogSectionID,
							Value: &contentv1.PageSectionLocale_ExternalVideo{ExternalVideo: &contentv1.ExternalVideoSectionLocale{
								Props: &contentv1.ExternalVideoSectionLocaleProps{Caption: caption},
							}},
						},
					}},
				}},
			}},
		}
	}

	assertPageCaptionPresence := func(t *testing.T, caption *string, wantPresent bool) {
		t.Helper()
		storage, err := contentv1.FlattenPageMutationBatchStorage(
			pageBatch(caption),
			contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
		)
		if err != nil {
			t.Fatalf("flatten Page locale mutation: %v", err)
		}
		localized := &contentv1.PageSectionLocaleData{}
		if err := protojson.Unmarshal(storage.LocaleGroups[0].Upserts[0].LocalizedData, localized); err != nil {
			t.Fatalf("decode Page locale mutation: %v", err)
		}
		gotPresent := localized.GetExternalVideo().GetProps().Caption != nil
		if gotPresent != wantPresent {
			t.Fatalf("Page caption presence = %t, want %t; JSON = %s", gotPresent, wantPresent, storage.LocaleGroups[0].Upserts[0].LocalizedData)
		}
	}

	t.Run("Page explicit empty", func(t *testing.T) { assertPageCaptionPresence(t, &empty, true) })
	t.Run("Page missing", func(t *testing.T) { assertPageCaptionPresence(t, nil, false) })

	richBatch := func(title *string) *contentv1.RichTextBlockMutationBatch {
		return &contentv1.RichTextBlockMutationBatch{
			BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
			Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
			ExpectedRevision:        blockCatalogColumnID,
			ContributorMemberIds:    []string{blockCatalogUnitID},
			LocaleMutationGroups: []*contentv1.RichTextLocaleMutationGroup{{
				Locale: "ko",
				Mutations: []*contentv1.RichTextBlockLocaleMutation{{
					Operation: &contentv1.RichTextBlockLocaleMutation_Upsert{Upsert: &contentv1.UpsertRichTextBlockLocale{
						Block: &contentv1.RichTextBlockLocale{
							BlockId: blockCatalogTestID,
							Value: &contentv1.RichTextBlockLocale_CodeBlock{CodeBlock: &contentv1.CodeBlockBlockLocale{
								Props: &contentv1.CodeBlockLocaleProps{Title: title},
							}},
						},
					}},
				}},
			}},
		}
	}

	assertRichTitlePresence := func(t *testing.T, title *string, wantPresent bool) {
		t.Helper()
		storage, err := contentv1.FlattenRichTextMutationBatchStorage(
			richBatch(title),
			contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
		)
		if err != nil {
			t.Fatalf("flatten Rich Text locale mutation: %v", err)
		}
		localized := &contentv1.RichTextBlockLocaleData{}
		if err := protojson.Unmarshal(storage.LocaleGroups[0].Upserts[0].LocalizedData, localized); err != nil {
			t.Fatalf("decode Rich Text locale mutation: %v", err)
		}
		gotPresent := localized.GetCodeBlock().GetProps().Title != nil
		if gotPresent != wantPresent {
			t.Fatalf("Rich Text title presence = %t, want %t; JSON = %s", gotPresent, wantPresent, storage.LocaleGroups[0].Upserts[0].LocalizedData)
		}
	}

	t.Run("Rich Text explicit empty", func(t *testing.T) { assertRichTitlePresence(t, &empty, true) })
	t.Run("Rich Text missing", func(t *testing.T) { assertRichTitlePresence(t, nil, false) })
}

func TestExplicitEmptyLocalePreservesStructureAndClearsOnlyTranslatableValues(t *testing.T) {
	t.Parallel()

	rich, err := contentv1.BuildExplicitEmptyContentStorageLocale(
		"post",
		"paragraph",
		[]byte(`{"paragraph":{"props":{},"content":[{"text":{"text":"Hello","styles":{"bold":true}}},{"hardBreak":{}},{"link":{"href":"https://example.test","content":[{"text":"World","styles":{"italic":true}}]}},{"mathInline":{"source":"x^2"}}]}}`),
	)
	if err != nil {
		t.Fatalf("build explicit-empty Rich Text locale: %v", err)
	}
	richLocale := &contentv1.RichTextBlockLocaleData{}
	if err := protojson.Unmarshal(rich, richLocale); err != nil {
		t.Fatalf("decode explicit-empty Rich Text locale: %v", err)
	}
	content := richLocale.GetParagraph().GetContent()
	if len(content) != 4 || content[0].GetText().GetText() != "" ||
		!content[0].GetText().GetStyles().GetBold() ||
		content[1].GetHardBreak() == nil ||
		content[2].GetLink().GetHref() != "https://example.test" ||
		content[2].GetLink().GetContent()[0].GetText() != "" ||
		!content[2].GetLink().GetContent()[0].GetStyles().GetItalic() ||
		content[3].GetMathInline().GetSource() != "x^2" {
		t.Fatalf("explicit-empty Rich Text locale changed protected structure: %s", rich)
	}

	code, err := contentv1.BuildExplicitEmptyContentStorageLocale(
		"post",
		"code-block",
		[]byte(`{"codeBlock":{"props":{"title":"Example"},"content":"const answer = 42"}}`),
	)
	if err != nil {
		t.Fatalf("build explicit-empty Code Block locale: %v", err)
	}
	codeLocale := &contentv1.RichTextBlockLocaleData{}
	if err := protojson.Unmarshal(code, codeLocale); err != nil {
		t.Fatalf("decode explicit-empty Code Block locale: %v", err)
	}
	if codeLocale.GetCodeBlock().GetProps().Title == nil ||
		codeLocale.GetCodeBlock().GetProps().GetTitle() != "" ||
		codeLocale.GetCodeBlock().GetContent() != "" {
		t.Fatalf("Code Block translatable values were not explicit empty: %s", code)
	}
	if !strings.Contains(string(code), `"content":""`) {
		t.Fatalf("Code Block explicit-empty content presence was lost: %s", code)
	}

	page, err := contentv1.BuildExplicitEmptyContentStorageLocale(
		"page",
		"immersive-scene",
		[]byte(`{"immersiveScene":{"props":{},"units":[{"unitId":"`+blockCatalogUnitID+`","props":{"title":"Title","text":"Body"}}]}}`),
	)
	if err != nil {
		t.Fatalf("build explicit-empty Page locale: %v", err)
	}
	pageLocale := &contentv1.PageSectionLocaleData{}
	if err := protojson.Unmarshal(page, pageLocale); err != nil {
		t.Fatalf("decode explicit-empty Page locale: %v", err)
	}
	unit := pageLocale.GetImmersiveScene().GetUnits()[0]
	if unit.GetUnitId() != blockCatalogUnitID || unit.GetProps().Title == nil ||
		unit.GetProps().Text == nil || unit.GetProps().GetTitle() != "" ||
		unit.GetProps().GetText() != "" {
		t.Fatalf("Page unit identity or explicit-empty values changed: %s", page)
	}
}

func TestLocaleStoragePreservesNestedPageAndEmptyContentShapes(t *testing.T) {
	t.Parallel()

	nested, err := contentv1.NormalizeContentStorageLocale(
		"page",
		"immersive-scene",
		[]byte(`{"immersiveScene":{"props":{},"units":[{"unitId":"`+blockCatalogUnitID+`","props":{"title":""}},{"unitId":"`+blockCatalogFileID+`","props":{}}]}}`),
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("normalize nested Page locale: %v", err)
	}
	page := &contentv1.PageSectionLocaleData{}
	if err := protojson.Unmarshal(nested, page); err != nil {
		t.Fatalf("decode nested Page locale: %v", err)
	}
	units := page.GetImmersiveScene().GetUnits()
	if len(units) != 2 || units[0].GetProps().Title == nil || units[1].GetProps().Title != nil {
		t.Fatalf("nested locale title presence was not preserved: %s", nested)
	}

	blank, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"paragraph",
		[]byte(`{"paragraph":{"props":{},"content":[]}}`),
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("normalize empty paragraph locale: %v", err)
	}
	paragraph := &contentv1.RichTextBlockLocaleData{}
	if err := protojson.Unmarshal(blank, paragraph); err != nil {
		t.Fatalf("decode empty paragraph locale: %v", err)
	}
	if paragraph.GetParagraph() == nil || len(paragraph.GetParagraph().GetContent()) != 0 {
		t.Fatalf("empty paragraph locale payload was not preserved: %s", blank)
	}
	if !strings.Contains(string(blank), `"content":[]`) {
		t.Fatalf("empty paragraph locale content presence was lost: %s", blank)
	}

	tableJSON := []byte(`{"table":{"props":{},"content":{"rows":[{"rowId":"` + blockCatalogColumnID + `","cells":[{"cellId":"` + blockCatalogSectionID + `","content":[]}]}]}}}`)
	table, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"table",
		tableJSON,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("normalize empty table cell locale: %v", err)
	}
	tableLocale := &contentv1.RichTextBlockLocaleData{}
	if err := protojson.Unmarshal(table, tableLocale); err != nil {
		t.Fatalf("decode empty table cell locale: %v", err)
	}
	rows := tableLocale.GetTable().GetContent().GetRows()
	if len(rows) != 1 || len(rows[0].GetCells()) != 1 || len(rows[0].GetCells()[0].GetContent()) != 0 {
		t.Fatalf("empty table cell locale payload was not preserved: %s", table)
	}
	if !strings.Contains(string(table), `"content":[]`) {
		t.Fatalf("empty table cell locale content presence was lost: %s", table)
	}
}

func TestLocaleStorageValidatesInlineStyleDescriptorsRecursively(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		kind      string
		localized string
		wantPath  string
	}{
		{
			name:      "styled text",
			kind:      "paragraph",
			localized: `{"paragraph":{"props":{},"content":[{"text":{"text":"invalid","styles":{"textColor":"RGB"}}}]}}`,
			wantPath:  "$.locale.content[0].text.styles.textColor",
		},
		{
			name:      "linked styled text",
			kind:      "paragraph",
			localized: `{"paragraph":{"props":{},"content":[{"link":{"href":"https://geul.io","content":[{"text":"invalid","styles":{"backgroundColor":"RGB"}}]}}]}}`,
			wantPath:  "$.locale.content[0].link.content[0].styles.backgroundColor",
		},
		{
			name:      "table cell styled text",
			kind:      "table",
			localized: `{"table":{"props":{},"content":{"rows":[{"rowId":"` + blockCatalogColumnID + `","cells":[{"cellId":"` + blockCatalogSectionID + `","content":[{"text":{"text":"invalid","styles":{"textColor":"RGB"}}}]}]}]}}}`,
			wantPath:  "$.locale.content.rows[0].cells[0].content[0].text.styles.textColor",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := contentv1.NormalizeContentStorageLocale(
				"post",
				test.kind,
				[]byte(test.localized),
				contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
			)
			if err == nil {
				t.Fatal("malformed inline style color was accepted")
			}
			var validationErr *contentv1.ContentBlockValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("error = %T %v, want ContentBlockValidationError", err, err)
			}
			if validationErr.Code != "color" || validationErr.Path != test.wantPath {
				t.Fatalf("validation issue = %#v, want color at %s", validationErr, test.wantPath)
			}
		})
	}

	normalized, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"paragraph",
		[]byte(`{"paragraph":{"props":{},"content":[{"text":{"text":"valid","styles":{"bold":true,"textColor":"#A1B2C3"}}},{"link":{"href":"https://geul.io","content":[{"text":"linked","styles":{"backgroundColor":"blue"}}]}}]}}`),
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("valid inline style color rejected: %v", err)
	}
	if !strings.Contains(string(normalized), `"textColor":"#a1b2c3"`) {
		t.Fatalf("inline style color was not canonicalized: %s", normalized)
	}
	if !strings.Contains(string(normalized), `"bold":true`) || !strings.Contains(string(normalized), `"href":"https://geul.io"`) {
		t.Fatalf("valid mark or nested link was not preserved: %s", normalized)
	}

	validTable := []byte(`{"table":{"props":{},"content":{"rows":[{"rowId":"` + blockCatalogColumnID + `","cells":[{"cellId":"` + blockCatalogSectionID + `","content":[{"text":{"text":"valid","styles":{"backgroundColor":"green"}}}]}]}]}}}`)
	if _, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"table",
		validTable,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err != nil {
		t.Fatalf("valid table inline content rejected: %v", err)
	}

	unsafeLink := []byte(`{"paragraph":{"props":{},"content":[{"link":{"href":"javascript:alert(1)","content":[{"text":"unsafe"}]}}]}}`)
	if _, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"paragraph",
		unsafeLink,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err == nil {
		t.Fatal("unsafe nested link was accepted")
	}
}

func TestLocaleStoragePreservesProfileSpecificInlineMath(t *testing.T) {
	t.Parallel()

	localized := []byte(`{"paragraph":{"props":{},"content":[{"mathInline":{"source":"x^2"}}]}}`)
	if _, err := contentv1.NormalizeContentStorageLocale(
		"post",
		"paragraph",
		localized,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err != nil {
		t.Fatalf("Post profile rejected inline math: %v", err)
	}
	if _, err := contentv1.NormalizeContentStorageLocale(
		"page",
		"paragraph",
		localized,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err == nil {
		t.Fatal("Page profile accepted inline math")
	}
}

func TestLocaleMutationFlattenerRejectsMalformedInlineStyleColor(t *testing.T) {
	t.Parallel()

	malformedColor := "RGB"
	batch := &contentv1.RichTextBlockMutationBatch{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		ExpectedRevision:        blockCatalogColumnID,
		ContributorMemberIds:    []string{blockCatalogUnitID},
		LocaleMutationGroups: []*contentv1.RichTextLocaleMutationGroup{{
			Locale: "ko",
			Mutations: []*contentv1.RichTextBlockLocaleMutation{{
				Operation: &contentv1.RichTextBlockLocaleMutation_Upsert{Upsert: &contentv1.UpsertRichTextBlockLocale{
					Block: &contentv1.RichTextBlockLocale{
						BlockId: blockCatalogTestID,
						Value: &contentv1.RichTextBlockLocale_Paragraph{Paragraph: &contentv1.ParagraphBlockLocale{
							Props: &contentv1.ParagraphLocaleProps{},
							Content: []*contentv1.RichTextInline{{Value: &contentv1.RichTextInline_Text{Text: &contentv1.RichTextStyledText{
								Text:   "invalid",
								Styles: &contentv1.RichTextStyle{TextColor: &malformedColor},
							}}}},
						}},
					},
				}},
			}},
		}},
	}

	if _, err := contentv1.FlattenRichTextMutationBatchStorage(
		batch,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	); err == nil {
		t.Fatal("locale mutation flattener accepted malformed inline style color")
	}
}

func TestPageColumnPlacementMustReferenceDeclaredColumn(t *testing.T) {
	t.Parallel()

	parent := []byte(`{"columns":{"props":{"columns":[{"id":"` + blockCatalogColumnID + `","ratio":1}]}}}`)
	if err := contentv1.ValidateContentStorageParent(
		"page",
		"columns",
		parent,
		"rich-text",
		nil,
		"column-"+blockCatalogColumnID,
	); err != nil {
		t.Fatalf("declared column rejected: %v", err)
	}
	if err := contentv1.ValidateContentStorageParent(
		"page",
		"columns",
		parent,
		"rich-text",
		nil,
		"column-00000000-0000-4000-8000-000000000003",
	); err == nil {
		t.Fatal("undeclared column placement was accepted")
	}
}

func TestPageStorageFlattensAndRebuildsInnerRichTextBlocks(t *testing.T) {
	t.Parallel()

	rich := paragraphDocument("nested")
	parentBlockID := blockCatalogTestID
	rich.Base.Nodes = append(rich.Base.Nodes, &contentv1.RichTextBlockNode{
		Block: &contentv1.RichTextBlock{
			Id: blockCatalogColumnID,
			Value: &contentv1.RichTextBlock_CodeBlock{CodeBlock: &contentv1.CodeBlockBlock{
				Props: &contentv1.CodeBlockProps{},
			}},
		},
		Placement: &contentv1.ContentBlockPlacement{ParentBlockId: &parentBlockID, Index: 0},
	})
	rich.LocaleOverlays[0].Blocks = append(rich.LocaleOverlays[0].Blocks, &contentv1.RichTextBlockLocale{
		BlockId: blockCatalogColumnID,
		Value: &contentv1.RichTextBlockLocale_CodeBlock{CodeBlock: &contentv1.CodeBlockBlockLocale{
			Props:   &contentv1.CodeBlockLocaleProps{},
			Content: "second nested block",
		}},
	})
	document := &contentv1.PageDocument{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		SourceLocale:            "en",
		Base: &contentv1.PageSectionGraph{Nodes: []*contentv1.PageSectionNode{{
			Section: &contentv1.PageSection{
				Id:       blockCatalogSectionID,
				Settings: &contentv1.PageSectionSettings{},
				Value: &contentv1.PageSection_RichText{RichText: &contentv1.RichTextSection{
					Props:  &contentv1.RichTextSectionProps{},
					Blocks: rich.Base,
				}},
			},
			Placement: &contentv1.PageSectionPlacement{Index: 0},
		}}},
		LocaleOverlays: []*contentv1.PageLocaleOverlay{{
			Locale: "en",
			Sections: []*contentv1.PageSectionLocale{{
				SectionId: blockCatalogSectionID,
				Value: &contentv1.PageSectionLocale_RichText{RichText: &contentv1.RichTextSectionLocale{
					Props:  &contentv1.RichTextSectionLocaleProps{},
					Blocks: rich.LocaleOverlays[0],
				}},
			}},
		}},
	}

	rows, err := contentv1.FlattenPageDocumentStorage(
		document,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten Page: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want section and two inner rich Blocks", len(rows))
	}
	richRows := map[string]*contentv1.ContentStorageRow{}
	for index := range rows {
		if rows[index].BlockID == blockCatalogTestID || rows[index].BlockID == blockCatalogColumnID {
			richRows[rows[index].BlockID] = &rows[index]
		}
	}
	expectedStorageParents := map[string]string{
		blockCatalogTestID:   blockCatalogSectionID,
		blockCatalogColumnID: blockCatalogTestID,
	}
	for id, expectedParent := range expectedStorageParents {
		richRow := richRows[id]
		if richRow == nil || richRow.ParentBlockID != expectedParent || richRow.ContainerSlot != "content" {
			t.Fatalf("inner rich Block %s has wrong placement: %#v", id, richRow)
		}
	}

	rebuilt, err := contentv1.MaterializePageDocumentStorage("en", rows)
	if err != nil {
		t.Fatalf("materialize Page: %v", err)
	}
	rebuiltRows, err := contentv1.FlattenPageDocumentStorage(
		rebuilt,
		contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE,
	)
	if err != nil {
		t.Fatalf("flatten rebuilt Page: %v", err)
	}
	originalHash, err := contentv1.ContentStorageCanonicalHash("page", rows)
	if err != nil {
		t.Fatalf("hash Page: %v", err)
	}
	rebuiltHash, err := contentv1.ContentStorageCanonicalHash("page", rebuiltRows)
	if err != nil {
		t.Fatalf("hash rebuilt Page: %v", err)
	}
	if originalHash != rebuiltHash {
		t.Fatalf("Page round-trip hash differs: %s != %s", originalHash, rebuiltHash)
	}
	if got := len(rebuilt.GetBase().GetNodes()[0].GetSection().GetRichText().GetBlocks().GetNodes()); got != 2 {
		t.Fatalf("rebuilt Page rich-text section has %d Blocks, want 2", got)
	}
	expectedGraphParents := map[string]string{
		blockCatalogTestID:   "",
		blockCatalogColumnID: blockCatalogTestID,
	}
	for _, node := range rebuilt.GetBase().GetNodes()[0].GetSection().GetRichText().GetBlocks().GetNodes() {
		if parent := node.GetPlacement().GetParentBlockId(); parent != expectedGraphParents[node.GetBlock().GetId()] {
			t.Fatalf("rebuilt Page rich-text Block %s has parent %s, want %s", node.GetBlock().GetId(), parent, expectedGraphParents[node.GetBlock().GetId()])
		}
	}
}

func paragraphDocument(text string) *contentv1.RichTextDocument {
	return &contentv1.RichTextDocument{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		SourceLocale:            "en",
		Base: &contentv1.RichTextBlockGraph{Nodes: []*contentv1.RichTextBlockNode{{
			Block: &contentv1.RichTextBlock{
				Id: blockCatalogTestID,
				Value: &contentv1.RichTextBlock_Paragraph{Paragraph: &contentv1.ParagraphBlock{
					Props: &contentv1.ParagraphProps{},
				}},
			},
			Placement: &contentv1.ContentBlockPlacement{Index: 0},
		}}},
		LocaleOverlays: []*contentv1.RichTextLocaleOverlay{{
			Locale: "en",
			Blocks: []*contentv1.RichTextBlockLocale{{
				BlockId: blockCatalogTestID,
				Value: &contentv1.RichTextBlockLocale_Paragraph{Paragraph: &contentv1.ParagraphBlockLocale{
					Props: &contentv1.ParagraphLocaleProps{},
					Content: []*contentv1.RichTextInline{{Value: &contentv1.RichTextInline_Text{
						Text: &contentv1.RichTextStyledText{Text: text},
					}}},
				}},
			}},
		}},
	}
}

func codeBlockDocument(text string) *contentv1.RichTextDocument {
	return &contentv1.RichTextDocument{
		BlockCatalogFingerprint: contentv1.ContentBlockCatalogFingerprint,
		Profile:                 contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST,
		SourceLocale:            "en",
		Base: &contentv1.RichTextBlockGraph{Nodes: []*contentv1.RichTextBlockNode{{
			Block: &contentv1.RichTextBlock{
				Id: blockCatalogTestID,
				Value: &contentv1.RichTextBlock_CodeBlock{CodeBlock: &contentv1.CodeBlockBlock{
					Props: &contentv1.CodeBlockProps{},
				}},
			},
			Placement: &contentv1.ContentBlockPlacement{Index: 0},
		}}},
		LocaleOverlays: []*contentv1.RichTextLocaleOverlay{{
			Locale: "en",
			Blocks: []*contentv1.RichTextBlockLocale{{
				BlockId: blockCatalogTestID,
				Value: &contentv1.RichTextBlockLocale_CodeBlock{CodeBlock: &contentv1.CodeBlockBlockLocale{
					Props:   &contentv1.CodeBlockLocaleProps{},
					Content: text,
				}},
			}},
		}},
	}
}

func shaderStorageRow(t *testing.T) contentv1.ContentStorageRow {
	t.Helper()

	stageKinds := []contentv1.ShaderProps_StagesItem_Kind{
		contentv1.ShaderProps_StagesItem_KIND_COMMON,
		contentv1.ShaderProps_StagesItem_KIND_VERTEX,
		contentv1.ShaderProps_StagesItem_KIND_BUFFER_A,
		contentv1.ShaderProps_StagesItem_KIND_BUFFER_B,
		contentv1.ShaderProps_StagesItem_KIND_BUFFER_C,
		contentv1.ShaderProps_StagesItem_KIND_BUFFER_D,
		contentv1.ShaderProps_StagesItem_KIND_CUBEMAP,
		contentv1.ShaderProps_StagesItem_KIND_SOUND,
		contentv1.ShaderProps_StagesItem_KIND_IMAGE,
	}
	stages := make([]*contentv1.ShaderProps_StagesItem, len(stageKinds))
	for index, kind := range stageKinds {
		channels := []*contentv1.ShaderProps_StagesItem_ChannelsItem{
			{Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_NONE},
			{Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_NONE},
			{Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_NONE},
			{Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_NONE},
		}
		if index == 0 {
			channels[0] = &contentv1.ShaderProps_StagesItem_ChannelsItem{
				Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_VIDEO_FILE,
				File: activeAttachment(blockCatalogFileID),
			}
			channels[1] = &contentv1.ShaderProps_StagesItem_ChannelsItem{
				Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_TEXTURE_FILE,
				File: missingAttachment(blockCatalogMissingFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE),
			}
			channels[2] = &contentv1.ShaderProps_StagesItem_ChannelsItem{
				Kind: contentv1.ShaderProps_StagesItem_ChannelsItem_KIND_TEXTURE_FILE,
				File: activeAttachment(blockCatalogPreservedFileID),
			}
		}
		stages[index] = &contentv1.ShaderProps_StagesItem{Kind: kind, Source: "void main() {}", Channels: channels}
	}
	shared := mustProtoJSON(t, &contentv1.RichTextBlockData{
		Value: &contentv1.RichTextBlockData_Shader{Shader: &contentv1.ShaderBlock{
			Props: &contentv1.ShaderProps{Stages: stages},
		}},
	})
	localized := mustProtoJSON(t, &contentv1.RichTextBlockLocaleData{
		Value: &contentv1.RichTextBlockLocaleData_Shader{Shader: &contentv1.ShaderBlockLocale{
			Props: &contentv1.ShaderLocaleProps{},
		}},
	})
	return contentv1.ContentStorageRow{
		BlockID:       blockCatalogTestID,
		ContainerSlot: "content",
		Kind:          "shader",
		SharedData:    shared,
		Locales:       []contentv1.ContentStorageLocale{{Locale: "en", LocalizedData: localized}},
	}
}

func immersiveStorageRow(t *testing.T) contentv1.ContentStorageRow {
	t.Helper()

	shared := mustProtoJSON(t, &contentv1.PageSectionData{
		Settings: &contentv1.PageSectionSettings{},
		Value: &contentv1.PageSectionData_ImmersiveScene{ImmersiveScene: &contentv1.ImmersiveSceneSection{
			Props: &contentv1.ImmersiveSceneSectionProps{},
			Units: []*contentv1.PageImmersiveUnit{{
				Id: blockCatalogUnitID,
				Props: &contentv1.PageImmersiveUnitProps{
					MeshFile:    activeAttachment(blockCatalogFileID),
					TextureFile: missingAttachment(blockCatalogMissingFileID, contentv1.MissingAttachmentMediaKind_MISSING_ATTACHMENT_MEDIA_KIND_IMAGE),
				},
			}},
		}},
	})
	localized := mustProtoJSON(t, &contentv1.PageSectionLocaleData{
		Value: &contentv1.PageSectionLocaleData_ImmersiveScene{ImmersiveScene: &contentv1.ImmersiveSceneSectionLocale{
			Props: &contentv1.ImmersiveSceneSectionLocaleProps{},
		}},
	})
	return contentv1.ContentStorageRow{
		BlockID:       blockCatalogSectionID,
		ContainerSlot: "sections",
		Kind:          "immersive-scene",
		SharedData:    shared,
		Locales:       []contentv1.ContentStorageLocale{{Locale: "en", LocalizedData: localized}},
	}
}

func activeAttachment(fileID string) *contentv1.FileAttachment {
	return &contentv1.FileAttachment{State: &contentv1.FileAttachment_ActiveFileId{ActiveFileId: fileID}}
}

func missingAttachment(fileID string, mediaKind contentv1.MissingAttachmentMediaKind) *contentv1.FileAttachment {
	return &contentv1.FileAttachment{State: &contentv1.FileAttachment_MissingAttachment{
		MissingAttachment: &contentv1.MissingAttachment{FormerFileId: fileID, MediaKind: mediaKind},
	}}
}

func assertMissingAttachment(t *testing.T, attachment *contentv1.FileAttachment, fileID string, mediaKind contentv1.MissingAttachmentMediaKind) {
	t.Helper()
	missing := attachment.GetMissingAttachment()
	if missing == nil || missing.GetFormerFileId() != fileID || missing.GetMediaKind() != mediaKind {
		t.Fatalf("attachment = %#v, want missing %s/%s", attachment, fileID, mediaKind)
	}
}

func mustProtoJSON(t *testing.T, message proto.Message) []byte {
	t.Helper()
	encoded, err := protojson.Marshal(message)
	if err != nil {
		t.Fatalf("marshal proto JSON: %v", err)
	}
	return encoded
}

func cloneStorageRows(rows []contentv1.ContentStorageRow) []contentv1.ContentStorageRow {
	result := make([]contentv1.ContentStorageRow, len(rows))
	for index, row := range rows {
		result[index] = row
		result[index].SharedData = append([]byte(nil), row.SharedData...)
		result[index].Locales = make([]contentv1.ContentStorageLocale, len(row.Locales))
		for localeIndex, locale := range row.Locales {
			result[index].Locales[localeIndex] = locale
			result[index].Locales[localeIndex].LocalizedData = append([]byte(nil), locale.LocalizedData...)
		}
	}
	return result
}
