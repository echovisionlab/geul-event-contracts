package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestCreateCampaignUsesAnExplicitExclusiveTarget(t *testing.T) {
	request := (&managev1.CreateCampaignRequest{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, request, 5)
	requireMessageField(t, request, "subject", 1, protoreflect.StringKind, "")
	requireMessageField(t, request, "name", 2, protoreflect.StringKind, "")
	requireMessageField(t, request, "source_locale", 5, protoreflect.StringKind, "")

	all := requireMessageField(
		t,
		request,
		"all",
		3,
		protoreflect.MessageKind,
		"google.protobuf.Empty",
	)
	criteriaID := requireMessageField(
		t,
		request,
		"member_criteria_id",
		4,
		protoreflect.StringKind,
		"",
	)
	if all.ContainingOneof() == nil || all.ContainingOneof() != criteriaID.ContainingOneof() {
		t.Fatal("CreateCampaignRequest.all and member_criteria_id must be the same oneof")
	}
	if got := all.ContainingOneof().Fields().Len(); got != 2 {
		t.Fatalf("CreateCampaignRequest.target case count = %d, want 2", got)
	}
	if request.Fields().ByName("target_mode") != nil {
		t.Fatal("CreateCampaignRequest must not reintroduce a target_mode fallback")
	}

	allRequest := &managev1.CreateCampaignRequest{
		Target: &managev1.CreateCampaignRequest_All{All: &emptypb.Empty{}},
	}
	if got := allRequest.ProtoReflect().WhichOneof(all.ContainingOneof()); got != all {
		t.Fatalf("explicit all target = %v, want all", got)
	}

	criteriaRequest := &managev1.CreateCampaignRequest{
		Target: &managev1.CreateCampaignRequest_MemberCriteriaId{MemberCriteriaId: "criteria-1"},
	}
	if got := criteriaRequest.ProtoReflect().WhichOneof(all.ContainingOneof()); got != criteriaID {
		t.Fatalf("explicit criteria target = %v, want member_criteria_id", got)
	}

	missingTarget := (&managev1.CreateCampaignRequest{}).ProtoReflect()
	if got := missingTarget.WhichOneof(all.ContainingOneof()); got != nil {
		t.Fatalf("missing target oneof = %v, want nil for API validation", got)
	}
}
