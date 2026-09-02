package contracttest_test

import "testing"

func TestAccountSourceOwnsPersonalAccessTokenLifecycle(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/account.proto")
	service := sourceDeclaration(t, source, "service", "AccountService")
	requireSourceContains(t, service,
		"rpc CreateMyPersonalAccessToken(",
		"rpc ListMyPersonalAccessTokens(",
		"rpc RegenerateMyPersonalAccessToken(",
		"rpc DeleteMyPersonalAccessToken(",
		"rpc CreateAccountPersonalAccessToken(",
		"rpc ListAccountPersonalAccessTokens(",
		"rpc RegenerateAccountPersonalAccessToken(",
		"rpc DeleteAccountPersonalAccessToken(",
		"normal authenticated browser Session",
		"Browser Session only",
		"Account freshness",
		"machine bearer",
		"Admin browser Session only",
	)
	requireSourceExcludes(t, source,
		"PersonalAccessTokenCapability",
		"PersonalAccessTokenScope",
		"personal_access_token_read_scope",
		"personal_access_token_write_scope",
		"mcp_read",
		"mcp_write",
	)
}

func TestAccountPersonalAccessTokenAdminRequestsTargetOneMember(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/account.proto")
	for _, name := range []string{
		"CreateAccountPersonalAccessTokenRequest",
		"ListAccountPersonalAccessTokensRequest",
		"RegenerateAccountPersonalAccessTokenRequest",
		"DeleteAccountPersonalAccessTokenRequest",
	} {
		request := sourceDeclaration(t, source, "message", name)
		requireSourceContains(t, request, "string member_id = 1;")
		requireSourceExcludes(t, request, "name", "capabilities", "mcp")
	}
}

func TestPersonalAccessTokenIsSingularAndSecretsAreOneTimeOnly(t *testing.T) {
	source := readSourceProto(t, "api/manage/v1/account.proto")
	token := sourceDeclaration(t, source, "message", "PersonalAccessToken")
	requireSourceContains(t, token,
		"string id = 1;",
		"google.protobuf.Timestamp created_at = 2;",
	)
	requireSourceExcludes(t, token, "secret", "read_scope", "write_scope", "reserved", "name", "capabilities")

	create := sourceDeclaration(t, source, "message", "CreateMyPersonalAccessTokenResponse")
	regenerate := sourceDeclaration(t, source, "message", "RegenerateMyPersonalAccessTokenResponse")
	list := sourceDeclaration(t, source, "message", "ListMyPersonalAccessTokensResponse")
	requireSourceContains(t, create, "string secret = 2;", "Returned exactly once")
	requireSourceContains(t, regenerate, "string secret = 2;", "Returned exactly once")
	requireSourceExcludes(t, list, "secret")

	createRequest := sourceDeclaration(t, source, "message", "CreateMyPersonalAccessTokenRequest")
	requireSourceExcludes(t, createRequest, "reserved", "name", "capabilities")
}
