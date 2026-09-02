package collaboration

import (
	"strings"
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
)

func validRelayInteractiveAIDocumentMutationRequest() *intrav1.RelayInteractiveAIDocumentMutationRequest {
	return &intrav1.RelayInteractiveAIDocumentMutationRequest{
		MutationId: "mutation-1",
		Origin:     intrav1.InteractiveAIDocumentMutationOrigin_INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP,
		Document: &managev1.AIDocumentReference{
			Domain:    managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_POST,
			Reference: "post-1",
		},
		Locale:                   &managev1.AIDocumentLocale{Code: "ko"},
		ExpectedDocumentRevision: "revision-before",
		AcceptedDocumentRevision: "revision-after",
		Operations: []*managev1.AIDocumentOperation{
			{
				Operation: &managev1.AIDocumentOperation_DeleteBlock{
					DeleteBlock: &managev1.AIDocumentDeleteBlockOperation{BlockHandle: "block-1"},
				},
			},
		},
		ActorMemberId: "member-1",
	}
}

func TestValidateRelayInteractiveAIDocumentMutationRequest(t *testing.T) {
	t.Parallel()

	if err := ValidateRelayInteractiveAIDocumentMutationRequest(nil); err == nil {
		t.Fatal("nil request must fail closed")
	}

	for _, origin := range []intrav1.InteractiveAIDocumentMutationOrigin{
		intrav1.InteractiveAIDocumentMutationOrigin_INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_IN_EDITOR_AI,
		intrav1.InteractiveAIDocumentMutationOrigin_INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_MCP,
	} {
		request := validRelayInteractiveAIDocumentMutationRequest()
		request.Origin = origin
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Errorf("active origin %s: %v", origin, err)
		}
	}
}

func TestValidateRelayInteractiveAIDocumentMutationRequestRejectsMalformedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*intrav1.RelayInteractiveAIDocumentMutationRequest)
		want   string
	}{
		{name: "blank mutation id", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.MutationId = "  " }, want: "id"},
		{name: "unspecified origin", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Origin = intrav1.InteractiveAIDocumentMutationOrigin_INTERACTIVE_AI_DOCUMENT_MUTATION_ORIGIN_UNSPECIFIED
		}, want: "origin"},
		{name: "unknown origin", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Origin = intrav1.InteractiveAIDocumentMutationOrigin(99)
		}, want: "origin"},
		{name: "missing document", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.Document = nil }, want: "document"},
		{name: "unspecified domain", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Document.Domain = managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_UNSPECIFIED
		}, want: "domain"},
		{name: "unknown domain", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Document.Domain = managev1.AIDocumentDomain(99)
		}, want: "domain"},
		{name: "blank reference", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.Document.Reference = "  " }, want: "reference"},
		{name: "missing locale", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.Locale = nil }, want: "locale"},
		{name: "blank locale", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.Locale.Code = "  " }, want: "locale"},
		{name: "blank expected document revision", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.ExpectedDocumentRevision = "  "
		}, want: "expected document revision"},
		{name: "blank accepted document revision", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.AcceptedDocumentRevision = "  "
		}, want: "accepted document revision"},
		{name: "blank expected target revision", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			value := "  "
			request.ExpectedTargetRevision = &value
		}, want: "expected target revision"},
		{name: "blank accepted target revision", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			value := "  "
			request.AcceptedTargetRevision = &value
		}, want: "accepted target revision"},
		{name: "target revision disappears outside delete", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			value := "target-before"
			request.ExpectedTargetRevision = &value
			request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		}, want: "cannot disappear"},
		{name: "source mutation preserves document revision", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		}, want: "must advance"},
		{name: "missing operations", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.Operations = nil }, want: "operations"},
		{name: "nil operation", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Operations = []*managev1.AIDocumentOperation{nil}
		}, want: "operation 0"},
		{name: "missing operation arm", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Operations = []*managev1.AIDocumentOperation{{}}
		}, want: "operation 0"},
		{name: "mixed locale lifecycle operation", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) {
			request.Operations = append(request.Operations, &managev1.AIDocumentOperation{Operation: &managev1.AIDocumentOperation_CreateTranslation{CreateTranslation: &managev1.AIDocumentCreateTranslationOperation{}}})
		}, want: "must be exclusive"},
		{name: "blank actor member id", mutate: func(request *intrav1.RelayInteractiveAIDocumentMutationRequest) { request.ActorMemberId = "  " }, want: "actor member id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := validRelayInteractiveAIDocumentMutationRequest()
			test.mutate(request)
			err := ValidateRelayInteractiveAIDocumentMutationRequest(request)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestValidateRelayInteractiveAIDocumentMutationRequestTargetRevisionTransitions(t *testing.T) {
	t.Parallel()

	t.Run("source mutation", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("existing target value write", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		expected := "target-before"
		accepted := "target-after"
		request.ExpectedTargetRevision = &expected
		request.AcceptedTargetRevision = &accepted
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("missing target first write", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		accepted := "target-created"
		request.AcceptedTargetRevision = &accepted
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("explicit create", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		accepted := "target-created"
		request.AcceptedTargetRevision = &accepted
		request.Operations = []*managev1.AIDocumentOperation{{Operation: &managev1.AIDocumentOperation_CreateTranslation{CreateTranslation: &managev1.AIDocumentCreateTranslationOperation{}}}}
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("explicit delete", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		request.AcceptedDocumentRevision = request.ExpectedDocumentRevision
		expected := "target-before"
		request.ExpectedTargetRevision = &expected
		request.Operations = []*managev1.AIDocumentOperation{{Operation: &managev1.AIDocumentOperation_DeleteTranslation{DeleteTranslation: &managev1.AIDocumentDeleteTranslationOperation{}}}}
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("target write changing document revision", func(t *testing.T) {
		request := validRelayInteractiveAIDocumentMutationRequest()
		expected := "target-before"
		accepted := "target-after"
		request.ExpectedTargetRevision = &expected
		request.AcceptedTargetRevision = &accepted
		if err := ValidateRelayInteractiveAIDocumentMutationRequest(request); err == nil || !strings.Contains(err.Error(), "preserve the document revision") {
			t.Fatalf("error = %v", err)
		}
	})
}
