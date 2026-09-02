import { pascal } from "./catalog-source.mjs";

const quote = (value) => JSON.stringify(value);

export function namespaceName(model, definitionName) {
  return model.apiNames[definitionName] ?? pascal(definitionName);
}

export function relationshipMethodName(
  model,
  definitionName,
  relation,
  subject,
  operation,
) {
  const operationName = operation === "touch" ? "Touch" : "Delete";
  let methodName = `${operationName}${pascal(relation)}`;
  if (
    subject !== "platform" &&
    subject !== "account_identity" &&
    model.definitions[definitionName].relations[relation].length > 1
  ) {
    methodName += pascal(subject);
  }
  return methodName;
}

function actionMethods(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  const actions = [
    ...Object.entries(model.actions[definitionName]).map(
      ([name, permission]) => ({ name, permission }),
    ),
    ...Object.entries(model.actionAliases[definitionName]).map(
      ([name, permission]) => ({ name, permission }),
    ),
  ];
  return actions
    .map(
      ({
        name,
        permission,
      }) => `// ${pascal(name)} describes the ${name} business action on one ${definitionName}.
func (${receiver}) ${pascal(name)}(id string) (Can, error) {
	return newCan(${quote(definitionName)}, id, ${quote(name)}, ${quote(permission)})
}`,
    )
    .join("\n\n");
}

function platformActionMethods(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  return Object.entries(model.platformActions[definitionName])
    .map(
      ([
        name,
        permission,
      ]) => `// ${pascal(name)} describes the ${definitionName}.${name} business action on platform:global.
func (${receiver}) ${pascal(name)}() (Can, error) {
	return newCan("platform", ${quote(model.platform.id)}, ${quote(`${definitionName}.${name}`)}, ${quote(permission)})
}`,
    )
    .join("\n\n");
}

function lookupMethods(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  return model.lookupPermissions[definitionName]
    .map(
      (
        permission,
      ) => `// Lookup${pascal(permission)} describes an object-ID-less ${definitionName}.${permission} LookupResources query.
func (${receiver}) Lookup${pascal(permission)}() ResourceLookup {
	return newResourceLookup(${quote(definitionName)}, ${quote(permission)})
}`,
    )
    .join("\n\n");
}

function relationshipMethod(
  model,
  definitionName,
  relation,
  subject,
  operation,
) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  const relationshipOperation =
    operation === "touch" ? "RelationshipTouch" : "RelationshipDelete";
  const methodName = relationshipMethodName(
    model,
    definitionName,
    relation,
    subject,
    operation,
  );
  let parameters = "id string";
  let subjectSetup;

  if (subject === "platform") {
    subjectSetup = "subject := platformRelationshipSubject()";
  } else if (subject === "account_identity") {
    parameters += ", actor Actor";
    subjectSetup = `subject, err := accountIdentityRelationshipSubject(actor)
	if err != nil {
		return RelationshipMutation{}, err
	}`;
  } else {
    const [subjectType, optionalRelation = ""] = subject.split("#");
    parameters += `, ${relation}ID string`;
    subjectSetup = `subject, err := resourceRelationshipSubject(${quote(subjectType)}, ${relation}ID, ${quote(optionalRelation)})
	if err != nil {
		return RelationshipMutation{}, err
	}`;
  }

  return `// ${methodName} creates a closed ${operation} descriptor for ${definitionName}.${relation}.
func (${receiver}) ${methodName}(${parameters}) (RelationshipMutation, error) {
	resource, err := newResource(${quote(definitionName)}, id)
	if err != nil {
		return RelationshipMutation{}, err
	}
	${subjectSetup}
	return newRelationshipMutation(resource, ${quote(relation)}, subject, ${relationshipOperation})
}`;
}

function relationshipMethods(model, definitionName) {
  return Object.entries(model.definitions[definitionName].relations)
    .flatMap(([relation, subjects]) =>
      subjects.flatMap((subject) =>
        ["touch", "delete"].map((operation) =>
          relationshipMethod(
            model,
            definitionName,
            relation,
            subject,
            operation,
          ),
        ),
      ),
    )
    .join("\n\n");
}

function snapshotMethod(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  const permissions = model.definitions[definitionName].permissions;
  if (permissions.length === 0) return "";
  const incomingParent = (
    model.definitions[definitionName].relations.parent ?? []
  ).includes(definitionName);
  return `// Snapshot describes the exact typed relationship-read boundary for one ${definitionName}.
func (${receiver}) Snapshot(id string) (RelationshipSnapshotPlan, error) {
	return newRelationshipSnapshotPlan(${quote(definitionName)}, id, ${quote(permissions[0])}, ${quote(incomingParent ? "parent" : "")})
}`;
}

function renderDomainDefinition(model, definitionName) {
  const namespace = namespaceName(model, definitionName);
  const receiver = `${namespace[0].toLowerCase()}${namespace.slice(1)}Domain`;
  return `type ${receiver} struct{}

// ${namespace} is the closed ${definitionName} authorization namespace.
var ${namespace} = ${receiver}{}

// Resource constructs one typed ${definitionName} resource descriptor.
func (${receiver}) Resource(id string) (Resource, error) {
	return newResource(${quote(definitionName)}, id)
}

${lookupMethods(model, definitionName)}

${actionMethods(model, definitionName)}

${platformActionMethods(model, definitionName)}

${snapshotMethod(model, definitionName)}

${relationshipMethods(model, definitionName)}`;
}

export function renderGoDomains(model) {
  return Object.fromEntries(
    model.domains.map((domain) => [
      `spicedb_${domain.domain}.go`,
      `// Code generated by scripts/generated/generate-spicedb-catalog.mjs.
// DO NOT EDIT.
//
// Domain: ${domain.domain} — ${domain.description}

package policyv1

${Object.keys(domain.definitions)
  .map((definitionName) => renderDomainDefinition(model, definitionName))
  .join("\n\n")}
`,
    ]),
  );
}

function roleFunctions(model) {
  return model.roles
    .map(
      (role) => `// ${pascal(role)} returns the closed ${role} role descriptor.
func (roleDomain) ${pascal(role)}() RoleID {
	return RoleID{key: ${quote(role)}, valid: true}
}`,
    )
    .join("\n\n");
}

function directRoleRelationshipMethods(model) {
  return ["touch", "delete"]
    .map((operation) => {
      const method = operation === "touch" ? "TouchMember" : "DeleteMember";
      const value =
        operation === "touch" ? "RelationshipTouch" : "RelationshipDelete";
      return `// ${method} describes one direct account-identity role membership mutation.
func (roleDomain) ${method}(role RoleID, actor Actor) (RelationshipMutation, error) {
	resource, err := roleResource(role)
	if err != nil {
		return RelationshipMutation{}, err
	}
	subject, err := accountIdentityRelationshipSubject(actor)
	if err != nil {
		return RelationshipMutation{}, err
	}
	return newRelationshipMutation(resource, "member", subject, ${value})
}`;
    })
    .join("\n\n");
}

function inheritedRoleRelationshipMethods(model) {
  return model.roleInheritance
    .flatMap(({ role, includes }) =>
      ["touch", "delete"].map((operation) => {
        const method = `${operation === "touch" ? "Touch" : "Delete"}${pascal(role)}Includes${pascal(includes)}`;
        const value =
          operation === "touch" ? "RelationshipTouch" : "RelationshipDelete";
        return `// ${method} describes the declared ${role} -> ${includes} role inheritance edge.
func (roleDomain) ${method}() (RelationshipMutation, error) {
	resource, err := roleResource(Role.${pascal(role)}())
	if err != nil {
		return RelationshipMutation{}, err
	}
	subject, err := resourceRelationshipSubject("role", Role.${pascal(includes)}().ID(), "member")
	if err != nil {
		return RelationshipMutation{}, err
	}
	return newRelationshipMutation(resource, "member", subject, ${value})
}`;
      }),
    )
    .join("\n\n");
}

function platformRoleRelationshipMethods(model) {
  return model.roles
    .flatMap((role) =>
      ["touch", "delete"].map((operation) => {
        const method = `${operation === "touch" ? "Touch" : "Delete"}${pascal(role)}Role`;
        const value =
          operation === "touch" ? "RelationshipTouch" : "RelationshipDelete";
        return `// ${method} describes the fixed platform.${role} role-set mutation.
func (platformDomain) ${method}() (RelationshipMutation, error) {
	subject, err := resourceRelationshipSubject("role", Role.${pascal(role)}().ID(), "member")
	if err != nil {
		return RelationshipMutation{}, err
	}
	return newRelationshipMutation(Platform.Resource(), ${quote(role)}, subject, ${value})
}`;
      }),
    )
    .join("\n\n");
}

function platformSubjectLookupMethods(model) {
  return Object.entries(model.subjectLookups)
    .map(
      ([
        name,
        permission,
      ]) => `// Lookup${pascal(name)}Subjects describes one platform:global account_identity subject enumeration.
func (platformDomain) Lookup${pascal(name)}Subjects() SubjectLookup {
	return newSubjectLookup("platform", ${quote(model.platform.id)}, ${quote(permission)}, ${quote(model.subjectType)})
}`,
    )
    .join("\n\n");
}

export function renderGoCore(model, schemaSHA256) {
  const platformActions = Object.entries(model.actions.platform)
    .map(
      ([
        action,
        permission,
      ]) => `// ${pascal(action)} describes the platform ${action} check.
func (platformDomain) ${pascal(action)}() (Can, error) {
	return newCan("platform", ${quote(model.platform.id)}, ${quote(action)}, ${quote(permission)})
}`,
    )
    .join("\n\n");

  return `// Code generated by scripts/generated/generate-spicedb-catalog.mjs.
// DO NOT EDIT.
//
// Shared account-identity, delegation, role, and platform authorization core.

package policyv1

import (
	"fmt"
	"strings"
)

// SpiceDBCatalogFingerprint identifies the exact assembled modular catalog.
const SpiceDBCatalogFingerprint = ${quote(model.fingerprint)}

// SpiceDBSchemaSHA256 identifies the exact generated deployable Zed schema.
const SpiceDBSchemaSHA256 = ${quote(schemaSHA256)}

type resourceKey string
type actionKey string
type permissionKey string
type relationKey string
type roleKey string

func validateIdentifier(kind, value string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must be non-empty and trimmed", kind)
	}
	return nil
}

// Actor is one immutable account_identity authorization actor.
type Actor struct {
	accountIdentityID string
	valid             bool
}

// NewAccountIdentityActor validates and constructs the direct SpiceDB subject.
func NewAccountIdentityActor(accountIdentityID string) (Actor, error) {
	if err := validateIdentifier("account identity id", accountIdentityID); err != nil {
		return Actor{}, err
	}
	return Actor{accountIdentityID: accountIdentityID, valid: true}, nil
}

func (actor Actor) AccountIdentityID() string { return actor.accountIdentityID }
func (actor Actor) SubjectType() string { return "account_identity" }
func (actor Actor) Valid() bool               { return actor.valid }

// DelegationKind records attribution only; it never selects a permission.
type DelegationKind uint8

const (
	DelegationDirectSession DelegationKind = iota + 1
	DelegationMCPOAuth
)

// Delegation is immutable request attribution independent from Can.
type Delegation struct {
	kind                  DelegationKind
	sessionID             string
	delegationID          string
	delegationDisplayName string
	valid                 bool
}

// DirectSession validates direct browser/session attribution.
func DirectSession(sessionID string) (Delegation, error) {
	if err := validateIdentifier("session id", sessionID); err != nil {
		return Delegation{}, err
	}
	return Delegation{kind: DelegationDirectSession, sessionID: sessionID, valid: true}, nil
}

func mcpDelegation(kind DelegationKind, delegationID, delegationDisplayName string) (Delegation, error) {
	if err := validateIdentifier("delegation id", delegationID); err != nil {
		return Delegation{}, err
	}
	if err := validateIdentifier("delegation display name", delegationDisplayName); err != nil {
		return Delegation{}, err
	}
	if kind != DelegationMCPOAuth {
		return Delegation{}, fmt.Errorf("MCP delegation kind is invalid")
	}
	return Delegation{kind: kind, delegationID: delegationID, delegationDisplayName: delegationDisplayName, valid: true}, nil
}

// MCPOAuth validates MCP OAuth client/grant attribution. It carries no permission.
func MCPOAuth(delegationID, delegationDisplayName string) (Delegation, error) {
	return mcpDelegation(DelegationMCPOAuth, delegationID, delegationDisplayName)
}

func (delegation Delegation) Kind() DelegationKind         { return delegation.kind }
func (delegation Delegation) SessionID() string            { return delegation.sessionID }
func (delegation Delegation) DelegationID() string         { return delegation.delegationID }
func (delegation Delegation) DelegationDisplayName() string { return delegation.delegationDisplayName }
func (delegation Delegation) Valid() bool                   { return delegation.valid }

// Resource is an immutable, domain-constructed SpiceDB object descriptor.
type Resource struct {
	key   resourceKey
	id    string
	valid bool
}

func newResource(key resourceKey, id string) (Resource, error) {
	if key == "" {
		return Resource{}, fmt.Errorf("resource key is missing")
	}
	if err := validateIdentifier("resource id", id); err != nil {
		return Resource{}, err
	}
	return Resource{key: key, id: id, valid: true}, nil
}

func (resource Resource) Type() string { return string(resource.key) }
func (resource Resource) ID() string   { return resource.id }
func (resource Resource) Valid() bool  { return resource.valid }

// RelationshipSnapshotPlan is the closed relationship-read descriptor for one
// domain resource. It carries the generated probe permission and whether the
// domain owns an incoming same-type parent edge; callers cannot assemble it.
type RelationshipSnapshotPlan struct {
	resource       Resource
	probePermission permissionKey
	incomingParent relationKey
	valid          bool
}

func newRelationshipSnapshotPlan(resourceType resourceKey, resourceID string, probePermission permissionKey, incomingParent relationKey) (RelationshipSnapshotPlan, error) {
	resource, err := newResource(resourceType, resourceID)
	if err != nil {
		return RelationshipSnapshotPlan{}, err
	}
	if probePermission == "" {
		return RelationshipSnapshotPlan{}, fmt.Errorf("snapshot probe permission is required")
	}
	return RelationshipSnapshotPlan{resource: resource, probePermission: probePermission, incomingParent: incomingParent, valid: true}, nil
}

func (plan RelationshipSnapshotPlan) Resource() Resource { return plan.resource }
func (plan RelationshipSnapshotPlan) ProbePermission() string { return string(plan.probePermission) }
func (plan RelationshipSnapshotPlan) IncomingParentRelation() string { return string(plan.incomingParent) }
func (plan RelationshipSnapshotPlan) Valid() bool { return plan.valid && plan.resource.Valid() && plan.probePermission != "" }

// ResourceLookup is an immutable object-ID-less LookupResources descriptor.
type ResourceLookup struct {
	resourceType resourceKey
	permission   permissionKey
	valid        bool
}

func newResourceLookup(resourceType resourceKey, permission permissionKey) ResourceLookup {
	return ResourceLookup{
		resourceType: resourceType,
		permission:   permission,
		valid:        resourceType != "" && permission != "",
	}
}

func (lookup ResourceLookup) ResourceType() string { return string(lookup.resourceType) }
func (lookup ResourceLookup) Permission() string   { return string(lookup.permission) }
func (lookup ResourceLookup) Valid() bool          { return lookup.valid }

// SubjectLookup is an immutable LookupSubjects descriptor with a fixed
// resource, permission, and subject type.
type SubjectLookup struct {
	resourceType resourceKey
	resourceID   string
	permission   permissionKey
	subjectType  resourceKey
	valid        bool
}

func newSubjectLookup(resourceType resourceKey, resourceID string, permission permissionKey, subjectType resourceKey) SubjectLookup {
	return SubjectLookup{
		resourceType: resourceType,
		resourceID:   resourceID,
		permission:   permission,
		subjectType:  subjectType,
		valid:        resourceType != "" && resourceID != "" && permission != "" && subjectType != "",
	}
}

func (lookup SubjectLookup) ResourceType() string { return string(lookup.resourceType) }
func (lookup SubjectLookup) ResourceID() string   { return lookup.resourceID }
func (lookup SubjectLookup) Permission() string   { return string(lookup.permission) }
func (lookup SubjectLookup) SubjectType() string  { return string(lookup.subjectType) }
func (lookup SubjectLookup) Valid() bool           { return lookup.valid }

// Action is an immutable business action with one exact engine permission.
type Action struct {
	key        actionKey
	permission permissionKey
	valid      bool
}

func (action Action) Name() string       { return string(action.key) }
func (action Action) Permission() string { return string(action.permission) }
func (action Action) Valid() bool        { return action.valid }

// Can binds one valid domain Resource to one valid domain Action.
type Can struct {
	resource Resource
	action   Action
	valid    bool
}

func newCan(resourceType resourceKey, resourceID string, action actionKey, permission permissionKey) (Can, error) {
	resource, err := newResource(resourceType, resourceID)
	if err != nil {
		return Can{}, err
	}
	if action == "" || permission == "" {
		return Can{}, fmt.Errorf("action and permission keys are required")
	}
	return Can{resource: resource, action: Action{key: action, permission: permission, valid: true}, valid: true}, nil
}

func (can Can) Resource() Resource { return can.resource }
func (can Can) Action() Action     { return can.action }
func (can Can) Valid() bool        { return can.valid && can.resource.valid && can.action.valid }
func (can Can) EngineKey() string {
	return can.Resource().Type() + "\\x00" + can.Resource().ID() + "\\x00" + can.Action().Permission()
}

// AuthorizationDecision is an immutable descriptor consumed by a PEP adapter.
// Constructing it performs no authorization I/O.
type AuthorizationDecision struct {
	actor      Actor
	delegation Delegation
	resource   Resource
	action     Action
	valid      bool
}

// NewAuthorizationDecision combines attribution with an already typed Can.
func NewAuthorizationDecision(actor Actor, delegation Delegation, can Can) (AuthorizationDecision, error) {
	if !actor.Valid() || !delegation.Valid() || !can.Valid() {
		return AuthorizationDecision{}, fmt.Errorf("actor, delegation, and can must be valid")
	}
	return AuthorizationDecision{actor: actor, delegation: delegation, resource: can.Resource(), action: can.Action(), valid: true}, nil
}

func (decision AuthorizationDecision) Actor() Actor             { return decision.actor }
func (decision AuthorizationDecision) Delegation() Delegation   { return decision.delegation }
func (decision AuthorizationDecision) Resource() Resource       { return decision.resource }
func (decision AuthorizationDecision) Action() Action           { return decision.action }
func (decision AuthorizationDecision) Valid() bool               { return decision.valid }
func (decision AuthorizationDecision) EngineKey() string {
	return decision.Resource().Type() + "\\x00" + decision.Resource().ID() + "\\x00" + decision.Action().Permission()
}

// RelationshipOperation is the closed mutation operation set.
type RelationshipOperation uint8

const (
	RelationshipTouch RelationshipOperation = iota + 1
	RelationshipDelete
)

type relationshipSubject struct {
	resource Resource
	relation relationKey
	valid    bool
}

func accountIdentityRelationshipSubject(actor Actor) (relationshipSubject, error) {
	if !actor.Valid() {
		return relationshipSubject{}, fmt.Errorf("account identity actor is invalid")
	}
	resource, err := newResource("account_identity", actor.AccountIdentityID())
	if err != nil {
		return relationshipSubject{}, err
	}
	return relationshipSubject{resource: resource, valid: true}, nil
}

func platformRelationshipSubject() relationshipSubject {
	return relationshipSubject{resource: Platform.Resource(), valid: true}
}

func resourceRelationshipSubject(resourceType resourceKey, resourceID string, relation relationKey) (relationshipSubject, error) {
	resource, err := newResource(resourceType, resourceID)
	if err != nil {
		return relationshipSubject{}, err
	}
	return relationshipSubject{resource: resource, relation: relation, valid: true}, nil
}

// RelationshipMutation is immutable and can only be created by a typed domain constructor.
type RelationshipMutation struct {
	resource  Resource
	relation  relationKey
	subject   relationshipSubject
	operation RelationshipOperation
	valid     bool
}

func newRelationshipMutation(resource Resource, relation relationKey, subject relationshipSubject, operation RelationshipOperation) (RelationshipMutation, error) {
	if !resource.Valid() || relation == "" || !subject.valid || (operation != RelationshipTouch && operation != RelationshipDelete) {
		return RelationshipMutation{}, fmt.Errorf("relationship mutation descriptor is invalid")
	}
	return RelationshipMutation{resource: resource, relation: relation, subject: subject, operation: operation, valid: true}, nil
}

func (mutation RelationshipMutation) Resource() Resource              { return mutation.resource }
func (mutation RelationshipMutation) Relation() string                { return string(mutation.relation) }
func (mutation RelationshipMutation) SubjectType() string             { return mutation.subject.resource.Type() }
func (mutation RelationshipMutation) SubjectID() string               { return mutation.subject.resource.ID() }
func (mutation RelationshipMutation) SubjectRelation() string         { return string(mutation.subject.relation) }
func (mutation RelationshipMutation) Operation() RelationshipOperation { return mutation.operation }
func (mutation RelationshipMutation) Valid() bool                      { return mutation.valid }

// RoleID is a closed role descriptor.
type RoleID struct {
	key   roleKey
	valid bool
}

func (role RoleID) ID() string  { return string(role.key) }
func (role RoleID) Valid() bool { return role.valid }

type roleDomain struct{}

type accountIdentityDomain struct{}

// AccountIdentity is the sole direct authorization subject namespace.
var AccountIdentity = accountIdentityDomain{}

// Type returns the generated direct subject object type for provider adapters.
func (accountIdentityDomain) Type() string { return "account_identity" }

// Role is the closed role and role-relationship namespace.
var Role = roleDomain{}

${roleFunctions(model)}

func (roleDomain) Parse(value string) (RoleID, bool) {
	switch value {
${model.roles.map((role) => `	case ${quote(role)}:\n		return Role.${pascal(role)}(), true`).join("\n")}
	default:
		return RoleID{}, false
	}
}

func roleResource(role RoleID) (Resource, error) {
	if !role.Valid() {
		return Resource{}, fmt.Errorf("role is invalid")
	}
	return newResource("role", role.ID())
}

${directRoleRelationshipMethods(model)}

${inheritedRoleRelationshipMethods(model)}

type platformDomain struct{}

// Platform is the singleton platform authorization namespace.
var Platform = platformDomain{}

func (platformDomain) Resource() Resource {
	return Resource{key: "platform", id: ${quote(model.platform.id)}, valid: true}
}

${platformActions}

${platformSubjectLookupMethods(model)}

${platformRoleRelationshipMethods(model)}
`;
}
