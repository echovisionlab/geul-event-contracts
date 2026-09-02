package contracttest_test

import (
	"testing"

	managev1 "github.com/echovisionlab/geul-event-contracts/gen/api/manage/v1"
	openv1 "github.com/echovisionlab/geul-event-contracts/gen/api/open/v1"
	policyv1 "github.com/echovisionlab/geul-event-contracts/gen/api/policy/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestPostSeriesLifecycleContractIsDraftAndPublishedOnly(t *testing.T) {
	status := managev1.File_api_manage_v1_series_proto.Enums().ByName("SeriesStatus")
	if status == nil {
		t.Fatal("api.manage.v1.SeriesStatus is missing")
	}

	expected := map[protoreflect.Name]protoreflect.EnumNumber{
		"SERIES_STATUS_UNSPECIFIED": 0,
		"SERIES_STATUS_DRAFT":       1,
		"SERIES_STATUS_PUBLISHED":   2,
	}
	if status.Values().Len() != len(expected) {
		t.Fatalf("SeriesStatus must contain only unspecified, draft, and published; got %d values", status.Values().Len())
	}
	for name, number := range expected {
		value := status.Values().ByName(name)
		if value == nil || value.Number() != number {
			t.Errorf("SeriesStatus.%s must remain value %d", name, number)
		}
	}

}

func TestPublicPostSeriesOrderingProjectionIsExplicit(t *testing.T) {
	postSummary := (&openv1.PostSummary{}).ProtoReflect().Descriptor()
	seriesOrder := postSummary.Fields().ByName("series_order")
	if seriesOrder == nil || seriesOrder.Number() != 11 || seriesOrder.Cardinality() != protoreflect.Optional {
		t.Fatal("api.open.v1.PostSummary.series_order must remain optional field 11")
	}
}

func TestPostSeriesGatewayTiersPreserveExactServiceAuthority(t *testing.T) {
	service := managev1.File_api_manage_v1_series_proto.Services().ByName("SeriesService")

	for _, method := range []protoreflect.Name{
		"ListSeriesPosts",
		"ListMySeries",
		"GetSeriesWithManagers",
		"UpdateSeries",
		"SetSeriesFeaturedImage",
		"DeleteSeriesFeaturedImage",
		"ListSeriesManagers",
		"AssignPostToSeries",
		"UnassignPostFromSeries",
		"ReorderSeriesPosts",
		"CheckSeriesSlugAvailable",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_USER)
	}

	for _, method := range []protoreflect.Name{
		"ListSeriesAdmin",
		"ListSeriesSimple",
		"CreateSeries",
		"DeleteSeries",
		"AddSeriesManager",
		"RemoveSeriesManager",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_ADMIN)
	}
}

func TestPostSeriesTranslationGatewayDefersExactAuthorityToAPI(t *testing.T) {
	service := managev1.File_api_manage_v1_translation_proto.Services().ByName("TranslationService")

	for _, method := range []protoreflect.Name{
		"SetEntitySourceLocale",
		"ListTranslationLocales",
		"ListEntityTranslations",
		"GetEntityTranslation",
		"ListTranslationJobs",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_USER)
	}

	for _, method := range []protoreflect.Name{
		"RegenerateEntityTranslations",
		"CancelTranslationJob",
	} {
		requireMethodTier(t, service, method, policyv1.AuthorizationRole_USER)
	}
}
