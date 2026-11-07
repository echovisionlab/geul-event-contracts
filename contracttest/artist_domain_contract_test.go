package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestArtistContractOwnsGalleryParticipantsAndDeletePreview(t *testing.T) {
	service := managev1.File_api_manage_v1_artist_proto.Services().ByName("ArtistService")
	for _, method := range []protoreflect.Name{
		"GetArtistEditorData",
		"SetArtistImages",
		"ListArtistParticipants",
		"SetArtistParticipant",
		"RemoveArtistParticipant",
		"PreviewDeleteArtist",
		"DeleteArtist",
	} {
		if service.Methods().ByName(method) == nil {
			t.Errorf("ArtistService.%s is missing", method)
		}
	}
	requireMethodTier(t, service, "DeleteArtist", policyv1.AuthorizationRole_USER)

	image := (&managev1.ArtistImage{}).ProtoReflect().Descriptor()
	requireMessageField(t, image, "file_id", 1, protoreflect.StringKind, "")
	requireMessageField(t, image, "asset", 2, protoreflect.MessageKind, "api.common.v1.AssetRef")
	requireMessageField(t, image, "sort_order", 3, protoreflect.Int32Kind, "")
	requireMessageField(t, image, "primary", 4, protoreflect.BoolKind, "")

	editor := (&managev1.GetArtistEditorDataResponse{}).ProtoReflect().Descriptor()
	requireMessageField(t, editor, "image_revision", 4, protoreflect.StringKind, "")
	setImages := (&managev1.SetArtistImagesRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, setImages, "file_ids", 2, protoreflect.StringKind, "")
	requireMessageField(t, setImages, "expected_revision", 3, protoreflect.StringKind, "")
	deleteRequest := (&managev1.DeleteArtistRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, deleteRequest, "expected_revision", 2, protoreflect.StringKind, "")
	setParent := (&managev1.SetArtistParentRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, setParent, "parent_artist_id", 2, protoreflect.StringKind, "")
	setLabels := (&managev1.SetArtistLabelsRequest{}).ProtoReflect().Descriptor()
	requireMessageField(t, setLabels, "label_ids", 2, protoreflect.StringKind, "")

	publicArtist := (&openv1.Artist{}).ProtoReflect().Descriptor()
	requireMessageField(t, publicArtist, "images", 20, protoreflect.MessageKind, "api.open.v1.ArtistImage")
}
