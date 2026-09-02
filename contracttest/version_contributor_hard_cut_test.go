package contracttest_test

import (
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestContentVersionsExposeContributorsOnly(t *testing.T) {
	contributor := (&managev1.VersionContributor{}).ProtoReflect().Descriptor()
	requireMessageFieldCount(t, contributor, 2)
	memberID := requireMessageField(t, contributor, "member_id", 1, protoreflect.StringKind, "")
	if memberID.HasPresence() {
		t.Error("VersionContributor.member_id must be required by the application contract")
	}
	nickname := requireMessageField(t, contributor, "nickname", 2, protoreflect.StringKind, "")
	if nickname.HasPresence() {
		t.Error("VersionContributor.nickname must be required by the application contract")
	}

	versions := []struct {
		descriptor     protoreflect.MessageDescriptor
		contributorsNo protoreflect.FieldNumber
	}{
		{(&managev1.PostVersion{}).ProtoReflect().Descriptor(), 7},
		{(&managev1.PageVersion{}).ProtoReflect().Descriptor(), 7},
		{(&managev1.WorkVersion{}).ProtoReflect().Descriptor(), 7},
	}
	for _, version := range versions {
		field := requireMessageField(
			t,
			version.descriptor,
			"contributors",
			version.contributorsNo,
			protoreflect.MessageKind,
			"api.manage.v1.VersionContributor",
		)
		if field.Cardinality() != protoreflect.Repeated {
			t.Errorf("%s.contributors must be repeated", version.descriptor.FullName())
		}
	}
}

func TestContentVersionsExposeCanonicalSourceLocale(t *testing.T) {
	versions := []protoreflect.MessageDescriptor{
		(&managev1.PostVersion{}).ProtoReflect().Descriptor(),
		(&managev1.PageVersion{}).ProtoReflect().Descriptor(),
		(&managev1.WorkVersion{}).ProtoReflect().Descriptor(),
	}
	for _, version := range versions {
		requireMessageFieldCount(t, version, 8)
		field := requireMessageField(t, version, "source_locale", 8, protoreflect.StringKind, "")
		if field.HasPresence() || field.Cardinality() == protoreflect.Repeated {
			t.Errorf("%s.source_locale must be one canonical required scalar", version.FullName())
		}
	}
}

func TestCollaborationSaveContractsCarryActualContributors(t *testing.T) {
	postApply := (&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor()
	postBatch := requireMessageField(t, postApply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")
	requireContributorIDsField(t, postBatch.Message(), 6)
	requireContributorIDsField(t, (&intrav1.UpdatePostLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5)
	requireContributorIDsField(t, (&intrav1.UpdatePostDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 5)
	requireContributorIDsField(t, (&intrav1.CreatePostVersionCheckpointRequest{}).ProtoReflect().Descriptor(), 3)

	pageApply := (&intrav1.ApplyPageBlockBatchRequest{}).ProtoReflect().Descriptor()
	pageBatch := requireMessageField(t, pageApply, "batch", 2, protoreflect.MessageKind, "api.content.v1.PageSectionMutationBatch")
	requireContributorIDsField(t, pageBatch.Message(), 5)
	pageCheckpoint := (&intrav1.CreatePageVersionCheckpointRequest{}).ProtoReflect().Descriptor()
	requireContributorIDsField(t, pageCheckpoint, 3)

	workApply := (&intrav1.ApplyWorkBlockBatchRequest{}).ProtoReflect().Descriptor()
	workBatch := requireMessageField(t, workApply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")
	requireContributorIDsField(t, workBatch.Message(), 6)
	requireContributorIDsField(t, (&intrav1.UpdateWorkLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 6)
	requireContributorIDsField(t, (&intrav1.CreateWorkVersionCheckpointRequest{}).ProtoReflect().Descriptor(), 3)

	releaseApply := (&intrav1.ApplyReleaseBlockBatchRequest{}).ProtoReflect().Descriptor()
	releaseBatch := requireMessageField(t, releaseApply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")
	requireMessageField(t, releaseBatch.Message(), "expected_revision", 3, protoreflect.StringKind, "")
	requireContributorIDsField(t, releaseBatch.Message(), 6)
	requireContributorIDsField(t, (&intrav1.UpdateReleaseLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4)

	programEventApply := (&intrav1.ApplyProgramEventBlockBatchRequest{}).ProtoReflect().Descriptor()
	programEventBatch := requireMessageField(t, programEventApply, "batch", 2, protoreflect.MessageKind, "api.content.v1.RichTextBlockMutationBatch")
	requireContributorIDsField(t, programEventBatch.Message(), 6)
	requireContributorIDsField(t, (&intrav1.UpdateProgramEventLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5)

	nonVersioned := []struct {
		descriptor protoreflect.MessageDescriptor
		memberNo   protoreflect.FieldNumber
	}{
		{(&intrav1.UpdateArtistDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.UpdateArtistLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.UpdateLabelDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.UpdateLabelLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 4},
	}
	for _, contract := range nonVersioned {
		requireContributorIDsField(t, contract.descriptor, contract.memberNo)
	}

	updated := (&managev1.ContentUpdatedEvent{}).ProtoReflect().Descriptor()
	requireContributorIDsField(t, updated, 6)
	requireMessageField(t, updated, "document_revision", 5, protoreflect.StringKind, "")
}

func requireContributorIDsField(
	t *testing.T,
	descriptor protoreflect.MessageDescriptor,
	memberNumber protoreflect.FieldNumber,
) {
	t.Helper()
	field := requireMessageField(t, descriptor, "contributor_member_ids", memberNumber, protoreflect.StringKind, "")
	if field.Cardinality() != protoreflect.Repeated {
		t.Errorf("%s.contributor_member_ids must be repeated", descriptor.FullName())
	}
}
