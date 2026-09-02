import { describe, expect, it } from "vitest";

import { Artist } from "@echovisionlab/geul-proto/policy/spicedb/artist.ts";
import {
  EmailEventMapping,
  EmailLayout,
} from "@echovisionlab/geul-proto/policy/spicedb/email_authoring.ts";
import {
  EmailSuppression,
  MailAdapter,
} from "@echovisionlab/geul-proto/policy/spicedb/email_delivery.ts";
import { File } from "@echovisionlab/geul-proto/policy/spicedb/file.ts";
import {
  DelegationKind,
  DirectSession,
  MCPOAuth,
  NewAccountIdentityActor,
  NewAuthorizationDecision,
  Platform,
  RelationshipOperation,
  spiceDbCatalogFingerprint,
} from "@echovisionlab/geul-proto/policy/spicedb/core.ts";
import type {
  Can,
  RelationshipMutation,
  Resource,
  ResourceLookup,
  SubjectLookup,
} from "@echovisionlab/geul-proto/policy/spicedb/core.ts";
import { Label } from "@echovisionlab/geul-proto/policy/spicedb/label.ts";
import { Post } from "@echovisionlab/geul-proto/policy/spicedb/post.ts";
import { ProgramEvent } from "@echovisionlab/geul-proto/policy/spicedb/program_event.ts";
import {
  TranslationProvider,
  TranslationSettings,
} from "@echovisionlab/geul-proto/policy/spicedb/translation.ts";

function compileTimeInvalidSurfaceMustStayUnrepresentable(): void {
  const actor = NewAccountIdentityActor("identity-id");

  // @ts-expect-error Resource keys cannot be caller-authored.
  const rawResource: Resource = { type: "post", id: "post-id" };
  // @ts-expect-error Action and permission keys cannot be caller-authored.
  const rawCan: Can = {};
  // @ts-expect-error Lookup resource and permission keys cannot be caller-authored.
  const rawLookup: ResourceLookup = {
    resourceType: "post",
    permission: "view",
  };
  // @ts-expect-error Subject lookup resource, permission, and subject keys cannot be caller-authored.
  const rawSubjectLookup: SubjectLookup = {
    resourceType: "platform",
    resourceId: "global",
    permission: "is_admin",
    subjectType: "account_identity",
  };
  // @ts-expect-error Relation and operation combinations require a domain constructor.
  const rawRelationship: RelationshipMutation = {};

  const validCan = Post.Edit("post-id");
  const spreadCan = {
    ...validCan,
    action: { ...validCan.action, permission: "delete" },
  };
  // @ts-expect-error Spreading a valid descriptor does not copy its private nominal key.
  const forgedCan: Can = spreadCan;
  const spreadResource = { ...Post.Resource("post-id"), type: "artist" };
  // @ts-expect-error Spreading a valid resource does not copy its private nominal key.
  const forgedResource: Resource = spreadResource;
  const spreadLookup = {
    ...Post.LookupView(),
    permission: "delete",
  };
  // @ts-expect-error Spreading a lookup does not copy its private nominal key.
  const forgedLookup: ResourceLookup = spreadLookup;

  // @ts-expect-error Post has no Owner relation.
  Post.TouchOwner("post-id", actor);
  // @ts-expect-error Artist has no Collaborator relation.
  Artist.TouchCollaborator("artist-id", actor);
  // @ts-expect-error platform_admin is a schema helper, not a Post action.
  Post.PlatformAdmin("post-id");
  // @ts-expect-error ManageParticipants requires a Post id.
  Post.ManageParticipants();
  // @ts-expect-error Artist owns only the manage resource lookup.
  Artist.LookupView();

  void rawResource;
  void rawCan;
  void rawRelationship;
  void rawLookup;
  void rawSubjectLookup;
  void forgedCan;
  void forgedResource;
  void forgedLookup;
}

void compileTimeInvalidSurfaceMustStayUnrepresentable;

describe("domain-shaped SpiceDB policy descriptors", () => {
  it("keeps direct-session and MCP engine checks byte-identical", () => {
    expect(spiceDbCatalogFingerprint).toBe(
      "b1c92baf58538b463ae2a94d9b26c3c7e920b22bf92ca71841196b017592a2fe",
    );
    const actor = NewAccountIdentityActor("identity-id");
    const can = Post.Edit("post-id");
    const direct = NewAuthorizationDecision(
      actor,
      DirectSession("session-id"),
      can,
    );
    const oauth = NewAuthorizationDecision(
      actor,
      MCPOAuth("oauth-grant-id", "Example Member · Example Client"),
      can,
    );
    expect(new TextEncoder().encode(oauth.engineKey)).toEqual(
      new TextEncoder().encode(direct.engineKey),
    );
    expect(oauth.engineKey).toBe("post\0post-id\0edit");
    expect(oauth.delegation.kind).toBe(DelegationKind.MCPOAuth);
    expect(oauth.delegation.delegationId).toBe("oauth-grant-id");
    expect(oauth.delegation.delegationDisplayName).toBe(
      "Example Member · Example Client",
    );
    expect(Object.keys(oauth.delegation)).toEqual([
      "kind",
      "sessionId",
      "delegationId",
      "delegationDisplayName",
    ]);
    expect(Object.isFrozen(oauth)).toBe(true);
  });

  it("keeps exact global admin actions in their owning domains", () => {
    const actions = {
      "file.manage_library": File.ManageLibrary(),
      "program_event.update_type": ProgramEvent.UpdateType(),
      "program_event.delete_type": ProgramEvent.DeleteType(),
      "email_event_mapping.update": EmailEventMapping.Update(),
      "email_layout.preview_content": EmailLayout.PreviewContent(),
      "translation_settings.view": TranslationSettings.View(),
      "translation_settings.update": TranslationSettings.Update(),
      "translation_provider.list": TranslationProvider.List(),
      "translation_provider.create": TranslationProvider.Create(),
      "translation_provider.update": TranslationProvider.Update(),
      "translation_provider.delete": TranslationProvider.Delete(),
      "mail_adapter.list": MailAdapter.List(),
      "mail_adapter.create": MailAdapter.Create(),
      "mail_adapter.update": MailAdapter.Update(),
      "mail_adapter.delete": MailAdapter.Delete(),
      "mail_adapter.test": MailAdapter.Test(),
      "email_suppression.release": EmailSuppression.Release(),
    };
    for (const [name, can] of Object.entries(actions)) {
      expect(can).toMatchObject({
        resource: { type: "platform", id: "global" },
        action: { name, permission: "is_admin" },
      });
    }
  });

  it("exposes business aliases without schema helper actions", () => {
    expect(Post.RemoveAuthor("post-id").action).toMatchObject({
      name: "remove_author",
      permission: "platform_admin",
    });
    expect(Artist.RemoveOwner("artist-id").action.permission).toBe(
      "platform_admin",
    );
    expect(Label.RemoveOwner("label-id").action.permission).toBe(
      "platform_admin",
    );
    expect(Post.Create()).toMatchObject({
      resource: { type: "platform", id: "global" },
      action: { name: "post.create", permission: "is_author" },
    });
    expect(Post).not.toHaveProperty("PlatformAdmin");
    expect(Post).not.toHaveProperty("AuthorManage");
    expect(Artist).not.toHaveProperty("OwnerManage");
    expect(Label).not.toHaveProperty("ManagerManage");
  });

  it("constructs exact object-ID-less resource lookups", () => {
    const lookup = Post.LookupPlatformAdmin();
    expect(lookup).toEqual({
      resourceType: "post",
      permission: "platform_admin",
    });
    expect(Object.isFrozen(lookup)).toBe(true);
    expect(Post).not.toHaveProperty("PlatformAdmin");
    expect(Artist).not.toHaveProperty("LookupView");
  });

  it("constructs only the closed platform account-identity subject lookups", () => {
    expect(Platform.LookupAdminSubjects()).toEqual({
      resourceType: "platform",
      resourceId: "global",
      permission: "is_admin",
      subjectType: "account_identity",
    });
    expect(Platform.LookupAuthorSubjects().permission).toBe("is_author");
    expect(Platform.LookupUserSubjects().permission).toBe("is_user");
    expect(Object.isFrozen(Platform.LookupAdminSubjects())).toBe(true);
  });

  it("binds valid relation subjects and operations in domain constructors", () => {
    const actor = NewAccountIdentityActor("identity-id");
    expect(Post.TouchAuthor("post-id", actor)).toMatchObject({
      resource: { type: "post", id: "post-id" },
      relation: "author",
      subject: { type: "account_identity", id: "identity-id", relation: "" },
      operation: RelationshipOperation.Touch,
    });
    expect(Post.DeleteCollaborator("post-id", actor).operation).toBe(
      RelationshipOperation.Delete,
    );
    expect(Artist.TouchParent("artist-id", "parent-id").subject).toEqual({
      type: "artist",
      id: "parent-id",
      relation: "",
    });
    expect(Post).not.toHaveProperty("TouchOwner");
    expect(Artist).not.toHaveProperty("TouchCollaborator");
  });
});
