package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/proto"
)

func TestContentUpdatedOriginWireRoundTrip(t *testing.T) {
	origin := "11111111-1111-4111-8111-111111111111"
	for _, value := range []*string{nil, &origin} {
		source := &managev1.ContentUpdatedEvent{EntityId: "post-1", Origin: value}
		wire, err := proto.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		decoded := &managev1.ContentUpdatedEvent{}
		if err := proto.Unmarshal(wire, decoded); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(source, decoded) {
			t.Fatalf("origin lost: %v", decoded)
		}
	}
	field := (&managev1.ContentUpdatedEvent{}).ProtoReflect().Descriptor().Fields().ByName("origin")
	if field == nil || field.Number() != 12 || !field.HasPresence() {
		t.Fatal("origin must be optional field 12")
	}
}
