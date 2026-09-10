package contracttest_test

import (
	"strings"
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestGatewayAuthorizationIsAClosedCoarseBoundary(t *testing.T) {
	request := (&intrav1.AuthorizeGatewayAccessRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 3)
	requireMessageField(t, request, "account_identity_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, request, "session_id", 2, protoreflect.StringKind, "")
	requireMessageField(t, request, "role", 3, protoreflect.EnumKind, "")

	roleField := request.Fields().ByName("role")
	if got, want := roleField.Enum().FullName(), policyv1.AuthorizationRole(0).Descriptor().FullName(); got != want {
		t.Fatalf("gateway role enum = %s, want shared %s", got, want)
	}
	role := roleField.Enum()
	if role.Values().Len() != 5 {
		t.Fatalf("AuthorizationRole must expose unspecified plus anon, user, author, and admin, got %d values", role.Values().Len())
	}
	for _, name := range []protoreflect.Name{
		"UNSPECIFIED",
		"ANON",
		"USER",
		"AUTHOR",
		"ADMIN",
	} {
		if role.Values().ByName(name) == nil {
			t.Errorf("AuthorizationRole.%s is missing", name)
		}
	}

	response := (&intrav1.AuthorizeGatewayAccessResponse{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, response, 0)

	service := intrav1.File_api_intra_v1_gateway_authorization_proto.Services().ByName("InternalGatewayAuthorizationService")
	if service == nil || service.Methods().ByName("AuthorizeGatewayAccess") == nil {
		t.Fatal("InternalGatewayAuthorizationService.AuthorizeGatewayAccess is missing")
	}
}

func TestGatewayAuthorizationJSONUsesShortAuthorizationRoleNames(t *testing.T) {
	for _, role := range []policyv1.AuthorizationRole{policyv1.AuthorizationRole_AUTHOR, policyv1.AuthorizationRole_ADMIN} {
		payload, err := protojson.Marshal(&intrav1.AuthorizeGatewayAccessRequest{Role: role})
		if err != nil {
			t.Fatalf("marshal %s gateway request: %v", role, err)
		}
		want := `"role":"` + role.String() + `"`
		if !strings.Contains(string(payload), want) {
			t.Fatalf("gateway request JSON = %s, want %s", payload, want)
		}
	}
}
