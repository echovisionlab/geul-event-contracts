package contentblock_test

import (
	"testing"

	contentv1 "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
)

func TestPostRichTextCatalogDescriptorExposesClosedRecursiveAuthority(t *testing.T) {
	t.Parallel()

	descriptor, err := contentv1.DescribeRichTextCatalog(contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST)
	if err != nil {
		t.Fatal(err)
	}
	if descriptor.Fingerprint != contentv1.ContentBlockCatalogFingerprint || descriptor.Profile != "post" {
		t.Fatalf("unexpected descriptor identity: %+v", descriptor)
	}
	if len(descriptor.Blocks) != 17 || descriptor.Blocks[0].Kind != "paragraph" || descriptor.Blocks[15].Kind != "callout" || descriptor.Blocks[16].Kind != "mermaid" {
		t.Fatalf("unexpected Post Block order: %+v", descriptor.Blocks)
	}
	if !descriptor.InlineMath || !descriptor.ParagraphExternalVideo {
		t.Fatalf("Post profile flags were lost: %+v", descriptor)
	}
	wantMarks := []string{"bold", "italic", "underline", "strike", "code", "textColor", "backgroundColor", "link"}
	if len(descriptor.Marks) != len(wantMarks) {
		t.Fatalf("marks = %v, want %v", descriptor.Marks, wantMarks)
	}
	for index := range wantMarks {
		if descriptor.Marks[index].Name != wantMarks[index] {
			t.Fatalf("marks = %v, want %v", descriptor.Marks, wantMarks)
		}
	}
	for _, markName := range []string{"textColor", "backgroundColor"} {
		mark := requireMarkDescriptor(t, descriptor, markName)
		if mark.Parameter == nil || mark.Parameter.Name != "value" || mark.Parameter.Type != "editor_color" || mark.Parameter.Ownership != "locale" {
			t.Fatalf("%s mark parameter is incomplete: %+v", markName, mark)
		}
	}
	link := requireMarkDescriptor(t, descriptor, "link")
	if link.Parameter == nil || link.Parameter.Name != "href" || link.Parameter.Type != "uri" {
		t.Fatalf("link parameter is incomplete: %+v", link)
	}

	columnWidths := requireFieldDescriptor(t, descriptor.Table.Fields, "columnWidths")
	if columnWidths.Type != "array" || columnWidths.Item == nil || columnWidths.Item.Type != "number" || columnWidths.ItemIdentity == nil || columnWidths.ItemIdentity.Strategy != "position" || columnWidths.ItemIdentity.Mutation != "whole-list-replace" {
		t.Fatalf("table column widths descriptor is incomplete: %+v", columnWidths)
	}
	if descriptor.Table.RowIdentity.Name != "id" || descriptor.Table.RowIdentity.Type != "uuid" || !descriptor.Table.RowIdentity.Required || descriptor.Table.CellIdentity.Name != "id" || descriptor.Table.CellIdentity.Type != "uuid" || !descriptor.Table.CellIdentity.Required {
		t.Fatalf("stable table identities are incomplete: %+v", descriptor.Table)
	}
	if requireFieldDescriptor(t, descriptor.Table.CellFields, "backgroundColor").Type != "editor_color" || len(descriptor.Table.CellContent.Kinds) != 4 {
		t.Fatalf("table cell schema is incomplete: %+v", descriptor.Table)
	}

	shader := requireBlockDescriptor(t, descriptor, "shader")
	stages := requireFieldDescriptor(t, shader.Fields, "stages")
	if stages.Type != "array" || stages.Ownership != "source" || !stages.Required || stages.ExactLength == nil || *stages.ExactLength != 9 || stages.Item == nil || stages.Item.Type != "object" || stages.ItemIdentity == nil || stages.ItemIdentity.Strategy != "field" || stages.ItemIdentity.Field != "kind" {
		t.Fatalf("shader stages descriptor is incomplete: %+v", stages)
	}
	stageKind := requireFieldDescriptor(t, stages.Item.Fields, "kind")
	if stageKind.Type != "enum" || len(stageKind.Values) != 9 {
		t.Fatalf("shader stage kinds are incomplete: %+v", stageKind)
	}
	channels := requireFieldDescriptor(t, stages.Item.Fields, "channels")
	if channels.Type != "array" || channels.ExactLength == nil || *channels.ExactLength != 4 || channels.Item == nil || channels.ItemIdentity == nil || channels.ItemIdentity.Strategy != "fixed" || !equalStrings(channels.ItemIdentity.Values, []string{"channel-a", "channel-b", "channel-c", "channel-d"}) {
		t.Fatalf("shader channels descriptor is incomplete: %+v", channels)
	}
	file := requireFieldDescriptor(t, channels.Item.Fields, "file")
	if file.Type != "file_attachment" || file.Ownership != "source" || file.FileReference != "shader_channel" {
		t.Fatalf("shader channel File policy was lost: %+v", file)
	}
	faces := requireFieldDescriptor(t, channels.Item.Fields, "faces")
	if faces.Item == nil || faces.Item.Type != "file_attachment" || faces.Item.FileReference != "shader_cubemap_face" || faces.ItemIdentity == nil || faces.ItemIdentity.Strategy != "fixed" || !equalStrings(faces.ItemIdentity.Values, []string{"px", "nx", "py", "ny", "pz", "nz"}) {
		t.Fatalf("shader cubemap face File policy was lost: %+v", faces)
	}

	capabilities := requireFieldDescriptor(t, requireBlockDescriptor(t, descriptor, "p5-sketch").Fields, "capabilities")
	if !capabilities.HasDefault || capabilities.Item == nil || capabilities.Item.Type != "enum" || len(capabilities.Item.Values) != 8 || capabilities.ItemIdentity == nil || capabilities.ItemIdentity.Strategy != "value" {
		t.Fatalf("p5 capabilities descriptor is incomplete: %+v", capabilities)
	}
	if values, ok := capabilities.Default.([]any); !ok || len(values) != 0 {
		t.Fatalf("p5 capabilities default presence was lost: %#v", capabilities.Default)
	}
	mapPlaces := requireFieldDescriptor(t, requireBlockDescriptor(t, descriptor, "map").Fields, "mapPlaceIds")
	if mapPlaces.Item == nil || mapPlaces.Item.Type != "uuid" || !mapPlaces.Unique || mapPlaces.ItemIdentity == nil || mapPlaces.ItemIdentity.Strategy != "value" {
		t.Fatalf("mapPlaceIds descriptor is incomplete: %+v", mapPlaces)
	}
	attachment := requireFieldDescriptor(t, requireBlockDescriptor(t, descriptor, "file").Fields, "attachment")
	if !attachment.Required || attachment.FileReference != "file" {
		t.Fatalf("File attachment descriptor is incomplete: %+v", attachment)
	}
	callout := requireBlockDescriptor(t, descriptor, "callout")
	icon := requireFieldDescriptor(t, callout.Fields, "icon")
	background := requireFieldDescriptor(t, callout.Fields, "backgroundColor")
	if callout.Content != "inline" || icon.Type != "string" || !icon.HasDefault || icon.Default != "💡" || background.Type != "editor_color" || !background.HasDefault || background.Default != "gray" {
		t.Fatalf("Callout descriptor is incomplete: %+v", callout)
	}
	if len(descriptor.FilePolicies) != 3 || descriptor.FilePolicies[0].Name != "file" || descriptor.FilePolicies[1].Name != "shader_channel" || descriptor.FilePolicies[2].Name != "shader_cubemap_face" {
		t.Fatalf("Post descriptor leaked unrelated File policies: %+v", descriptor.FilePolicies)
	}
}

func TestRichTextCatalogDescriptorReturnsDeepCallerOwnedCopies(t *testing.T) {
	t.Parallel()

	first, err := contentv1.DescribeRichTextCatalog(contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST)
	if err != nil {
		t.Fatal(err)
	}
	first.Marks[0].Name = "mutated"
	first.Marks[5].Parameter.Type = "mutated"
	first.Table.RowIdentity.Name = "mutated"
	first.Blocks[0].Fields[0].Name = "mutated"
	capabilities := requireFieldDescriptor(t, requireBlockDescriptor(t, first, "p5-sketch").Fields, "capabilities")
	capabilities.Item.Values[0] = "mutated"

	second, err := contentv1.DescribeRichTextCatalog(contentv1.RichTextProfile_RICH_TEXT_PROFILE_POST)
	if err != nil {
		t.Fatal(err)
	}
	if second.Marks[0].Name != "bold" || second.Marks[5].Parameter.Type != "editor_color" || second.Table.RowIdentity.Name != "id" || second.Blocks[0].Fields[0].Name == "mutated" {
		t.Fatalf("descriptor mutation escaped into generated authority: %+v", second)
	}
	secondCapabilities := requireFieldDescriptor(t, requireBlockDescriptor(t, second, "p5-sketch").Fields, "capabilities")
	if secondCapabilities.Item.Values[0] != "camera" {
		t.Fatalf("nested descriptor mutation escaped: %+v", secondCapabilities)
	}
}

func TestPageCatalogDescriptorExposesLocaleLeafAuthorityAsCallerOwnedCopies(t *testing.T) {
	t.Parallel()

	first := contentv1.DescribePageCatalog()
	if first.Fingerprint != contentv1.ContentBlockCatalogFingerprint || len(first.Sections) != 21 {
		t.Fatalf("unexpected Page descriptor identity: %+v", first)
	}
	externalVideo := requirePageSectionDescriptor(t, first, "external-video")
	caption := requireFieldDescriptor(t, externalVideo.Fields, "caption")
	if caption.Type != "string" || caption.Ownership != "locale" || !caption.Translatable || !caption.HasDefault || caption.Default != "" {
		t.Fatalf("Page external-video caption authority is incomplete: %+v", caption)
	}
	immersiveTitle := requireFieldDescriptor(t, first.ImmersiveUnitFields, "title")
	immersiveText := requireFieldDescriptor(t, first.ImmersiveUnitFields, "text")
	if immersiveTitle.Ownership != "locale" || !immersiveTitle.Translatable || immersiveText.Ownership != "locale" || !immersiveText.Translatable {
		t.Fatalf("Page immersive locale authority is incomplete: title=%+v text=%+v", immersiveTitle, immersiveText)
	}

	first.Sections[0].Kind = "mutated"
	caption.Name = "mutated"
	first.ImmersiveUnitFields[0].Name = "mutated"
	second := contentv1.DescribePageCatalog()
	if second.Sections[0].Kind == "mutated" || requireFieldDescriptor(t, requirePageSectionDescriptor(t, second, "external-video").Fields, "caption").Name != "caption" || second.ImmersiveUnitFields[0].Name == "mutated" {
		t.Fatalf("Page descriptor mutation escaped into generated authority: %+v", second)
	}
}

func requireMarkDescriptor(t *testing.T, descriptor contentv1.RichTextCatalogDescriptor, name string) contentv1.ContentInlineMarkDescriptor {
	t.Helper()
	for _, mark := range descriptor.Marks {
		if mark.Name == name {
			return mark
		}
	}
	t.Fatalf("mark descriptor %q is missing", name)
	return contentv1.ContentInlineMarkDescriptor{}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestRichTextCatalogDescriptorRejectsUnknownProfile(t *testing.T) {
	t.Parallel()
	if _, err := contentv1.DescribeRichTextCatalog(contentv1.RichTextProfile_RICH_TEXT_PROFILE_UNSPECIFIED); err == nil {
		t.Fatal("unspecified Rich Text profile was accepted")
	}
}

func requireBlockDescriptor(t *testing.T, descriptor contentv1.RichTextCatalogDescriptor, kind string) contentv1.ContentBlockDescriptor {
	t.Helper()
	for _, block := range descriptor.Blocks {
		if block.Kind == kind {
			return block
		}
	}
	t.Fatalf("Block descriptor %q is missing", kind)
	return contentv1.ContentBlockDescriptor{}
}

func requirePageSectionDescriptor(t *testing.T, descriptor contentv1.PageCatalogDescriptor, kind string) contentv1.PageSectionDescriptor {
	t.Helper()
	for _, section := range descriptor.Sections {
		if section.Kind == kind {
			return section
		}
	}
	t.Fatalf("Page Section descriptor %q is missing", kind)
	return contentv1.PageSectionDescriptor{}
}

func requireFieldDescriptor(t *testing.T, fields []contentv1.ContentFieldDescriptor, name string) contentv1.ContentFieldDescriptor {
	t.Helper()
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	t.Fatalf("field descriptor %q is missing", name)
	return contentv1.ContentFieldDescriptor{}
}
