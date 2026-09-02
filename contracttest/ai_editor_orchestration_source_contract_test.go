package contracttest_test

import "testing"

func TestAIEditorOrchestrationSourceOwnsTransientBrowserSessionTurn(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	request := sourceDeclaration(t, source, "message", "StartAIEditorTurnRequest")
	requireSourceContains(t, request,
		"AIDocumentReference document = 1;",
		"AIDocumentLocale locale = 2;",
		"string expected_document_revision = 3;",
		"repeated string block_handles = 4;",
		"string action = 5;",
		"optional string prompt = 6;",
		"optional string expected_target_revision = 7;",
		"Stable Block handles only",
	)
	service := sourceDeclaration(t, source, "service", "AIEditorOrchestrationService")
	requireSourceContains(t, service,
		"rpc StartAIEditorTurn(StartAIEditorTurnRequest) returns (stream AIEditorTurnEvent)",
		"rpc ResolveAIEditorToolCall(ResolveAIEditorToolCallRequest) returns (ResolveAIEditorToolCallResponse)",
		"rpc CancelAIEditorTurn(CancelAIEditorTurnRequest) returns (CancelAIEditorTurnResponse)",
		"option (api.policy.v1.access) = {role: USER};",
		"server-configured provider loop",
		"session-local undo boundary",
	)
	requireSourceExcludes(t, service, "MCP", "PAT", "OAuth", "bearer", "provider credential")
}

func TestAIEditorOrchestrationSourceStreamsTypedApprovalAndTerminalOutcome(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	event := sourceDeclaration(t, source, "message", "AIEditorTurnEvent")
	requireSourceContains(t, event,
		"string turn_id = 1;",
		"AIEditorAssistantTextDelta assistant_text = 2;",
		"AIEditorTurnPhaseUpdate phase = 3;",
		"AIEditorDocumentToolApprovalRequired approval_required = 4;",
		"AIEditorDocumentToolResult document_result = 5;",
		"AIEditorTurnTerminalOutcome terminal = 6;",
	)
	approval := sourceDeclaration(t, source, "message", "AIEditorDocumentToolApprovalRequired")
	requireSourceContains(t, approval,
		"string tool_call_id = 1;",
		"AIDocumentMutation mutation = 2;",
		"optional string summary = 3;",
		"closed DCDP mutation",
	)
	resolution := sourceDeclaration(t, source, "message", "ResolveAIEditorToolCallRequest")
	requireSourceContains(t, resolution,
		"string turn_id = 1;",
		"string tool_call_id = 2;",
		"AIEditorToolCallDecision decision = 3;",
	)
	result := sourceDeclaration(t, source, "message", "AIEditorDocumentToolResult")
	requireSourceContains(t, result,
		"AIDocumentAcceptedMutation accepted = 2;",
		"AIDocumentValidation rejected = 3;",
	)
}

func TestAIEditorOrchestrationSourceExcludesOldEditorAndDurableConversationPayloads(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/ai.proto")
	start := sourceDeclaration(t, source, "message", "StartAIEditorTurnRequest")
	event := sourceDeclaration(t, source, "message", "AIEditorTurnEvent")
	for _, declaration := range []string{start, event} {
		requireSourceExcludes(t, declaration,
			"html",
			"tiptap",
			"prosemirror",
			"yjs",
			"position",
			"conversation",
			"history",
			"messages",
			"provider",
			"model",
			"credential",
			"token",
			"pat",
			"mcp",
		)
	}
}
