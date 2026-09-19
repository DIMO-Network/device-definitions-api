package services

import (
	"context"
	"errors"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// GetVIN's vendor extra is what the decode handler writes to
// failed_vin_decodes when nothing could decode the VIN. Returning nil there
// left the caller dereferencing nothing, so the decode panicked, recorded no
// failure, and repeated the whole vendor fan-out on the next request.
//
// These paths need no Postgres: they return before any provider that touches
// the database.

func vendorExtraService(t *testing.T) (vinDecodingService, *mock_gateways.MockDeviceDefinitionCatalogService, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	logger := zerolog.Nop()
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	return vinDecodingService{logger: &logger, catalogSvc: catalog}, catalog, ctrl
}

// The guard that refuses a VIN before any provider is tried.
func TestGetVINReturnsAVendorExtraForAnInvalidVIN(t *testing.T) {
	svc, _, _ := vendorExtraService(t)

	info, extra, err := svc.GetVIN(context.Background(), "NOTAVIN", coremodels.AllProviders, "USA")

	require.Error(t, err)
	assert.Nil(t, info)
	require.NotNil(t, extra, "the caller records the failure from the vendor extra; it may never be nil")
	assert.Empty(t, extra.VendorsTried)
}

// The 0SC test-VIN branch reads a template from the catalog, so it now fails
// whenever the catalog does -- including the not-found this branch sees
// wherever the template import has not run.
func TestGetVINReturnsAVendorExtraWhenTheTestVINTemplateIsUnavailable(t *testing.T) {
	svc, catalog, _ := vendorExtraService(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "ford_escape_2020").
		Return(nil, nil, errors.New("template not found"))

	info, extra, err := svc.GetVIN(context.Background(), "0SC12345678901234", coremodels.AllProviders, "USA")

	require.Error(t, err)
	assert.Nil(t, info)
	require.NotNil(t, extra, "the caller records the failure from the vendor extra; it may never be nil")
}

// A successful test-VIN decode returns one too, so no caller has to care which
// branch produced the answer.
func TestGetVINReturnsAVendorExtraForADecodedTestVIN(t *testing.T) {
	svc, catalog, _ := vendorExtraService(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "ford_escape_2020").
		Return(&coremodels.Template{
			ID:           "ford_escape_2020",
			Model:        "Escape",
			Year:         2020,
			DeviceType:   "vehicle",
			Manufacturer: coremodels.TemplateManufacturer{Slug: "ford", Name: "Ford"},
		}, nil, nil)

	info, extra, err := svc.GetVIN(context.Background(), "0SC12345678901234", coremodels.AllProviders, "USA")

	require.NoError(t, err)
	require.NotNil(t, info)
	require.NotNil(t, extra)
}

// A 10-character Japan chassis number is one the handler admits and
// IsValidJapanChassis accepts. The provider guard excluded exactly 10, so it
// fell through to the 17-character VIN validator and was refused before any
// provider was tried.
func TestGetVINSendsATenCharacterChassisToTheJapanProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.Nop()
	carvx := mock_gateways.NewMockCarVxVINAPI(ctrl)
	japan := mock_gateways.NewMockJapan17VINAPI(ctrl)
	carvx.EXPECT().GetVINInfo("ZWR980001").Times(0)
	carvx.EXPECT().GetVINInfo("ZWR9-80001").Return(nil, nil, errors.New("carvx is down"))
	japan.EXPECT().GetVINInfo("ZWR9-80001").Return(nil, nil, errors.New("japan17 is down"))
	svc := vinDecodingService{logger: &logger, carvxAPI: carvx, japan17VINAPI: japan}

	info, extra, err := svc.GetVIN(context.Background(), "ZWR9-80001", coremodels.AllProviders, "JPN")

	require.Error(t, err)
	assert.Nil(t, info)
	require.NotNil(t, extra)
	assert.Equal(t, []string{string(coremodels.CarVXVIN), string(coremodels.Japan17VIN)}, extra.VendorsTried,
		"a 10 character chassis number belongs to the Japan providers, not the VIN validator")
	assert.NotContains(t, err.Error(), "invalid vin")
}

// The same chassis with no country hint: the length alone has to route it.
func TestGetVINSendsATenCharacterChassisToTheJapanProvidersWithoutACountry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logger := zerolog.Nop()
	carvx := mock_gateways.NewMockCarVxVINAPI(ctrl)
	japan := mock_gateways.NewMockJapan17VINAPI(ctrl)
	carvx.EXPECT().GetVINInfo("ZWR9-80001").Return(nil, nil, errors.New("carvx is down"))
	japan.EXPECT().GetVINInfo("ZWR9-80001").Return(nil, nil, errors.New("japan17 is down"))
	svc := vinDecodingService{logger: &logger, carvxAPI: carvx, japan17VINAPI: japan}

	_, extra, err := svc.GetVIN(context.Background(), "ZWR9-80001", coremodels.AllProviders, "")

	require.Error(t, err)
	require.NotNil(t, extra)
	assert.Equal(t, []string{string(coremodels.CarVXVIN), string(coremodels.Japan17VIN)}, extra.VendorsTried)
}
