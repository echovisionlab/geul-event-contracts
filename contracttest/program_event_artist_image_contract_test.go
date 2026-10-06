package contracttest_test

import (
	commonv1 "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"testing"
)

func TestProgramEventArtistImageAssetPreservesExistingFieldsAndWirePresence(t *testing.T) {
	t.Parallel()
	descriptor := (&openv1.ProgramEventArtist{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, descriptor, 6)
	requireMessageField(t, descriptor, "id", 1, protoreflect.StringKind, "")
	requireMessageField(t, descriptor, "name", 2, protoreflect.StringKind, "")
	requireMessageField(t, descriptor, "slug", 3, protoreflect.StringKind, "")
	requireMessageField(t, descriptor, "role", 4, protoreflect.StringKind, "")
	requireMessageField(t, descriptor, "sort_order", 5, protoreflect.Int32Kind, "")
	image := requireMessageField(t, descriptor, "image_asset", 6, protoreflect.MessageKind, "api.common.v1.AssetRef")
	if !image.HasPresence() {
		t.Fatal("artist image asset must preserve absence")
	}
	for _, asset := range []*commonv1.AssetRef{nil, {AssetId: "artist-image", Url: "https://cdn.example.com/artist.webp", MimeType: "image/webp", Extension: "webp"}} {
		original := &openv1.ProgramEventArtist{Id: "artist-1", Name: "Artist", Slug: proto.String("artist"), Role: proto.String("Performer"), SortOrder: 2, ImageAsset: asset}
		wire, err := proto.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		decoded := &openv1.ProgramEventArtist{}
		if err := proto.Unmarshal(wire, decoded); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(original, decoded) {
			t.Fatal("public event artist image wire roundtrip differs")
		}
	}
}
