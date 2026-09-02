package contracttest_test

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	_ "github.com/echovisionlab/geul-event-contracts/gen/api/common/v1"
	_ "github.com/echovisionlab/geul-event-contracts/gen/api/content/v1"
	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	_ "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	_ "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	_ "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestRepositoryProtoWireNumbersAreDenseWithOnlyExplicitHardCutReservations(t *testing.T) {
	files := repositoryProtoFiles(t)
	for _, file := range files {
		file := file
		t.Run(file.Path(), func(t *testing.T) {
			for i := 0; i < file.Services().Len(); i++ {
				service := file.Services().Get(i)
				if service.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
					t.Errorf("service %s has a deprecated option", service.FullName())
				}
				for methodIndex := 0; methodIndex < service.Methods().Len(); methodIndex++ {
					method := service.Methods().Get(methodIndex)
					if method.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
						t.Errorf("method %s has a deprecated option", method.FullName())
					}
				}
			}
			for i := 0; i < file.Extensions().Len(); i++ {
				checkFieldIsNotDeprecated(t, file.Extensions().Get(i))
			}
			for i := 0; i < file.Enums().Len(); i++ {
				checkDenseEnumNumbers(t, file.Enums().Get(i))
			}
			for i := 0; i < file.Messages().Len(); i++ {
				checkDenseMessageNumbers(t, file.Messages().Get(i))
			}
		})
	}
}

func TestRemovedTranslationAuthoringSymbolsStayAbsent(t *testing.T) {
	for _, name := range removedTranslationAuthoringSymbols() {
		if descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(name); err == nil {
			t.Errorf("removed proto symbol %s survives as %T", name, descriptor)
		}
	}
}

func TestRemovedGenericRenderQueueMessagesStayAbsent(t *testing.T) {
	for _, name := range []protoreflect.FullName{
		"api.manage.v1.PostRenderRequest",
		"api.manage.v1.WorkRenderRequest",
		"api.manage.v1.ProgramEventRenderRequest",
		"api.manage.v1.PrivacyRenderRequest",
		"api.manage.v1.TermsRenderRequest",
	} {
		if descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(name); err == nil {
			t.Errorf("removed proto symbol %s survives as %T", name, descriptor)
		}
	}
}

func TestCollaborationLocaleWireHardCut(t *testing.T) {
	contracts := []struct {
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{(&intrav1.CollaborationResource{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.AuthorizeCollaborationResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadPostBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor(), 7},
		{(&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyPostBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePostLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.UpdatePostLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePostDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.UpdatePostDocumentMetadataResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.CreatePostVersionCheckpointRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.CreatePostVersionCheckpointResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadPageBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadPageBlockDocumentResponse{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.ApplyPageBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyPageBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePageLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 7},
		{(&intrav1.UpdatePageLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePageDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePageDocumentMetadataResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.CreatePageVersionCheckpointRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.CreatePageVersionCheckpointResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadWorkBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadWorkBlockDocumentResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.ApplyWorkBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyWorkBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateWorkLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 7},
		{(&intrav1.UpdateWorkLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.CreateWorkVersionCheckpointRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.CreateWorkVersionCheckpointResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadProgramEventBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadProgramEventBlockDocumentResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.ApplyProgramEventBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyProgramEventBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateProgramEventLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.UpdateProgramEventLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadArtistBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadArtistBlockDocumentResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.ApplyArtistBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyArtistBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateArtistLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateArtistLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateArtistDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateArtistDocumentMetadataResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadLabelBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadLabelBlockDocumentResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.ApplyLabelBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyLabelBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateLabelLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateLabelLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateLabelDocumentMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateLabelDocumentMetadataResponse{}).ProtoReflect().Descriptor(), 4},

		{(&intrav1.LoadReleaseBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadReleaseBlockDocumentResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.ApplyReleaseBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyReleaseBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateReleaseLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.UpdateReleaseLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadCampaignDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadCampaignDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.ApplyCampaignBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyCampaignBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateCampaignLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateCampaignLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadEmailTemplateDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadEmailTemplateDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.ApplyEmailTemplateBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyEmailTemplateBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateEmailTemplateLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateEmailTemplateLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadPrivacyBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadPrivacyBlockDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.ApplyPrivacyBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyPrivacyBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePrivacyLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdatePrivacyLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadTermsBlockDocumentRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.LoadTermsBlockDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.ApplyTermsBlockBatchRequest{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.ApplyTermsBlockBatchResponse{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateTermsLocaleMetadataRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.UpdateTermsLocaleMetadataResponse{}).ProtoReflect().Descriptor(), 5},

		{(&intrav1.LoadFormDocumentRequest{}).ProtoReflect().Descriptor(), 2},
		{(&intrav1.LoadFormDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.SaveFormDocumentResponse{}).ProtoReflect().Descriptor(), 2},
		{(&intrav1.LoadEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor(), 2},
		{(&intrav1.LoadEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor(), 4},
		{(&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.SaveEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor(), 2},
		{(&intrav1.LoadMapThemeSnapshotRequest{}).ProtoReflect().Descriptor(), 2},
		{(&intrav1.LoadMapThemeSnapshotResponse{}).ProtoReflect().Descriptor(), 3},
		{(&intrav1.SaveMapThemeSnapshotRequest{}).ProtoReflect().Descriptor(), 5},
		{(&intrav1.SaveMapThemeSnapshotResponse{}).ProtoReflect().Descriptor(), 3},
	}

	for _, contract := range contracts {
		field := contract.descriptor.Fields().ByName("locale")
		if field == nil {
			t.Errorf("%s.locale is missing", contract.descriptor.FullName())
			continue
		}
		if field.Number() != contract.number || field.Kind() != protoreflect.StringKind || field.HasPresence() || field.Cardinality() == protoreflect.Repeated {
			t.Errorf("%s.locale = number %d kind %s presence %t cardinality %s, want required string %d", contract.descriptor.FullName(), field.Number(), field.Kind(), field.HasPresence(), field.Cardinality(), contract.number)
		}
	}

	emailLayoutSave := (&intrav1.SaveEmailLayoutDocumentResponse{}).ProtoReflect().Descriptor()
	requireField(t, emailLayoutSave, "document_revision", 3, protoreflect.StringKind)
	requireOptionalField(t, emailLayoutSave, "target_revision", 4, protoreflect.StringKind)
	requireNoField(t, emailLayoutSave, "edit_hash")

	for _, guard := range []struct {
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{(&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor(), 6},
	} {
		field := guard.descriptor.Fields().ByName("expected_document_revision")
		if field == nil || field.Number() != guard.number || field.Kind() != protoreflect.StringKind || field.HasPresence() || field.Cardinality() == protoreflect.Repeated {
			t.Errorf("%s.expected_document_revision must be the required shared/source CAS guard at field %d", guard.descriptor.FullName(), guard.number)
		}
	}
	requireOptionalField(t, (&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor(), "expected_target_revision", 7, protoreflect.StringKind)
	requireOptionalField(t, (&intrav1.SaveEmailLayoutDocumentRequest{}).ProtoReflect().Descriptor(), "expected_target_revision", 7, protoreflect.StringKind)

	for _, presence := range []struct {
		descriptor protoreflect.MessageDescriptor
		number     protoreflect.FieldNumber
	}{
		{(&intrav1.LoadPostBlockDocumentResponse{}).ProtoReflect().Descriptor(), 11},
		{(&intrav1.LoadPageBlockDocumentResponse{}).ProtoReflect().Descriptor(), 10},
		{(&intrav1.LoadWorkBlockDocumentResponse{}).ProtoReflect().Descriptor(), 9},
		{(&intrav1.LoadProgramEventBlockDocumentResponse{}).ProtoReflect().Descriptor(), 9},
		{(&intrav1.LoadArtistBlockDocumentResponse{}).ProtoReflect().Descriptor(), 9},
		{(&intrav1.LoadLabelBlockDocumentResponse{}).ProtoReflect().Descriptor(), 9},
		{(&intrav1.LoadReleaseBlockDocumentResponse{}).ProtoReflect().Descriptor(), 9},
		{(&intrav1.LoadCampaignDocumentResponse{}).ProtoReflect().Descriptor(), 8},
		{(&intrav1.LoadEmailTemplateDocumentResponse{}).ProtoReflect().Descriptor(), 8},
		{(&intrav1.LoadPrivacyBlockDocumentResponse{}).ProtoReflect().Descriptor(), 8},
		{(&intrav1.LoadTermsBlockDocumentResponse{}).ProtoReflect().Descriptor(), 8},
		{(&intrav1.LoadFormDocumentResponse{}).ProtoReflect().Descriptor(), 6},
		{(&intrav1.SaveFormDocumentRequest{}).ProtoReflect().Descriptor(), 5},
	} {
		field := presence.descriptor.Fields().ByName("present_locale_values")
		if field == nil || field.Number() != presence.number || field.Kind() != protoreflect.MessageKind ||
			field.Cardinality() != protoreflect.Repeated || field.Message().FullName() != "api.manage.v1.AIDocumentFieldTarget" {
			t.Errorf("%s.present_locale_values must be repeated api.manage.v1.AIDocumentFieldTarget at field %d", presence.descriptor.FullName(), presence.number)
		}
	}

	for _, affected := range []protoreflect.MessageDescriptor{
		(&intrav1.ApplyPostBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyPageBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyWorkBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyProgramEventBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyArtistBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyLabelBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyReleaseBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyCampaignBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyEmailTemplateBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyPrivacyBlockBatchRequest{}).ProtoReflect().Descriptor(),
		(&intrav1.ApplyTermsBlockBatchRequest{}).ProtoReflect().Descriptor(),
	} {
		field := affected.Fields().ByName("affected_locale_values")
		if field == nil || field.Number() != 5 || field.Kind() != protoreflect.MessageKind ||
			field.Cardinality() != protoreflect.Repeated || field.Message().FullName() != "api.manage.v1.AIDocumentFieldTarget" {
			t.Errorf("%s.affected_locale_values must be repeated api.manage.v1.AIDocumentFieldTarget at field 5", affected.FullName())
		}
	}

	relay := (&intrav1.RelayInteractiveAIDocumentMutationRequest{}).ProtoReflect().Descriptor()
	if relay.Fields().Len() != 10 {
		t.Errorf("%s fields = %d, want 10 revision-only relay fields", relay.FullName(), relay.Fields().Len())
	}
	requireNoField(t, relay, "accepted_source_epoch")
}

func repositoryProtoFiles(t *testing.T) []protoreflect.FileDescriptor {
	t.Helper()

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test source path")
	}
	protoRoot := filepath.Join(filepath.Dir(sourceFile), "..", "proto")
	sourcePaths := make(map[string]struct{})
	if err := filepath.WalkDir(protoRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".proto" {
			return nil
		}
		relative, err := filepath.Rel(protoRoot, path)
		if err != nil {
			return err
		}
		sourcePaths[filepath.ToSlash(relative)] = struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("inventory proto sources: %v", err)
	}

	descriptors := make(map[string]protoreflect.FileDescriptor, len(sourcePaths))
	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if _, owned := sourcePaths[file.Path()]; owned {
			descriptors[file.Path()] = file
		}
		return true
	})

	missing := make([]string, 0)
	for path := range sourcePaths {
		if descriptors[path] == nil {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("generated descriptors missing for proto sources: %v", missing)
	}

	paths := make([]string, 0, len(descriptors))
	for path := range descriptors {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	files := make([]protoreflect.FileDescriptor, 0, len(paths))
	for _, path := range paths {
		files = append(files, descriptors[path])
	}
	return files
}

func checkDenseMessageNumbers(t *testing.T, message protoreflect.MessageDescriptor) {
	t.Helper()
	t.Run(string(message.FullName()), func(t *testing.T) {
		if message.Options().(*descriptorpb.MessageOptions).GetDeprecated() {
			t.Error("deprecated message option is forbidden")
		}
		if message.FullName() == "api.intra.v1.ProgramEventLocaleMetadata" {
			assertProgramEventSourceMetadataReservations(t, message)
		} else {
			if message.ReservedRanges().Len() != 0 {
				t.Errorf("reserved ranges = %v, want none", message.ReservedRanges())
			}
			if message.ReservedNames().Len() != 0 {
				t.Errorf("reserved names = %v, want none", message.ReservedNames())
			}
		}

		for i := 0; i < message.Fields().Len(); i++ {
			field := message.Fields().Get(i)
			checkFieldIsNotDeprecated(t, field)
			if want := protoreflect.FieldNumber(i + 1); field.Number() != want {
				t.Errorf("field %s is declaration %d with number %d, want %d", field.FullName(), i, field.Number(), want)
			}
		}
		for i := 0; i < message.Extensions().Len(); i++ {
			checkFieldIsNotDeprecated(t, message.Extensions().Get(i))
		}

		for i := 0; i < message.Enums().Len(); i++ {
			checkDenseEnumNumbers(t, message.Enums().Get(i))
		}
		for i := 0; i < message.Messages().Len(); i++ {
			checkDenseMessageNumbers(t, message.Messages().Get(i))
		}
	})
}

func checkDenseEnumNumbers(t *testing.T, enum protoreflect.EnumDescriptor) {
	t.Helper()
	t.Run(string(enum.FullName()), func(t *testing.T) {
		if enum.Options().(*descriptorpb.EnumOptions).GetDeprecated() {
			t.Error("deprecated enum option is forbidden")
		}
		if enum.FullName() == "api.intra.v1.CollaborationConflictReason" {
			assertCollaborationConflictReasonHardCut(t, enum)
			return
		}
		if enum.ReservedRanges().Len() != 0 {
			t.Errorf("reserved ranges = %v, want none", enum.ReservedRanges())
		}
		if enum.ReservedNames().Len() != 0 {
			t.Errorf("reserved names = %v, want none", enum.ReservedNames())
		}

		for i := 0; i < enum.Values().Len(); i++ {
			value := enum.Values().Get(i)
			if value.Options().(*descriptorpb.EnumValueOptions).GetDeprecated() {
				t.Errorf("enum value %s has a deprecated option", value.FullName())
			}
			if want := protoreflect.EnumNumber(i); value.Number() != want {
				t.Errorf("enum value %s is declaration %d with number %d, want %d", value.FullName(), i, value.Number(), want)
			}
		}
	})
}

func assertCollaborationConflictReasonHardCut(t *testing.T, enum protoreflect.EnumDescriptor) {
	t.Helper()
	ranges := enum.ReservedRanges()
	if ranges.Len() != 3 {
		t.Fatalf("CollaborationConflictReason reserved ranges = %v, want 2, 3, 4", ranges)
	}
	for index, want := range []protoreflect.EnumNumber{2, 3, 4} {
		if ranges.Get(index)[0] != want || ranges.Get(index)[1] != want {
			t.Fatalf("CollaborationConflictReason reserved ranges = %v, want 2, 3, 4", ranges)
		}
	}
	wantNames := []protoreflect.Name{
		"COLLABORATION_CONFLICT_REASON_SOURCE_REVISION_CHANGED",
		"COLLABORATION_CONFLICT_REASON_SHARED_STRUCTURE_CHANGED",
		"COLLABORATION_CONFLICT_REASON_SHARED_MATERIALIZATION_CHANGED",
	}
	if enum.ReservedNames().Len() != len(wantNames) {
		t.Fatalf("CollaborationConflictReason reserved names = %v", enum.ReservedNames())
	}
	for index, want := range wantNames {
		if enum.ReservedNames().Get(index) != want {
			t.Fatalf("CollaborationConflictReason reserved name %d = %s, want %s", index, enum.ReservedNames().Get(index), want)
		}
	}
	wantNumbers := []protoreflect.EnumNumber{0, 1, 5, 6}
	for index, want := range wantNumbers {
		value := enum.Values().Get(index)
		if value.Number() != want {
			t.Fatalf("CollaborationConflictReason value %s = %d, want %d", value.Name(), value.Number(), want)
		}
	}
}

func checkFieldIsNotDeprecated(t *testing.T, field protoreflect.FieldDescriptor) {
	t.Helper()
	if field.Options().(*descriptorpb.FieldOptions).GetDeprecated() {
		t.Errorf("field %s has a deprecated option", field.FullName())
	}
}

func removedTranslationAuthoringSymbols() []protoreflect.FullName {
	return []protoreflect.FullName{
		"api.manage.v1.AccountSecurity.delivery_primary_email",
		"api.manage.v1.AccountCurrentVerifiedEmailRecipient",
		"api.manage.v1.SendEmailEvent.account_current_verified_email",
		"api.intra.v1.InternalPostService.UpdatePostSourceMetadata",
		"api.intra.v1.PostSourceMetadata",
		"api.intra.v1.UpdatePostSourceMetadataRequest",
		"api.intra.v1.UpdatePostSourceMetadataResponse",
		"api.intra.v1.InternalPageService.UpdatePageSourceMetadata",
		"api.intra.v1.PageSourceMetadata",
		"api.intra.v1.UpdatePageSourceMetadataRequest",
		"api.intra.v1.UpdatePageSourceMetadataResponse",
		"api.intra.v1.InternalWorkService.UpdateWorkSourceMetadata",
		"api.intra.v1.WorkSourceMetadata",
		"api.intra.v1.UpdateWorkSourceMetadataRequest",
		"api.intra.v1.UpdateWorkSourceMetadataResponse",
		"api.intra.v1.InternalProgramEventService.UpdateProgramEventSourceMetadata",
		"api.intra.v1.ProgramEventSourceMetadata",
		"api.intra.v1.ProgramEventDocumentLocaleMetadata",
		"api.intra.v1.UpdateProgramEventSourceMetadataRequest",
		"api.intra.v1.UpdateProgramEventSourceMetadataResponse",
		"api.intra.v1.InternalArtistService.UpdateArtistSourceMetadata",
		"api.intra.v1.ArtistSourceMetadata",
		"api.intra.v1.UpdateArtistSourceMetadataRequest",
		"api.intra.v1.UpdateArtistSourceMetadataResponse",
		"api.intra.v1.InternalLabelService.UpdateLabelSourceMetadata",
		"api.intra.v1.LabelSourceMetadata",
		"api.intra.v1.UpdateLabelSourceMetadataRequest",
		"api.intra.v1.UpdateLabelSourceMetadataResponse",
		"api.intra.v1.InternalReleaseService.UpdateReleaseSourceMetadata",
		"api.intra.v1.ReleaseSourceMetadata",
		"api.intra.v1.UpdateReleaseSourceMetadataRequest",
		"api.intra.v1.UpdateReleaseSourceMetadataResponse",
		"api.intra.v1.InternalCampaignService.UpdateSourceMetadata",
		"api.intra.v1.CampaignSourceMetadata",
		"api.intra.v1.UpdateCampaignSourceMetadataRequest",
		"api.intra.v1.UpdateCampaignSourceMetadataResponse",
		"api.intra.v1.InternalCampaignService.UpdateDocumentMetadata",
		"api.intra.v1.CampaignDocumentLocaleMetadata",
		"api.intra.v1.UpdateCampaignDocumentMetadataRequest",
		"api.intra.v1.UpdateCampaignDocumentMetadataResponse",
		"api.intra.v1.InternalEmailTemplateService.UpdateSourceMetadata",
		"api.intra.v1.EmailTemplateSourceMetadata",
		"api.intra.v1.UpdateEmailTemplateSourceMetadataRequest",
		"api.intra.v1.UpdateEmailTemplateSourceMetadataResponse",
		"api.intra.v1.InternalEmailTemplateService.UpdateDocumentMetadata",
		"api.intra.v1.EmailTemplateDocumentLocaleMetadata",
		"api.intra.v1.UpdateEmailTemplateDocumentMetadataRequest",
		"api.intra.v1.UpdateEmailTemplateDocumentMetadataResponse",
		"api.intra.v1.InternalPrivacyService.UpdatePrivacySourceMetadata",
		"api.intra.v1.PrivacySourceMetadata",
		"api.intra.v1.UpdatePrivacySourceMetadataRequest",
		"api.intra.v1.UpdatePrivacySourceMetadataResponse",
		"api.intra.v1.InternalTermsService.UpdateTermsSourceMetadata",
		"api.intra.v1.TermsSourceMetadata",
		"api.intra.v1.UpdateTermsSourceMetadataRequest",
		"api.intra.v1.UpdateTermsSourceMetadataResponse",
		"api.intra.v1.InternalEmailLayoutService.SaveTranslationDocument",
		"api.intra.v1.InternalEmailLayoutService.LoadTranslationDocument",
		"api.intra.v1.SaveEmailLayoutTranslationDocumentRequest",
		"api.intra.v1.SaveEmailLayoutTranslationDocumentResponse",
		"api.intra.v1.LoadEmailLayoutTranslationDocumentRequest",
		"api.intra.v1.LoadEmailLayoutTranslationDocumentResponse",
		"api.intra.v1.InternalFormService.SaveTranslationDocument",
		"api.intra.v1.InternalFormService.LoadTranslationDocument",
		"api.intra.v1.SaveFormTranslationDocumentRequest",
		"api.intra.v1.SaveFormTranslationDocumentResponse",
		"api.intra.v1.LoadFormTranslationDocumentRequest",
		"api.intra.v1.LoadFormTranslationDocumentResponse",
		"api.manage.v1.ProgramEventLocale.is_source_locale",
		"api.manage.v1.TranslationService.EnsureEntityTranslation",
		"api.manage.v1.TranslationService.UpdateEntityTranslation",
		"api.manage.v1.TranslationService.DeleteEntityTranslation",
		"api.manage.v1.TranslationService.DeleteBlockDocumentLocale",
		"api.manage.v1.EnsureEntityTranslationRequest",
		"api.manage.v1.EnsureEntityTranslationResponse",
		"api.manage.v1.UpdateEntityTranslationRequest",
		"api.manage.v1.UpdateEntityTranslationResponse",
		"api.manage.v1.DeleteEntityTranslationRequest",
		"api.manage.v1.DeleteEntityTranslationResponse",
		"api.manage.v1.DeleteBlockDocumentLocaleRequest",
		"api.manage.v1.DeleteBlockDocumentLocaleResponse",
		"api.manage.v1.TranslationEntry.manual_edited",
		"api.manage.v1.TranslationEntry.edit_version",
		"api.manage.v1.TranslationEntry.edited_at",
		"api.manage.v1.TranslationEntry.edited_by_member_id",
		"api.manage.v1.TranslationEntry.source_hash",
		"api.manage.v1.TranslationEntry.source_revision",
		"api.manage.v1.TranslationEntry.source_epoch",
		"api.manage.v1.TranslationEntry.status",
		"api.manage.v1.TranslationEntry.published_at",
		"api.manage.v1.TranslationEntry.translation_spec_version",
		"api.manage.v1.TranslationEntry.machine_generated",
		"api.manage.v1.TranslationEntry.provider",
		"api.manage.v1.TranslationEntry.model",
		"api.manage.v1.TranslationEntryStatus",
		"api.manage.v1.TranslationJob.translation_spec_version",
		"api.manage.v1.TranslationJob.superseded_by_revision",
		"api.manage.v1.TranslationJob.superseded_by_source_epoch",
		"api.manage.v1.TRANSLATION_JOB_STATUS_PUBLISHED",
		"api.manage.v1.TRANSLATION_JOB_STATUS_SUPERSEDED",
		"api.manage.v1.TRANSLATION_JOB_STATUS_OBSOLETE",
		"api.manage.v1.TRANSLATION_FAILURE_REASON_SOURCE_NO_LONGER_CURRENT",
		"api.manage.v1.TranslationOverviewStats.published_entries",
		"api.manage.v1.TranslationOverviewStats.stale_entries",
		"api.manage.v1.TranslationLocaleHealth.published_entries",
		"api.manage.v1.TranslationLocaleHealth.stale_entries",
		"api.manage.v1.TranslationLocaleHealth.last_published_at",
		"api.manage.v1.TranslationEntityHealth.published_entries",
		"api.manage.v1.TranslationEntityHealth.stale_entries",
		"api.manage.v1.TranslationSettings.english_fallback_enabled",
		"api.manage.v1.TranslationSettings.stale_exact_enabled",
		"api.manage.v1.TranslationSettings.stale_english_enabled",
		"api.manage.v1.TranslationSettings.machine_generated_public_serve",
		"api.open.v1.LocalizationInfo.is_stale",
		"api.manage.v1.TRANSLATION_ENTITY_TYPE_SITE_SETTING",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_DELETED",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_QUEUED",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_RUNNING",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_PUBLISHED",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_CANCELLED",
		"api.manage.v1.TRANSLATION_LIFECYCLE_STATUS_SUPERSEDED",
		"api.open.v1.LOCALIZATION_FALLBACK_REASON_STALE_EXACT_LOCALE",
		"api.open.v1.LOCALIZATION_FALLBACK_REASON_ENGLISH",
		"api.open.v1.LOCALIZATION_FALLBACK_REASON_STALE_ENGLISH",
	}
}
