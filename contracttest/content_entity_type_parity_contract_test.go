package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
)

func TestContentEntityTypeCoversEveryRegisteredAIDocumentDomain(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		domain  managev1.AIDocumentDomain
		content managev1.ContentEntityType
	}{
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_POST, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_POST},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_PAGE, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PAGE},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_WORK, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_WORK},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_PROGRAM_EVENT, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PROGRAM_EVENT},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_RELEASE, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_RELEASE},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_ARTIST, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_ARTIST},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_LABEL, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_LABEL},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_MENU, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_MENU},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_EMAIL_TEMPLATE, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_EMAIL_TEMPLATE},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_EMAIL_LAYOUT, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_EMAIL_LAYOUT},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_CAMPAIGN, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_CAMPAIGN},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_FORM, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_FORM},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_PRIVACY, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PRIVACY},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_TERMS, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_TERMS},
		{managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_POST_SERIES, managev1.ContentEntityType_CONTENT_ENTITY_TYPE_POST_SERIES},
	}

	if got, want := len(managev1.AIDocumentDomain_name)-1, len(pairs); got != want {
		t.Fatalf("registered AI document domains = %d, parity catalog = %d", got, want)
	}
	if got, want := len(managev1.ContentEntityType_name)-1, len(pairs); got != want {
		t.Fatalf("registered content entity types = %d, parity catalog = %d", got, want)
	}

	domains := make(map[managev1.AIDocumentDomain]struct{}, len(pairs))
	contentTypes := make(map[managev1.ContentEntityType]struct{}, len(pairs))
	for _, pair := range pairs {
		if pair.domain == managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_UNSPECIFIED ||
			pair.content == managev1.ContentEntityType_CONTENT_ENTITY_TYPE_UNSPECIFIED {
			t.Fatal("interactive document parity must not contain unspecified values")
		}
		if _, duplicate := domains[pair.domain]; duplicate {
			t.Fatalf("duplicate AI document domain %s", pair.domain)
		}
		if _, duplicate := contentTypes[pair.content]; duplicate {
			t.Fatalf("duplicate content entity type %s", pair.content)
		}
		domains[pair.domain] = struct{}{}
		contentTypes[pair.content] = struct{}{}
	}

	for value := range managev1.AIDocumentDomain_name {
		domain := managev1.AIDocumentDomain(value)
		if domain == managev1.AIDocumentDomain_AI_DOCUMENT_DOMAIN_UNSPECIFIED {
			continue
		}
		if _, ok := domains[domain]; !ok {
			t.Errorf("AI document domain %s has no content entity type", domain)
		}
	}
	for value := range managev1.ContentEntityType_name {
		contentType := managev1.ContentEntityType(value)
		if contentType == managev1.ContentEntityType_CONTENT_ENTITY_TYPE_UNSPECIFIED {
			continue
		}
		if _, ok := contentTypes[contentType]; !ok {
			t.Errorf("content entity type %s has no AI document domain", contentType)
		}
	}
}

func TestContentEntityTypeAdditiveNumericHardCut(t *testing.T) {
	t.Parallel()

	expected := map[managev1.ContentEntityType]int32{
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_POST:           1,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PAGE:           2,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_WORK:           3,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_ARTIST:         4,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_LABEL:          5,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_RELEASE:        6,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_FORM:           7,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PROGRAM_EVENT:  8,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_POST_SERIES:    9,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_MENU:           10,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_EMAIL_TEMPLATE: 11,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_EMAIL_LAYOUT:   12,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_CAMPAIGN:       13,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_PRIVACY:        14,
		managev1.ContentEntityType_CONTENT_ENTITY_TYPE_TERMS:          15,
	}
	for value, number := range expected {
		if int32(value) != number {
			t.Errorf("%s number = %d, want %d", value, value, number)
		}
	}
}

func TestContentEntityTypeSourceOwnsCompleteInteractiveDomainSet(t *testing.T) {
	t.Parallel()

	source := readSourceProto(t, "api/manage/v1/events.proto")
	declaration := sourceDeclaration(t, source, "enum", "ContentEntityType")
	requireSourceContains(t, declaration,
		"CONTENT_ENTITY_TYPE_POST = 1;",
		"CONTENT_ENTITY_TYPE_PAGE = 2;",
		"CONTENT_ENTITY_TYPE_WORK = 3;",
		"CONTENT_ENTITY_TYPE_ARTIST = 4;",
		"CONTENT_ENTITY_TYPE_LABEL = 5;",
		"CONTENT_ENTITY_TYPE_RELEASE = 6;",
		"CONTENT_ENTITY_TYPE_FORM = 7;",
		"CONTENT_ENTITY_TYPE_PROGRAM_EVENT = 8;",
		"CONTENT_ENTITY_TYPE_POST_SERIES = 9;",
		"CONTENT_ENTITY_TYPE_MENU = 10;",
		"CONTENT_ENTITY_TYPE_EMAIL_TEMPLATE = 11;",
		"CONTENT_ENTITY_TYPE_EMAIL_LAYOUT = 12;",
		"CONTENT_ENTITY_TYPE_CAMPAIGN = 13;",
		"CONTENT_ENTITY_TYPE_PRIVACY = 14;",
		"CONTENT_ENTITY_TYPE_TERMS = 15;",
	)
}
