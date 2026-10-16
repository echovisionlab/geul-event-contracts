package contentblock_test

import (
	"encoding/json"
	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	"reflect"
	"slices"
	"testing"
)

func TestSiteVariableContracts(t *testing.T) {
	keys := []string{"site.title", "site.url", "site.companyName", "contact.legalEmail", "contact.email"}
	if !reflect.DeepEqual(contentv1.SiteVariableKeys, keys) {
		t.Fatal(contentv1.SiteVariableKeys)
	}
	normalize := func(profile string, content []any) ([]byte, error) {
		data, _ := json.Marshal(map[string]any{"paragraph": map[string]any{"content": content}})
		return contentv1.NormalizeContentStorageLocale(profile, "paragraph", data, contentv1.ContentValidationMode_CONTENT_VALIDATION_MODE_WRITE)
	}
	for _, key := range append(keys, "", "site.Title", "unknown", "contact.supportEmail", "contact.privacyEmail") {
		valid := contentv1.IsSiteVariableKey(key)
		if valid != slices.Contains(keys, key) {
			t.Fatal(key)
		}
		for _, profile := range []string{"page", "policy", "post", "email", "work", "program_event", "compact"} {
			atom := map[string]any{"siteVariable": map[string]any{"key": key, "styles": map[string]any{"bold": true}}}
			normalized, err := normalize(profile, []any{atom, atom})
			if (err == nil) != valid {
				t.Fatalf("%s %s: %v", profile, key, err)
			}
			if valid {
				var doc map[string]any
				if err = json.Unmarshal(normalized, &doc); err != nil {
					t.Fatal(err)
				}
				if len(doc["paragraph"].(map[string]any)["content"].([]any)) != 2 {
					t.Fatal("atoms merged")
				}
			}
		}
		_, err := normalize("post", []any{map[string]any{"siteVariable": map[string]any{"key": key, "linked": true}}})
		if (err == nil) != (valid && key != "site.title" && key != "site.companyName") {
			t.Fatalf("linked %s: %v", key, err)
		}
	}
	for _, href := range []string{"/", "/my", "/tools/hwp?x=1#a", "https://example.com", "mailto:a@example.com", "//x", "/\\x", "/ /x", "javascript:alert(1)", "relative"} {
		_, err := normalize("post", []any{map[string]any{"link": map[string]any{"href": href}}})
		valid := href != "//x" && href != "/\\x" && href != "/ /x" && href != "javascript:alert(1)" && href != "relative"
		if (err == nil) != valid {
			t.Fatalf("href %s: %v", href, err)
		}
	}
	_, err := normalize("post", []any{map[string]any{"siteVariable": map[string]any{"key": "site.title", "styles": map[string]any{"textColor": "bad"}}}})
	if err == nil {
		t.Fatal("invalid style accepted")
	}
}
