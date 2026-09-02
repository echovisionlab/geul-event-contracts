package contracttest_test

import (
	"encoding/json"
	"testing"

	intrav1 "github.com/echovisionlab/geul-event-contracts/gen/api/intra/v1"
	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
)

type formLocaleCatalogParityFixture struct {
	Valid      map[string]string `json:"valid"`
	IDs        map[string]string `json:"ids"`
	InvalidIDs []string          `json:"invalidIds"`
}

func TestFormLocaleCatalogSourceFixesStableIdentityHardCut(t *testing.T) {
	source := readRepositorySource(t, "config/content/form-locale-catalog.yaml")
	requireSourceContains(t, source,
		`pattern: "^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$"`,
		"reject_all_digits: true",
		"fixed_handle: document",
		`handle_prefix: "form:step:"`,
		`handle_prefix: "form:field:"`,
		`handle_prefix: "form:option:"`,
		`handle_prefix: "form:validator:"`,
		"locale_fields: [title, description]",
		"locale_fields: [label, description, placeholder, checkbox_label]",
		"locale_fields: [message]",
	)
}

func TestFormLocaleCatalogGoHandlesMatchGeneratedParityFixture(t *testing.T) {
	var fixture formLocaleCatalogParityFixture
	if err := json.Unmarshal([]byte(readRepositorySource(t, "packages/proto/gen/api/intra/v1/form_locale_catalog_parity.json")), &fixture); err != nil {
		t.Fatalf("decode Form locale parity fixture: %v", err)
	}

	step, err := intrav1.FormStepBlockHandle(fixture.IDs["step"])
	if err != nil {
		t.Fatal(err)
	}
	field, err := intrav1.FormFieldBlockHandle(fixture.IDs["field"])
	if err != nil {
		t.Fatal(err)
	}
	option, err := intrav1.FormOptionBlockHandle(fixture.IDs["option"])
	if err != nil {
		t.Fatal(err)
	}
	validator, err := intrav1.FormValidatorBlockHandle(fixture.IDs["validator"])
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]string{
		"root": intrav1.FormRootBlockHandle, "step": step, "field": field,
		"option": option, "validator": validator,
	}
	for kind, expected := range fixture.Valid {
		if actual[kind] != expected {
			t.Errorf("%s handle = %q, want %q", kind, actual[kind], expected)
		}
	}
	for _, id := range fixture.InvalidIDs {
		if err := intrav1.ValidateFormLocaleStableID(id); err == nil {
			t.Errorf("invalid stable ID %q was accepted", id)
		}
	}
}

func TestFormLocaleCatalogTargetsUseOnlyBlockOwnedLeafIdentity(t *testing.T) {
	stepID := "step_A-1"
	fieldID := "field_b-2"
	optionID := "option_C-3"
	validatorID := "validator_d-4"

	stepHandle, _ := intrav1.FormStepBlockHandle(stepID)
	fieldHandle, _ := intrav1.FormFieldBlockHandle(fieldID)
	optionHandle, _ := intrav1.FormOptionBlockHandle(optionID)
	validatorHandle, _ := intrav1.FormValidatorBlockHandle(validatorID)
	targets := []struct {
		target *managev1.AIDocumentFieldTarget
		block  string
		field  string
	}{
		{intrav1.FormRootTitleTarget(), intrav1.FormRootBlockHandle, intrav1.FormRootTitleFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormStepTitleTarget(stepID) }), stepHandle, intrav1.FormStepTitleFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormStepDescriptionTarget(stepID) }), stepHandle, intrav1.FormStepDescriptionFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormFieldLabelTarget(fieldID) }), fieldHandle, intrav1.FormFieldLabelFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormFieldDescriptionTarget(fieldID) }), fieldHandle, intrav1.FormFieldDescriptionFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormFieldPlaceholderTarget(fieldID) }), fieldHandle, intrav1.FormFieldPlaceholderFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormFieldCheckboxLabelTarget(fieldID) }), fieldHandle, intrav1.FormFieldCheckboxLabelFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) { return intrav1.FormOptionLabelTarget(optionID) }), optionHandle, intrav1.FormOptionLabelFieldHandle},
		{mustFormTarget(t, func() (*managev1.AIDocumentFieldTarget, error) {
			return intrav1.FormValidatorMessageTarget(validatorID)
		}), validatorHandle, intrav1.FormValidatorMessageFieldHandle},
	}
	for _, expected := range targets {
		if expected.target.GetBlockHandle() != expected.block || expected.target.GetFieldHandle() != expected.field {
			t.Errorf("target = %q/%q, want %q/%q", expected.target.GetBlockHandle(), expected.target.GetFieldHandle(), expected.block, expected.field)
		}
		if expected.target.GetRelationItem() != nil || len(expected.target.GetPath()) != 0 {
			t.Errorf("Form locale target must not use relation or path identity: %v", expected.target)
		}
	}
}

func mustFormTarget(t *testing.T, build func() (*managev1.AIDocumentFieldTarget, error)) *managev1.AIDocumentFieldTarget {
	t.Helper()
	target, err := build()
	if err != nil {
		t.Fatal(err)
	}
	return target
}
