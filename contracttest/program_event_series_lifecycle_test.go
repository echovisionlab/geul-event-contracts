package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestProgramEventSeriesLifecycleStatus(t *testing.T) {
	statuses := []protoreflect.EnumDescriptor{
		managev1.File_api_manage_v1_program_event_proto.Enums().ByName("ProgramEventSeriesStatus"),
		openv1.File_api_open_v1_program_event_proto.Enums().ByName("ProgramEventSeriesStatus"),
	}
	for _, status := range statuses {
		want := []protoreflect.Name{
			"PROGRAM_EVENT_SERIES_STATUS_UNSPECIFIED",
			"PROGRAM_EVENT_SERIES_STATUS_DRAFT",
			"PROGRAM_EVENT_SERIES_STATUS_PUBLISHED",
		}
		if status.Values().Len() != len(want) {
			t.Fatalf("%s has %d values, want only unspecified, draft, published", status.FullName(), status.Values().Len())
		}
		for number, name := range want {
			value := status.Values().ByName(name)
			if value == nil || value.Number() != protoreflect.EnumNumber(number) {
				t.Errorf("%s.%s must be value %d", status.FullName(), name, number)
			}
		}
	}

}

func TestProgramEventSeriesOwnsOneGlobalCopyAndPoster(t *testing.T) {
	manageSeries := managev1.File_api_manage_v1_program_event_proto.Messages().ByName("ProgramEventSeries")
	for _, name := range []protoreflect.Name{"title", "summary", "description", "poster_file_id"} {
		if field := manageSeries.Fields().ByName(name); field == nil {
			t.Errorf("manage ProgramEventSeries must expose global field %s", name)
		}
	}

	openSeries := openv1.File_api_open_v1_program_event_proto.Messages().ByName("ProgramEventSeries")
	if field := openSeries.Fields().ByName("poster_asset"); field == nil {
		t.Error("open ProgramEventSeries must expose its global poster_asset")
	}
}
