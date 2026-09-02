import { namespaceName, relationshipMethodName } from "./render-go.mjs";
import { pascal } from "./catalog-source.mjs";

const quote = (value) => JSON.stringify(value);

function renderDomainSurface(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const calls = [
    `assertResource(t, ${quote(`${namespace}.Resource`)}, ${quote(definitionName)}, resourceID, func() (policyv1.Resource, error) { return policyv1.${namespace}.Resource(resourceID) })`,
  ];

  for (const permission of model.lookupPermissions[definitionName]) {
    calls.push(
      `assertResourceLookup(t, ${quote(`${namespace}.Lookup${pascal(permission)}`)}, ${quote(definitionName)}, ${quote(permission)}, policyv1.${namespace}.Lookup${pascal(permission)}())`,
    );
  }

  for (const [action, permission] of [
    ...Object.entries(model.actions[definitionName]),
    ...Object.entries(model.actionAliases[definitionName]),
  ]) {
    calls.push(
      `assertCan(t, ${quote(`${namespace}.${pascal(action)}`)}, ${quote(definitionName)}, resourceID, ${quote(action)}, ${quote(permission)}, func() (policyv1.Can, error) { return policyv1.${namespace}.${pascal(action)}(resourceID) })`,
    );
  }

  for (const [action, permission] of Object.entries(
    model.platformActions[definitionName],
  )) {
    calls.push(
      `assertCan(t, ${quote(`${namespace}.${pascal(action)}`)}, "platform", "global", ${quote(`${definitionName}.${action}`)}, ${quote(permission)}, func() (policyv1.Can, error) { return policyv1.${namespace}.${pascal(action)}() })`,
    );
  }

  for (const [relation, subjects] of Object.entries(
    model.definitions[definitionName].relations,
  )) {
    for (const subject of subjects) {
      let subjectType;
      let subjectID;
      let subjectRelation = "";
      let argument = "";
      if (subject === "platform") {
        subjectType = "platform";
        subjectID = "global";
      } else if (subject === "account_identity") {
        subjectType = "account_identity";
        subjectID = "account-identity-id";
        argument = ", actor";
      } else {
        [subjectType, subjectRelation = ""] = subject.split("#");
        subjectID = `${relation}-subject-id`;
        argument = `, ${quote(subjectID)}`;
      }
      for (const operation of ["touch", "delete"]) {
        const method = relationshipMethodName(
          model,
          definitionName,
          relation,
          subject,
          operation,
        );
        const operationConstant =
          operation === "touch"
            ? "policyv1.RelationshipTouch"
            : "policyv1.RelationshipDelete";
        calls.push(
          `assertRelationship(t, ${quote(`${namespace}.${method}`)}, ${quote(definitionName)}, resourceID, ${quote(relation)}, ${quote(subjectType)}, ${quote(subjectID)}, ${quote(subjectRelation)}, ${operationConstant}, func() (policyv1.RelationshipMutation, error) { return policyv1.${namespace}.${method}(resourceID${argument}) })`,
        );
      }
    }
  }

  return calls.map((call) => `\t${call}`).join("\n");
}

function renderRoleSurface(model) {
  const calls = [];
  for (const role of model.roles) {
    const roleMethod = pascal(role);
    calls.push(
      `assertRole(t, ${quote(`Role.${roleMethod}`)}, ${quote(role)}, policyv1.Role.${roleMethod}())`,
      `parsed${roleMethod}, ok := policyv1.Role.Parse(${quote(role)})`,
      `if !ok { t.Fatalf(${quote(`Role.Parse(${role}) rejected a generated role`)}) }`,
      `assertRole(t, ${quote(`Role.Parse(${role})`)}, ${quote(role)}, parsed${roleMethod})`,
    );
  }
  calls.push(
    `assertRelationship(t, "Role.TouchMember", "role", "admin", "member", "account_identity", "account-identity-id", "", policyv1.RelationshipTouch, func() (policyv1.RelationshipMutation, error) { return policyv1.Role.TouchMember(policyv1.Role.Admin(), actor) })`,
    `assertRelationship(t, "Role.DeleteMember", "role", "admin", "member", "account_identity", "account-identity-id", "", policyv1.RelationshipDelete, func() (policyv1.RelationshipMutation, error) { return policyv1.Role.DeleteMember(policyv1.Role.Admin(), actor) })`,
  );
  for (const { role, includes } of model.roleInheritance) {
    for (const operation of ["Touch", "Delete"]) {
      const method = `${operation}${pascal(role)}Includes${pascal(includes)}`;
      calls.push(
        `assertRelationship(t, ${quote(`Role.${method}`)}, "role", ${quote(role)}, "member", "role", ${quote(includes)}, "member", policyv1.Relationship${operation}, func() (policyv1.RelationshipMutation, error) { return policyv1.Role.${method}() })`,
      );
    }
  }
  return calls.map((call) => `\t${call}`).join("\n");
}

function renderPlatformSurface(model) {
  const calls = [
    `assertResourceValue(t, "Platform.Resource", "platform", "global", policyv1.Platform.Resource())`,
  ];
  for (const [action, permission] of Object.entries(model.actions.platform)) {
    calls.push(
      `assertCan(t, ${quote(`Platform.${pascal(action)}`)}, "platform", "global", ${quote(action)}, ${quote(permission)}, func() (policyv1.Can, error) { return policyv1.Platform.${pascal(action)}() })`,
    );
  }
  for (const [name, permission] of Object.entries(model.subjectLookups)) {
    calls.push(
      `assertSubjectLookup(t, ${quote(`Platform.Lookup${pascal(name)}Subjects`)}, "platform", ${quote(model.platform.id)}, ${quote(permission)}, ${quote(model.subjectType)}, policyv1.Platform.Lookup${pascal(name)}Subjects())`,
    );
  }
  for (const role of model.roles) {
    for (const operation of ["Touch", "Delete"]) {
      const method = `${operation}${pascal(role)}Role`;
      calls.push(
        `assertRelationship(t, ${quote(`Platform.${method}`)}, "platform", "global", ${quote(role)}, "role", ${quote(role)}, "member", policyv1.Relationship${operation}, func() (policyv1.RelationshipMutation, error) { return policyv1.Platform.${method}() })`,
      );
    }
  }
  return calls.map((call) => `\t${call}`).join("\n");
}

export function renderGoTypedSurfaceTest(model) {
  const domains = model.domains
    .flatMap((domain) => Object.keys(domain.definitions))
    .map((definitionName) => renderDomainSurface(model, definitionName))
    .join("\n");

  return `// Code generated by scripts/generated/generate-spicedb-catalog.mjs.
// DO NOT EDIT.
//
// This executable contract references every generated typed Go constructor.
// The catalog generator and Go deadcode check jointly reject an untested key.

package contracttest_test

import (
\t"testing"

\tpolicyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
)

func TestGeneratedSpiceDBTypedSurface(t *testing.T) {
\tconst resourceID = "resource-id"

\tactor, err := policyv1.NewAccountIdentityActor("account-identity-id")
\tif err != nil { t.Fatal(err) }
\tif !actor.Valid() || actor.AccountIdentityID() != "account-identity-id" {
\t\tt.Fatalf("unexpected generated actor: %#v", actor)
\t}

\tdirect, err := policyv1.DirectSession("session-id")
\tif err != nil { t.Fatal(err) }
\tassertDelegation(t, "DirectSession", policyv1.DelegationDirectSession, "session-id", "", "", direct)
\toauth, err := policyv1.MCPOAuth("oauth-client-id", "Example Member · Example Client")
\tif err != nil { t.Fatal(err) }
\tassertDelegation(t, "MCPOAuth", policyv1.DelegationMCPOAuth, "", "oauth-client-id", "Example Member · Example Client", oauth)
\tdecisionCan, err := policyv1.Platform.IsUser()
\tif err != nil { t.Fatal(err) }
\tdecision, err := policyv1.NewAuthorizationDecision(actor, oauth, decisionCan)
\tif err != nil { t.Fatal(err) }
\tif !decision.Valid() || decision.Actor().AccountIdentityID() != actor.AccountIdentityID() ||
\t\tdecision.Delegation().DelegationID() != oauth.DelegationID() ||
\t\tdecision.Resource().Type() != "platform" || decision.Resource().ID() != "global" ||
\t\tdecision.Action().Name() != "is_user" || decision.Action().Permission() != "is_user" ||
\t\tdecision.EngineKey() != decisionCan.EngineKey() {
\t\tt.Fatalf("unexpected generated authorization decision: %#v", decision)
\t}

${renderRoleSurface(model)}
${renderPlatformSurface(model)}
${domains}
}

func assertDelegation(t *testing.T, name string, kind policyv1.DelegationKind, sessionID, delegationID, displayName string, delegation policyv1.Delegation) {
\tt.Helper()
\tif !delegation.Valid() || delegation.Kind() != kind || delegation.SessionID() != sessionID ||
\t\tdelegation.DelegationID() != delegationID || delegation.DelegationDisplayName() != displayName {
\t\tt.Fatalf("%s returned an unexpected delegation: %#v", name, delegation)
\t}
}

func assertRole(t *testing.T, name, id string, role policyv1.RoleID) {
\tt.Helper()
\tif !role.Valid() || role.ID() != id {
\t\tt.Fatalf("%s returned an unexpected role: %#v", name, role)
\t}
}

func assertResource(t *testing.T, name, resourceType, resourceID string, construct func() (policyv1.Resource, error)) {
\tt.Helper()
\tresource, err := construct()
\tif err != nil { t.Fatalf("%s: %v", name, err) }
\tassertResourceValue(t, name, resourceType, resourceID, resource)
}

func assertResourceValue(t *testing.T, name, resourceType, resourceID string, resource policyv1.Resource) {
\tt.Helper()
\tif !resource.Valid() || resource.Type() != resourceType || resource.ID() != resourceID {
\t\tt.Fatalf("%s returned an unexpected resource: %#v", name, resource)
\t}
}

func assertResourceLookup(t *testing.T, name, resourceType, permission string, lookup policyv1.ResourceLookup) {
	t.Helper()
	if !lookup.Valid() || lookup.ResourceType() != resourceType || lookup.Permission() != permission {
		t.Fatalf("%s returned an unexpected resource lookup: %#v", name, lookup)
	}
}

func assertSubjectLookup(t *testing.T, name, resourceType, resourceID, permission, subjectType string, lookup policyv1.SubjectLookup) {
	t.Helper()
	if !lookup.Valid() || lookup.ResourceType() != resourceType || lookup.ResourceID() != resourceID ||
		lookup.Permission() != permission || lookup.SubjectType() != subjectType {
		t.Fatalf("%s returned an unexpected subject lookup: %#v", name, lookup)
	}
}

func assertCan(t *testing.T, name, resourceType, resourceID, actionName, permission string, construct func() (policyv1.Can, error)) {
\tt.Helper()
\tcan, err := construct()
\tif err != nil { t.Fatalf("%s: %v", name, err) }
\tif !can.Valid() || !can.Resource().Valid() || !can.Action().Valid() ||
\t\tcan.Resource().Type() != resourceType || can.Resource().ID() != resourceID ||
\t\tcan.Action().Name() != actionName || can.Action().Permission() != permission ||
\t\tcan.EngineKey() != resourceType+"\\x00"+resourceID+"\\x00"+permission {
\t\tt.Fatalf("%s returned an unexpected can descriptor: %#v", name, can)
\t}
}

func assertRelationship(t *testing.T, name, resourceType, resourceID, relation, subjectType, subjectID, subjectRelation string, operation policyv1.RelationshipOperation, construct func() (policyv1.RelationshipMutation, error)) {
\tt.Helper()
\tmutation, err := construct()
\tif err != nil { t.Fatalf("%s: %v", name, err) }
\tif !mutation.Valid() || !mutation.Resource().Valid() ||
\t\tmutation.Resource().Type() != resourceType || mutation.Resource().ID() != resourceID ||
\t\tmutation.Relation() != relation || mutation.SubjectType() != subjectType ||
\t\tmutation.SubjectID() != subjectID || mutation.SubjectRelation() != subjectRelation ||
\t\tmutation.Operation() != operation {
\t\tt.Fatalf("%s returned an unexpected relationship: %#v", name, mutation)
\t}
}
`;
}
