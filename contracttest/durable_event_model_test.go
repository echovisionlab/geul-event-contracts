package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestDurableMediaResultWrappersHaveOneTerminalOutcome(t *testing.T) {
	tests := []struct {
		name     string
		message  protoreflect.MessageDescriptor
		complete protoreflect.FullName
		failed   protoreflect.FullName
	}{
		{
			name:     "waveform",
			message:  (&managev1.WaveformResultEvent{}).ProtoReflect().Descriptor(),
			complete: "api.manage.v1.WaveformCompleteEvent",
			failed:   "api.manage.v1.WaveformFailEvent",
		},
		{
			name:     "mesh optimization",
			message:  (&managev1.MeshOptimizationResultEvent{}).ProtoReflect().Descriptor(),
			complete: "api.manage.v1.MeshOptimizationCompleteEvent",
			failed:   "api.manage.v1.MeshOptimizationFailEvent",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireMessageFieldCount(t, test.message, 2)
			for _, expected := range []struct {
				name        protoreflect.Name
				number      protoreflect.FieldNumber
				messageType protoreflect.FullName
			}{
				{name: "completed", number: 1, messageType: test.complete},
				{name: "failed", number: 2, messageType: test.failed},
			} {
				field := requireMessageField(
					t,
					test.message,
					expected.name,
					expected.number,
					protoreflect.MessageKind,
					expected.messageType,
				)
				if oneof := field.ContainingOneof(); oneof == nil || oneof.Name() != "outcome" {
					t.Errorf("%s.%s must be in the outcome oneof", test.message.FullName(), field.Name())
				}
			}
		})
	}
}
