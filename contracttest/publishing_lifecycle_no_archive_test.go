package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestArtistLabelAndReleaseStatusesHaveNoArchiveState(t *testing.T) {
	tests := []struct {
		name         string
		manageStatus protoreflect.EnumDescriptor
		openStatus   protoreflect.EnumDescriptor
	}{
		{
			name:         "artist",
			manageStatus: managev1.File_api_manage_v1_artist_proto.Enums().ByName("ArtistStatus"),
			openStatus:   openv1.File_api_open_v1_artist_proto.Enums().ByName("ArtistStatus"),
		},
		{
			name:         "label",
			manageStatus: managev1.File_api_manage_v1_label_proto.Enums().ByName("LabelStatus"),
			openStatus:   openv1.File_api_open_v1_label_proto.Enums().ByName("LabelStatus"),
		},
		{
			name:         "release",
			manageStatus: managev1.File_api_manage_v1_release_proto.Enums().ByName("ReleaseStatus"),
			openStatus:   openv1.File_api_open_v1_release_proto.Enums().ByName("ReleaseStatus"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireDraftPublishedStatus(t, test.manageStatus)
			requireDraftPublishedStatus(t, test.openStatus)

		})
	}
}

func TestReleaseAndTrackManageServicesAreAdminOnly(t *testing.T) {
	services := []protoreflect.ServiceDescriptor{
		managev1.File_api_manage_v1_release_proto.Services().ByName("ReleaseService"),
		managev1.File_api_manage_v1_track_proto.Services().ByName("TrackService"),
	}

	for _, service := range services {
		for index := 0; index < service.Methods().Len(); index++ {
			requireMethodTier(
				t,
				service,
				service.Methods().Get(index).Name(),
				policyv1.AuthorizationRole_ADMIN,
			)
		}
	}
}

func requireDraftPublishedStatus(t *testing.T, status protoreflect.EnumDescriptor) {
	t.Helper()

	// Enum names include the domain prefix rather than the descriptor name.
	unspecified := string(status.Values().Get(0).Name())
	prefix := unspecified[:len(unspecified)-len("UNSPECIFIED")]
	want := map[protoreflect.Name]protoreflect.EnumNumber{
		protoreflect.Name(prefix + "UNSPECIFIED"): 0,
		protoreflect.Name(prefix + "DRAFT"):       1,
		protoreflect.Name(prefix + "PUBLISHED"):   2,
	}

	if status.Values().Len() != len(want) {
		t.Fatalf("%s has %d values, want only unspecified, draft, published", status.FullName(), status.Values().Len())
	}
	for name, number := range want {
		value := status.Values().ByName(name)
		if value == nil || value.Number() != number {
			t.Errorf("%s.%s must remain value %d", status.FullName(), name, number)
		}
	}
}
