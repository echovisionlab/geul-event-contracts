package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestUserDeleteIdentityCommandHasExplicitNonDestructiveModeBoundary(t *testing.T) {
	command := (&managev1.UserDeleteIdentityCommand{}).ProtoReflect().Descriptor()
	mode := requireMessageField(t, command, "mode", 1, protoreflect.EnumKind, "api.manage.v1.UserDeleteIdentityMode")
	if mode.IsList() || mode.HasPresence() {
		t.Error("UserDeleteIdentityCommand.mode must be a singular proto3 enum with zero-value presence")
	}

	enum := mode.Enum()
	values := []struct {
		name  protoreflect.Name
		num   protoreflect.EnumNumber
		valid bool
	}{
		{name: "UNSPECIFIED", num: 0, valid: false},
		{name: "TOMBSTONE", num: 1, valid: true},
		{name: "UNONBOARDED_HARD_DELETE", num: 2, valid: true},
	}
	if enum.Values().Len() != len(values) {
		t.Fatalf("UserDeleteIdentityMode has %d values, want %d", enum.Values().Len(), len(values))
	}
	for _, expected := range values {
		value := enum.Values().ByName(expected.name)
		if value == nil || value.Number() != expected.num {
			t.Fatalf("UserDeleteIdentityMode.%s must have number %d", expected.name, expected.num)
		}
		if expected.valid && value.Number() == 0 {
			t.Errorf("valid deletion mode %s must not use zero", expected.name)
		}
	}

	if got := (&managev1.UserDeleteIdentityCommand{}).GetMode(); got != managev1.UserDeleteIdentityMode_UNSPECIFIED {
		t.Fatalf("zero-value command mode = %v, want UNSPECIFIED", got)
	}
	// The wire contract deliberately leaves rejection to the producer/consumer;
	// this test prevents a future enum edit from making zero destructive.
	if managev1.UserDeleteIdentityMode_UNSPECIFIED == managev1.UserDeleteIdentityMode_TOMBSTONE ||
		managev1.UserDeleteIdentityMode_UNSPECIFIED == managev1.UserDeleteIdentityMode_UNONBOARDED_HARD_DELETE {
		t.Fatal("UNSPECIFIED must never alias a deletion mode")
	}
}
