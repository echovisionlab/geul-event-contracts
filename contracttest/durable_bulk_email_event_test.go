package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	eventpkg "github.com/echovisionlab/geul-event-contracts/go/event"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSendEmailRequestMatchesIdentityCourierHTTPDataModel(t *testing.T) {
	request := (&intrav1.SendEmailRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 3)

	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
		kind   protoreflect.Kind
	}{
		{name: "recipient", number: 1, kind: protoreflect.StringKind},
		{name: "template_type", number: 2, kind: protoreflect.StringKind},
		{name: "template_data", number: 3, kind: protoreflect.MessageKind},
	} {
		requireMessageField(t, request, expected.name, expected.number, expected.kind, "")
	}
}

func TestSendEmailEventUsesOnlyTypedRecipientAuthority(t *testing.T) {
	event := (&managev1.SendEmailEvent{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, event, 17)

	for _, expected := range []struct {
		name    protoreflect.Name
		number  protoreflect.FieldNumber
		kind    protoreflect.Kind
		message protoreflect.FullName
	}{
		{name: "delivery_recipient_id", number: 7, kind: protoreflect.StringKind},
		{name: "message_id", number: 8, kind: protoreflect.StringKind},
		{name: "issuance_id", number: 9, kind: protoreflect.StringKind},
		{
			name:    "expires_at",
			number:  10,
			kind:    protoreflect.MessageKind,
			message: "google.protobuf.Timestamp",
		},
	} {
		field := requireMessageField(
			t,
			event,
			expected.name,
			expected.number,
			expected.kind,
			expected.message,
		)
		if !field.HasPresence() {
			t.Errorf("%s.%s must preserve optional presence", event.FullName(), field.Name())
		}
	}

	for _, expected := range []struct {
		name    protoreflect.Name
		number  protoreflect.FieldNumber
		message protoreflect.FullName
		inOneof bool
	}{
		{
			name:    "account_selected_primary_email",
			number:  11,
			message: "api.manage.v1.AccountSelectedPrimaryEmailRecipient",
			inOneof: true,
		},
		{
			name:    "newsletter_subscription",
			number:  15,
			message: "api.manage.v1.NewsletterSubscriptionRecipient",
			inOneof: true,
		},
		{
			name:    "auth_login",
			number:  16,
			message: "api.manage.v1.AuthLoginRecipient",
			inOneof: true,
		},
		{
			name:    "auth_registration",
			number:  17,
			message: "api.manage.v1.AuthRegistrationRecipient",
			inOneof: true,
		},
	} {
		field := requireMessageField(
			t,
			event,
			expected.name,
			expected.number,
			protoreflect.MessageKind,
			expected.message,
		)
		if expected.inOneof {
			if oneof := field.ContainingOneof(); oneof == nil || oneof.Name() != "recipient_context" {
				t.Errorf("%s.%s must remain in the recipient_context oneof", event.FullName(), field.Name())
			}
		}
	}
}

func TestAuthRecipientContextsKeepIdentityBoundary(t *testing.T) {
	login := (&managev1.AuthLoginRecipient{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, login, 2)
	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{name: "identity_id", number: 1},
		{name: "target_email", number: 2},
	} {
		field := requireMessageField(t, login, expected.name, expected.number, protoreflect.StringKind, "")
		if field.HasPresence() {
			t.Errorf("%s.%s must be required by the application contract", login.FullName(), field.Name())
		}
	}

	registration := (&managev1.AuthRegistrationRecipient{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, registration, 1)
	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{name: "target_email", number: 1},
	} {
		field := requireMessageField(t, registration, expected.name, expected.number, protoreflect.StringKind, "")
		if field.HasPresence() {
			t.Errorf("%s.%s must be required by the application contract", registration.FullName(), field.Name())
		}
	}
}

func TestSendBulkEmailBatchEventIsDurableRunOnly(t *testing.T) {
	if eventpkg.QueueEmailCampaign != "email.campaign" {
		t.Fatalf("campaign queue = %q, want email.campaign", eventpkg.QueueEmailCampaign)
	}

	event := (&managev1.SendBulkEmailBatchEvent{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, event, 3)

	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
		kind   protoreflect.Kind
	}{
		{name: "delivery_run_id", number: 1, kind: protoreflect.StringKind},
		{name: "batch_size", number: 2, kind: protoreflect.Int32Kind},
		{name: "rate_per_second", number: 3, kind: protoreflect.Int32Kind},
	} {
		field := requireMessageField(t, event, expected.name, expected.number, expected.kind, "")
		if field.HasPresence() {
			t.Errorf("%s.%s must be non-optional", event.FullName(), expected.name)
		}
	}
}
