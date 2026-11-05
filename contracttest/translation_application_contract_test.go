package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTranslationEntryIsExistenceAndContentOnly(t *testing.T) {
	entry := (&managev1.TranslationEntry{}).ProtoReflect().Descriptor()
	want := []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
	}{
		{"target", 1},
		{"locale", 2},
		{"title", 3},
		{"summary", 4},
		{"content_html", 5},
		{"content_text", 6},
		{"content_json", 7},
		{"updated_at", 8},
		{"og_asset", 9},
	}
	if entry.Fields().Len() != len(want) {
		t.Fatalf("TranslationEntry fields = %d, want %d", entry.Fields().Len(), len(want))
	}
	for index, expected := range want {
		field := entry.Fields().Get(index)
		if field.Name() != expected.name || field.Number() != expected.number {
			t.Fatalf("TranslationEntry[%d] = %s/%d, want %s/%d", index, field.Name(), field.Number(), expected.name, expected.number)
		}
	}
}
