package queries

import (
	"context"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// GET /device-definitions/:id has no database of its own: it is the catalog
// read and the mapping of its error, which is the whole of what these check.
// The HTTP layer maps a not-found by type assertion (internal/api/common/
// config.go), which a wrapped sentinel never satisfies, so the handler has to
// do the errors.Is check itself.

const unknownDefinitionID = "ford_escapexx_2020"

func catalogMockFor(t *testing.T) *mock_gateways.MockDeviceDefinitionCatalogService {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	return mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
}

func TestGetDeviceDefinitionByIDAnswersNotFoundForAnUnknownID(t *testing.T) {
	catalog := catalogMockFor(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), unknownDefinitionID).
		Return(nil, nil, errors.Wrapf(gateways.ErrTemplateNotFound, "template %q", unknownDefinitionID))

	got, err := NewGetDeviceDefinitionByIDQueryHandler(catalog, nil).
		Handle(context.Background(), &GetDeviceDefinitionByIDQuery{DeviceDefinitionID: unknownDefinitionID})

	require.Error(t, err)
	assert.Nil(t, got)
	var notFound *exceptions.NotFoundError
	assert.ErrorAs(t, err, &notFound, "an unknown id is a 404, and the HTTP layer only maps this type")
}

func TestGetDeviceDefinitionByIDDoesNotReportACatalogOutageAsNotFound(t *testing.T) {
	catalog := catalogMockFor(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), unknownDefinitionID).
		Return(nil, nil, errors.New("unexpected status 502 from catalog"))

	got, err := NewGetDeviceDefinitionByIDQueryHandler(catalog, nil).
		Handle(context.Background(), &GetDeviceDefinitionByIDQuery{DeviceDefinitionID: unknownDefinitionID})

	require.Error(t, err)
	assert.Nil(t, got)
	var notFound *exceptions.NotFoundError
	assert.NotErrorAs(t, err, &notFound, "a catalog outage must not be reported as a missing definition")
}

func TestGetDeviceDefinitionByIDReturnsTheTemplate(t *testing.T) {
	catalog := catalogMockFor(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "toyota_camry_2026").Return(hydrateTemplate(), nil, nil)

	got, err := NewGetDeviceDefinitionByIDQueryHandler(catalog, nil).
		Handle(context.Background(), &GetDeviceDefinitionByIDQuery{DeviceDefinitionID: "toyota_camry_2026"})

	require.NoError(t, err)
	tmpl, ok := got.(*coremodels.Template)
	require.True(t, ok)
	assert.Equal(t, "toyota_camry_2026", tmpl.ID)
}

func TestGetDeviceDefinitionByIDV2AnswersNotFoundForAnUnknownID(t *testing.T) {
	catalog := catalogMockFor(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), unknownDefinitionID).
		Return(nil, nil, errors.Wrapf(gateways.ErrTemplateNotFound, "template %q", unknownDefinitionID))

	got, err := NewGetDeviceDefinitionByIDQueryV2Handler(catalog, nil).
		Handle(context.Background(), &GetDeviceDefinitionByIDQueryV2{DefinitionID: unknownDefinitionID})

	require.Error(t, err)
	assert.Nil(t, got)
	var notFound *exceptions.NotFoundError
	assert.ErrorAs(t, err, &notFound, "an unknown id is a 404, and the HTTP layer only maps this type")
}

func TestGetDeviceDefinitionByIDV2DoesNotReportACatalogOutageAsNotFound(t *testing.T) {
	catalog := catalogMockFor(t)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), unknownDefinitionID).
		Return(nil, nil, errors.New("unexpected status 502 from catalog"))

	got, err := NewGetDeviceDefinitionByIDQueryV2Handler(catalog, nil).
		Handle(context.Background(), &GetDeviceDefinitionByIDQueryV2{DefinitionID: unknownDefinitionID})

	require.Error(t, err)
	assert.Nil(t, got)
	var notFound *exceptions.NotFoundError
	assert.NotErrorAs(t, err, &notFound, "a catalog outage must not be reported as a missing definition")
}
