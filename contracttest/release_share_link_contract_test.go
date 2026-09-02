package contracttest_test

import (
	"testing"

	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestReleaseShareLinkAcceptsOptionalPassword(t *testing.T) {
	request := (&openv1.GetReleaseRequest{}).ProtoReflect().Descriptor()
	field := requireMessageField(t, request, "share_password", 3, protoreflect.StringKind, "")
	if !field.HasOptionalKeyword() {
		t.Error("GetReleaseRequest.share_password must be optional")
	}
}
