package contracttest_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
)

func TestSpiceDBDecisionDelegationCannotChangeEngineCheck(t *testing.T) {
	const originalFingerprint = "b1c92baf58538b463ae2a94d9b26c3c7e920b22bf92ca71841196b017592a2fe"
	if policyv1.SpiceDBCatalogFingerprint != originalFingerprint {
		t.Fatalf("assembled catalog fingerprint changed: %s", policyv1.SpiceDBCatalogFingerprint)
	}
	actor, err := policyv1.NewAccountIdentityActor("identity-id")
	if err != nil {
		t.Fatal(err)
	}
	can, err := policyv1.Post.Edit("post-id")
	if err != nil {
		t.Fatal(err)
	}
	direct, err := policyv1.DirectSession("session-id")
	if err != nil {
		t.Fatal(err)
	}
	oauth, err := policyv1.MCPOAuth("oauth-grant-id", "Example Member · Example Client")
	if err != nil {
		t.Fatal(err)
	}

	directDecision, err := policyv1.NewAuthorizationDecision(actor, direct, can)
	if err != nil {
		t.Fatal(err)
	}
	oauthDecision, err := policyv1.NewAuthorizationDecision(actor, oauth, can)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(directDecision.EngineKey()), []byte(oauthDecision.EngineKey())) {
		t.Fatalf("delegation changed engine key: %q != %q", directDecision.EngineKey(), oauthDecision.EngineKey())
	}
	if got := oauthDecision.Delegation().Kind(); got != policyv1.DelegationMCPOAuth {
		t.Fatalf("OAuth delegation kind = %d", got)
	}
	if got := oauthDecision.Delegation().DelegationID(); got != "oauth-grant-id" {
		t.Fatalf("OAuth delegation ID = %q", got)
	}
	if got := oauthDecision.Delegation().DelegationDisplayName(); got != "Example Member · Example Client" {
		t.Fatalf("OAuth delegation display name = %q", got)
	}
	if got := oauthDecision.Action().Permission(); got != "edit" {
		t.Fatalf("OAuth permission = %q", got)
	}
}

func TestSpiceDBTypeScriptExportsOnlyPublicModules(t *testing.T) {
	root := filepath.Clean("..")
	packageSource, err := os.ReadFile(filepath.Join(root, "packages/proto/package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var packageDocument struct {
		Exports map[string]string `json:"exports"`
	}
	if err := json.Unmarshal(packageSource, &packageDocument); err != nil {
		t.Fatal(err)
	}
	if _, ok := packageDocument.Exports["./policy/spicedb/internal.ts"]; ok {
		t.Fatal("generic TypeScript policy constructors must not be exported")
	}
	entries, err := os.ReadDir(filepath.Join(root, "packages/proto/gen/api/policy/v1/spicedb"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") || entry.Name() == "internal.ts" {
			continue
		}
		key := "./policy/spicedb/" + entry.Name()
		want := "./gen/api/policy/v1/spicedb/" + entry.Name()
		if got := packageDocument.Exports[key]; got != want {
			t.Errorf("package export %s = %q, want %q", key, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "config/spicedb/catalog.yaml")); !os.IsNotExist(err) {
		t.Fatal("parallel monolithic catalog source must remain removed")
	}
}

func TestSpiceDBDomainActionsHideSchemaHelpersAndKeepAliases(t *testing.T) {
	postRemove, err := policyv1.Post.RemoveAuthor("post-id")
	if err != nil {
		t.Fatal(err)
	}
	artistRemove, err := policyv1.Artist.RemoveOwner("artist-id")
	if err != nil {
		t.Fatal(err)
	}
	labelRemove, err := policyv1.Label.RemoveOwner("label-id")
	if err != nil {
		t.Fatal(err)
	}
	for name, can := range map[string]policyv1.Can{
		"post.remove_author":  postRemove,
		"artist.remove_owner": artistRemove,
		"label.remove_owner":  labelRemove,
	} {
		if got := can.Action().Permission(); got != "platform_admin" {
			t.Errorf("%s permission = %q", name, got)
		}
	}

	for namespace, forbidden := range map[any][]string{
		policyv1.Post:   {"PlatformAdmin", "PlatformAuthor", "AuthorManage", "CollaboratorEdit"},
		policyv1.Artist: {"PlatformAdmin", "OwnerManage", "ManagerManage"},
		policyv1.Label:  {"PlatformAdmin", "OwnerManage", "ManagerManage"},
	} {
		typeOf := reflect.TypeOf(namespace)
		for _, method := range forbidden {
			if _, ok := typeOf.MethodByName(method); ok {
				t.Errorf("%s unexpectedly exposes schema helper %s", typeOf, method)
			}
		}
	}

	create, err := policyv1.Post.Create()
	if err != nil {
		t.Fatal(err)
	}
	if create.Resource().Type() != "platform" || create.Resource().ID() != "global" ||
		create.Action().Name() != "post.create" || create.Action().Permission() != "is_author" {
		t.Fatalf("unexpected Post.Create descriptor: %#v %#v", create.Resource(), create.Action())
	}
}

func TestSpiceDBGlobalDomainActionsKeepExactAdminBusinessKeys(t *testing.T) {
	tests := map[string]func() (policyv1.Can, error){
		"file.manage_library":          policyv1.File.ManageLibrary,
		"program_event.update_type":    policyv1.ProgramEvent.UpdateType,
		"program_event.delete_type":    policyv1.ProgramEvent.DeleteType,
		"email_event_mapping.update":   policyv1.EmailEventMapping.Update,
		"email_layout.preview_content": policyv1.EmailLayout.PreviewContent,
		"translation_settings.view":    policyv1.TranslationSettings.View,
		"translation_settings.update":  policyv1.TranslationSettings.Update,
		"translation_provider.list":    policyv1.TranslationProvider.List,
		"translation_provider.create":  policyv1.TranslationProvider.Create,
		"translation_provider.update":  policyv1.TranslationProvider.Update,
		"translation_provider.delete":  policyv1.TranslationProvider.Delete,
		"mail_adapter.list":            policyv1.MailAdapter.List,
		"mail_adapter.create":          policyv1.MailAdapter.Create,
		"mail_adapter.update":          policyv1.MailAdapter.Update,
		"mail_adapter.delete":          policyv1.MailAdapter.Delete,
		"mail_adapter.test":            policyv1.MailAdapter.Test,
		"email_suppression.release":    policyv1.EmailSuppression.Release,
	}
	for action, construct := range tests {
		can, err := construct()
		if err != nil {
			t.Fatalf("%s constructor: %v", action, err)
		}
		if can.Resource().Type() != "platform" || can.Resource().ID() != "global" ||
			can.Action().Name() != action || can.Action().Permission() != "is_admin" {
			t.Errorf("unexpected %s descriptor: %#v %#v", action, can.Resource(), can.Action())
		}
	}
}

func TestSpiceDBRelationshipMutationsAreClosedByDomainAndOperation(t *testing.T) {
	actor, err := policyv1.NewAccountIdentityActor("identity-id")
	if err != nil {
		t.Fatal(err)
	}
	touch, err := policyv1.Post.TouchAuthor("post-id", actor)
	if err != nil {
		t.Fatal(err)
	}
	remove, err := policyv1.Post.DeleteAuthor("post-id", actor)
	if err != nil {
		t.Fatal(err)
	}
	if touch.Resource().Type() != "post" || touch.Relation() != "author" ||
		touch.SubjectType() != "account_identity" || touch.SubjectID() != "identity-id" ||
		touch.Operation() != policyv1.RelationshipTouch {
		t.Fatalf("unexpected Post.TouchAuthor descriptor: %#v", touch)
	}
	if remove.Operation() != policyv1.RelationshipDelete {
		t.Fatalf("Post.DeleteAuthor operation = %v", remove.Operation())
	}
	parent, err := policyv1.Artist.TouchParent("artist-id", "parent-artist-id")
	if err != nil {
		t.Fatal(err)
	}
	if parent.SubjectType() != "artist" || parent.SubjectID() != "parent-artist-id" {
		t.Fatalf("unexpected Artist parent subject: %#v", parent)
	}
	artistSnapshot, err := policyv1.Artist.Snapshot("artist-id")
	if err != nil {
		t.Fatal(err)
	}
	if !artistSnapshot.Valid() || artistSnapshot.Resource().Type() != "artist" ||
		artistSnapshot.Resource().ID() != "artist-id" ||
		artistSnapshot.ProbePermission() != "platform_admin" ||
		artistSnapshot.IncomingParentRelation() != "parent" {
		t.Fatalf("unexpected Artist snapshot plan: %#v", artistSnapshot)
	}

	for namespace, forbidden := range map[any][]string{
		policyv1.Post:   {"TouchOwner", "DeleteOwner", "TouchParent"},
		policyv1.Artist: {"TouchAuthor", "TouchCollaborator"},
		policyv1.Label:  {"TouchAuthor", "TouchCollaborator"},
	} {
		typeOf := reflect.TypeOf(namespace)
		for _, method := range forbidden {
			if _, ok := typeOf.MethodByName(method); ok {
				t.Errorf("%s unexpectedly represents invalid relation operation %s", typeOf, method)
			}
		}
	}

	for _, value := range []any{
		policyv1.Resource{}, policyv1.ResourceLookup{}, policyv1.SubjectLookup{}, policyv1.Action{}, policyv1.Can{}, policyv1.RelationshipMutation{}, policyv1.RelationshipSnapshotPlan{},
	} {
		typeOf := reflect.TypeOf(value)
		for index := range typeOf.NumField() {
			if typeOf.Field(index).IsExported() {
				t.Errorf("%s field %s must remain unforgeable", typeOf, typeOf.Field(index).Name)
			}
		}
	}
	if (policyv1.Resource{}).Valid() || (policyv1.ResourceLookup{}).Valid() || (policyv1.SubjectLookup{}).Valid() || (policyv1.Can{}).Valid() || (policyv1.RelationshipMutation{}).Valid() || (policyv1.RelationshipSnapshotPlan{}).Valid() {
		t.Fatal("zero-value descriptors must not become valid")
	}
}

func TestSpiceDBRelationshipSnapshotPlansAreDomainOwned(t *testing.T) {
	tests := []struct {
		resourceType   string
		incomingParent string
		construct      func(string) (policyv1.RelationshipSnapshotPlan, error)
	}{
		{"artist", "parent", policyv1.Artist.Snapshot},
		{"audience_segment", "", policyv1.AudienceSegment.Snapshot},
		{"campaign", "", policyv1.Campaign.Snapshot},
		{"category", "", policyv1.Category.Snapshot},
		{"client", "", policyv1.Client.Snapshot},
		{"email_layout", "", policyv1.EmailLayout.Snapshot},
		{"email_template", "", policyv1.EmailTemplate.Snapshot},
		{"file", "", policyv1.File.Snapshot},
		{"form", "", policyv1.Form.Snapshot},
		{"format", "", policyv1.Format.Snapshot},
		{"genre", "", policyv1.Genre.Snapshot},
		{"label", "parent", policyv1.Label.Snapshot},
		{"map_place", "", policyv1.MapPlace.Snapshot},
		{"map_theme", "", policyv1.MapTheme.Snapshot},
		{"member", "", policyv1.Member.Snapshot},
		{"menu", "", policyv1.Menu.Snapshot},
		{"page", "", policyv1.Page.Snapshot},
		{"post", "", policyv1.Post.Snapshot},
		{"series", "", policyv1.PostSeries.Snapshot},
		{"privacy_history", "", policyv1.PrivacyHistory.Snapshot},
		{"program_event", "", policyv1.ProgramEvent.Snapshot},
		{"program_event_series", "", policyv1.ProgramEventSeries.Snapshot},
		{"release", "", policyv1.Release.Snapshot},
		{"site_setting", "", policyv1.SiteSetting.Snapshot},
		{"style", "", policyv1.Style.Snapshot},
		{"tag", "", policyv1.Tag.Snapshot},
		{"terms_history", "", policyv1.TermsHistory.Snapshot},
		{"track", "", policyv1.Track.Snapshot},
		{"work", "", policyv1.Work.Snapshot},
	}

	for _, test := range tests {
		t.Run(test.resourceType, func(t *testing.T) {
			plan, err := test.construct("resource-id")
			if err != nil {
				t.Fatal(err)
			}
			if !plan.Valid() || plan.Resource().Type() != test.resourceType ||
				plan.Resource().ID() != "resource-id" || plan.ProbePermission() == "" ||
				plan.IncomingParentRelation() != test.incomingParent {
				t.Fatalf("unexpected snapshot plan: %#v", plan)
			}
		})
	}

	if policyv1.AccountIdentity.Type() != "account_identity" {
		t.Fatalf("unexpected account identity subject type: %q", policyv1.AccountIdentity.Type())
	}
}

func TestSpiceDBPlatformSubjectLookupsAreClosedAndExhaustive(t *testing.T) {
	lookups := map[string]policyv1.SubjectLookup{
		"is_admin":  policyv1.Platform.LookupAdminSubjects(),
		"is_author": policyv1.Platform.LookupAuthorSubjects(),
		"is_user":   policyv1.Platform.LookupUserSubjects(),
	}
	for permission, lookup := range lookups {
		if !lookup.Valid() || lookup.ResourceType() != "platform" || lookup.ResourceID() != "global" ||
			lookup.Permission() != permission || lookup.SubjectType() != "account_identity" {
			t.Errorf("unexpected %s subject lookup: %#v", permission, lookup)
		}
	}
}

func TestSpiceDBResourceLookupsAreObjectIDLessAndDomainOwned(t *testing.T) {
	lookup := policyv1.Post.LookupPlatformAdmin()
	if !lookup.Valid() || lookup.ResourceType() != "post" || lookup.Permission() != "platform_admin" {
		t.Fatalf("unexpected Post.LookupPlatformAdmin descriptor: %#v", lookup)
	}

	postType := reflect.TypeOf(policyv1.Post)
	if _, ok := postType.MethodByName("PlatformAdmin"); ok {
		t.Fatal("lookup permission must not expose a point-check schema helper")
	}
	for namespace, forbidden := range map[any][]string{
		policyv1.Artist:     {"LookupView"},
		policyv1.PostSeries: {"LookupDelete"},
		policyv1.Page:       {"LookupEdit"},
	} {
		typeOf := reflect.TypeOf(namespace)
		for _, method := range forbidden {
			if _, ok := typeOf.MethodByName(method); ok {
				t.Errorf("%s unexpectedly exposes unowned lookup %s", typeOf, method)
			}
		}
	}
}

func TestSpiceDBGoAndTypeScriptDomainSurfacesStayInParity(t *testing.T) {
	root := filepath.Clean("..")
	goRoot := filepath.Join(root, "gen/api/policy/v1")
	tsRoot := filepath.Join(root, "packages/proto/gen/api/policy/v1/spicedb")
	goMethods := regexp.MustCompile(`(?m)^func \([^)]*\) ([A-Z][A-Za-z0-9]*)\(`)
	tsMethods := regexp.MustCompile(`(?m)^  ([A-Z][A-Za-z0-9]*):`)
	goCan := regexp.MustCompile(`newCan\("([^"]+)", (id|"global"), "([^"]+)", "([^"]+)"\)`)
	tsCan := regexp.MustCompile(`(?s)createCan\(\s*"([^"]+)",\s*(id|"global"),\s*"([^"]+)",\s*"([^"]+)",?\s*\)`)
	goLookup := regexp.MustCompile(`newResourceLookup\("([^"]+)", "([^"]+)"\)`)
	tsLookup := regexp.MustCompile(`createResourceLookup\("([^"]+)", "([^"]+)"\)`)

	entries, err := os.ReadDir(tsRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".ts") || name == "core.ts" || name == "internal.ts" {
			continue
		}
		domain := strings.TrimSuffix(name, ".ts")
		goSource, err := os.ReadFile(filepath.Join(goRoot, "spicedb_"+domain+".go"))
		if err != nil {
			t.Fatal(err)
		}
		tsSource, err := os.ReadFile(filepath.Join(tsRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := captureSet(goMethods, goSource), captureSet(tsMethods, tsSource); !slices.Equal(got, want) {
			t.Errorf("%s public method parity\nGo: %v\nTS: %v", domain, got, want)
		}
		if got, want := captureTuples(goCan, goSource), captureTuples(tsCan, tsSource); !slices.Equal(got, want) {
			t.Errorf("%s action parity\nGo: %v\nTS: %v", domain, got, want)
		}
		if got, want := captureTuples(goLookup, goSource), captureTuples(tsLookup, tsSource); !slices.Equal(got, want) {
			t.Errorf("%s lookup parity\nGo: %v\nTS: %v", domain, got, want)
		}
	}
}

func TestSpiceDBResourceLookupFixtureMatchesGeneratedSurfaces(t *testing.T) {
	root := filepath.Clean("..")
	fixtureSource, err := os.ReadFile(filepath.Join("testdata", "spicedb_resource_lookups.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Domain       string   `json:"domain"`
		ResourceType string   `json:"resource_type"`
		Permissions  []string `json:"permissions"`
	}
	if err := json.Unmarshal(fixtureSource, &fixtures); err != nil {
		t.Fatal(err)
	}
	wantByDomain := make(map[string][]string)
	for _, fixture := range fixtures {
		for _, permission := range fixture.Permissions {
			wantByDomain[fixture.Domain] = append(
				wantByDomain[fixture.Domain],
				fixture.ResourceType+"\x00"+permission,
			)
		}
	}

	goLookup := regexp.MustCompile(`newResourceLookup\("([^"]+)", "([^"]+)"\)`)
	tsLookup := regexp.MustCompile(`createResourceLookup\("([^"]+)", "([^"]+)"\)`)
	tsRoot := filepath.Join(root, "packages/proto/gen/api/policy/v1/spicedb")
	entries, err := os.ReadDir(tsRoot)
	if err != nil {
		t.Fatal(err)
	}
	seenDomains := make(map[string]bool)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".ts") || name == "core.ts" || name == "internal.ts" {
			continue
		}
		domain := strings.TrimSuffix(name, ".ts")
		goSource, err := os.ReadFile(filepath.Join(root, "gen/api/policy/v1", "spicedb_"+domain+".go"))
		if err != nil {
			t.Fatal(err)
		}
		tsSource, err := os.ReadFile(filepath.Join(tsRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		gotGo := captureTuples(goLookup, goSource)
		gotTypeScript := captureTuples(tsLookup, tsSource)
		want := append([]string(nil), wantByDomain[domain]...)
		slices.Sort(want)
		if !slices.Equal(gotGo, want) {
			t.Errorf("%s Go lookups\ngot: %v\nwant: %v", domain, gotGo, want)
		}
		if !slices.Equal(gotTypeScript, want) {
			t.Errorf("%s TypeScript lookups\ngot: %v\nwant: %v", domain, gotTypeScript, want)
		}
		if len(want) > 0 {
			seenDomains[domain] = true
		}
	}
	for domain := range wantByDomain {
		if !seenDomains[domain] {
			t.Errorf("lookup fixture domain %s has no generated surface", domain)
		}
	}
}

func TestSpiceDBRawConstructorsStayInsideGeneratedPolicyBoundary(t *testing.T) {
	root := filepath.Clean("..")
	needles := []string{
		"newCan(", "newResource(", "newResourceLookup(", "newRelationshipMutation(",
		"createCan(", "createResource(", "createResourceLookup(", "createRelationshipMutation(",
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "node_modules" || entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "contracttest/spicedb_typed_contract_test.go" ||
			(!strings.HasSuffix(rel, ".go") && !strings.HasSuffix(rel, ".ts") && !strings.HasSuffix(rel, ".mjs")) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contains := false
		for _, needle := range needles {
			contains = contains || bytes.Contains(content, []byte(needle))
		}
		if !contains {
			return nil
		}
		allowed := strings.HasPrefix(rel, "gen/api/policy/v1/spicedb_") ||
			strings.HasPrefix(rel, "packages/proto/gen/api/policy/v1/spicedb/") ||
			strings.HasPrefix(rel, "scripts/generated/spicedb/")
		if !allowed {
			t.Errorf("raw SpiceDB constructor escaped generated policy boundary: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func captureSet(pattern *regexp.Regexp, source []byte) []string {
	values := make([]string, 0)
	seen := make(map[string]bool)
	for _, match := range pattern.FindAllSubmatch(source, -1) {
		value := string(match[1])
		if !seen[value] {
			seen[value] = true
			values = append(values, value)
		}
	}
	slices.Sort(values)
	return values
}

func captureTuples(pattern *regexp.Regexp, source []byte) []string {
	values := make([]string, 0)
	for _, match := range pattern.FindAllSubmatch(source, -1) {
		parts := make([]string, 0, len(match)-1)
		for _, part := range match[1:] {
			parts = append(parts, string(part))
		}
		values = append(values, strings.Join(parts, "\x00"))
	}
	slices.Sort(values)
	return values
}
