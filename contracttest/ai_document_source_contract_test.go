package contracttest_test

import "testing"

func TestAIDocumentSourceOwnsDCDPMetadataAndMutations(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	domains := sourceDeclaration(t, source, "enum", "AIDocumentDomain")
	requireSourceContains(t, domains,
		"AI_DOCUMENT_DOMAIN_POST = 1;",
		"AI_DOCUMENT_DOMAIN_PAGE = 2;",
		"AI_DOCUMENT_DOMAIN_WORK = 3;",
		"AI_DOCUMENT_DOMAIN_PROGRAM_EVENT = 4;",
		"AI_DOCUMENT_DOMAIN_RELEASE = 5;",
		"AI_DOCUMENT_DOMAIN_ARTIST = 6;",
		"AI_DOCUMENT_DOMAIN_LABEL = 7;",
		"AI_DOCUMENT_DOMAIN_MENU = 8;",
		"AI_DOCUMENT_DOMAIN_EMAIL_TEMPLATE = 9;",
		"AI_DOCUMENT_DOMAIN_EMAIL_LAYOUT = 10;",
		"AI_DOCUMENT_DOMAIN_CAMPAIGN = 11;",
		"AI_DOCUMENT_DOMAIN_FORM = 12;",
		"AI_DOCUMENT_DOMAIN_PRIVACY = 13;",
		"AI_DOCUMENT_DOMAIN_TERMS = 14;",
		"AI_DOCUMENT_DOMAIN_POST_SERIES = 15;",
	)
	service := sourceDeclaration(t, source, "service", "AIDocumentService")
	requireSourceContains(t, service,
		"rpc OpenAIDocument(",
		"rpc ApplyAIDocumentOperations(",
		"Source-locale operations may change graph, relations, shared fields",
		"Non-source operations may set locale-owned translatable",
		"first field set implicitly creates",
		"server-side seeding",
		"same transaction",
	)
	metadata := sourceDeclaration(t, source, "message", "AIDocumentMetadata")
	requireSourceContains(t, metadata,
		"string protocol_version = 1;",
		"AIDocumentReference document = 2;",
		"string catalog_fingerprint = 3;",
		"string document_revision = 4;",
		"AIDocumentLocale source_locale = 5;",
		"AIDocumentLocale requested_locale = 6;",
		"AIDocumentLocaleRole locale_role = 7;",
		"bool locale_exists = 8;",
		"optional string target_revision = 9;",
		"Opaque persisted owning-domain revision for the shared/source graph",
		"Target-only writes never advance it",
	)
	requireSourceExcludes(t, source,
		"rpc ReadAIDocument(",
		"rpc ValidateAIDocumentOperations(",
		"message ReadAIDocumentRequest",
		"message ReadAIDocumentResponse",
		"message ValidateAIDocumentOperationsRequest",
		"message ValidateAIDocumentOperationsResponse",
	)
}

func TestAIDocumentSourceHasStableTypedValuesRelationsAndOperations(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	requireSourceContains(t, source,
		"Explicitly creates one missing non-source locale",
		"every locale-owned stable unit from the current source locale",
		"the client never submits seed values or structure",
		"Deletes one exact existing non-source locale",
		"The source locale cannot be",
		"deleted and the client must supply",
	)
	value := sourceDeclaration(t, source, "message", "AIDocumentValue")
	requireSourceContains(t, value,
		"oneof value",
		"string text = 1;",
		"bool boolean = 2;",
		"string number = 3;",
		"AIDocumentInlineContent inline = 4;",
		"AIDocumentListValue list = 5;",
		"AIDocumentObjectValue object = 6;",
		"text=\"\" is explicit empty",
	)
	listItem := sourceDeclaration(t, source, "message", "AIDocumentListItem")
	requireSourceContains(t, listItem,
		"string item_handle = 1;",
		"AIDocumentValue value = 2;",
	)
	fieldTarget := sourceDeclaration(t, source, "message", "AIDocumentFieldTarget")
	requireSourceContains(t, fieldTarget,
		"repeated AIDocumentFieldPathSegment path = 4;",
		"Whole-list positional arrays do not expose item mutations",
	)
	mark := sourceDeclaration(t, source, "message", "AIDocumentInlineMark")
	requireSourceContains(t, mark, "optional AIDocumentValue parameter = 3;")
	relation := sourceDeclaration(t, source, "message", "AIDocumentRelationItemReference")
	requireSourceContains(t, relation,
		"string block_handle = 1;",
		"string relation_handle = 2;",
		"string item_handle = 3;",
	)
	operation := sourceDeclaration(t, source, "message", "AIDocumentOperation")
	requireSourceContains(t, operation,
		"AIDocumentSetFieldOperation set_field = 1;",
		"AIDocumentUnsetFieldOperation unset_field = 2;",
		"AIDocumentInsertBlockOperation insert_block = 3;",
		"AIDocumentDeleteBlockOperation delete_block = 4;",
		"AIDocumentMoveBlockOperation move_block = 5;",
		"AIDocumentReplaceBlockKindOperation replace_block_kind = 6;",
		"AIDocumentAttachFileOperation attach_file = 7;",
		"AIDocumentDetachFileOperation detach_file = 8;",
		"AIDocumentCreateTranslationOperation create_translation = 9;",
		"AIDocumentDeleteTranslationOperation delete_translation = 10;",
		"AIDocumentInsertRelationItemOperation insert_relation_item = 11;",
		"AIDocumentDeleteRelationItemOperation delete_relation_item = 12;",
		"AIDocumentMoveRelationItemOperation move_relation_item = 13;",
	)
	for _, relationOperation := range []string{
		"AIDocumentInsertRelationItemOperation",
		"AIDocumentDeleteRelationItemOperation",
		"AIDocumentMoveRelationItemOperation",
	} {
		declaration := sourceDeclaration(t, source, "message", relationOperation)
		requireSourceExcludes(t, declaration, "json_path", "array_index", "item_index", "position")
	}
}

func TestAIDocumentSourceCarriesCASIssuesConflictsAndAcceptedChanges(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	mutation := sourceDeclaration(t, source, "message", "AIDocumentMutation")
	requireSourceContains(t, mutation,
		"string protocol_version = 1;",
		"string expected_document_revision = 4;",
		"repeated AIDocumentOperation operations = 5;",
		"optional string expected_target_revision = 6;",
		"unrelated target-locale writes do",
		"not invalidate this token",
	)
	issue := sourceDeclaration(t, source, "message", "AIDocumentIssue")
	requireSourceContains(t, issue,
		"uint32 operation_index = 1;",
		"AIDocumentIssueCode code = 2;",
		"optional string handle = 3;",
	)
	conflict := sourceDeclaration(t, source, "message", "AIDocumentConflict")
	requireSourceContains(t, conflict,
		"AIDocumentConflictCode code = 1;",
		"string current_document_revision = 2;",
		"repeated string affected_handles = 3;",
		"optional string current_target_revision = 4;",
	)
	accepted := sourceDeclaration(t, source, "message", "AIDocumentAcceptedMutation")
	requireSourceContains(t, accepted,
		"string document_revision = 1;",
		"repeated AIDocumentAcceptedChange changes = 2;",
		"optional string target_revision = 3;",
	)
	change := sourceDeclaration(t, source, "message", "AIDocumentAcceptedChange")
	requireSourceContains(t, change,
		"uint32 operation_index = 1;",
		"AIDocumentOperationKind kind = 2;",
		"repeated string affected_handles = 3;",
	)
}

func TestAIDocumentSourceExcludesEditorAndBinaryRepresentations(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	requireSourceExcludes(t, source,
		"AIChatTarget",
		"AIEditorSelection",
		"AIEditorRequest",
		"AIEditorEdit",
		"AIChatRequest",
		"AIChatResponse",
		"rpc Chat(",
		"selected_html",
		"source_blocks_html",
		"context_html",
		"preview_html",
		"tiptap",
		"prosemirror",
		"yjs",
		"xliff",
		"base64",
		"bytes payload",
		"generic json",
	)
}
