package commands

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// The admin create builds the template id from the make and model it was
// given. definitions-worker answers 422 for an id its ID_RE refuses, that
// error propagated as a plain error, and fiber reported it as a 500 -- the
// service telling an operator it is broken when the request is what needs
// correcting. The decode path returns a typed not-found for the very same
// input.
//
// No Postgres: sqlmock answers the device_types read and every gateway is a
// mock. Only the write against a real catalog needs CI.

func createHandler(t *testing.T) (CreateDeviceDefinitionCommandHandler, sqlmock.Sqlmock, *mock_gateways.MockDeviceDefinitionCatalogService) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	rw := &db.ReaderWriter{Reader: &db.DB{DB: sqlDB}, Writer: &db.DB{DB: sqlDB}}

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	identity := mock_gateways.NewMockIdentityAPI(ctrl)
	fuel := mock_gateways.NewMockFuelAPIService(ctrl)
	powertrain := mock_services.NewMockPowerTrainTypeService(ctrl)
	logger := zerolog.Nop()

	identity.EXPECT().GetManufacturer(gomock.Any()).
		Return(&coremodels.Manufacturer{TokenID: 131, Name: "Toyota"}, nil).AnyTimes()
	powertrain.EXPECT().ResolvePowerTrainType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return("ICE", nil).AnyTimes()
	fuel.EXPECT().FetchDeviceImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(gateways.FuelDeviceImages{}, nil).AnyTimes()

	h := NewCreateDeviceDefinitionCommandHandler(catalog, func() *db.ReaderWriter { return rw }, powertrain, fuel, &logger, identity)
	return h, mock, catalog
}

func expectDeviceTypeRead(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT .* FROM .*device_types.* WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("vehicle", "Vehicle"))
}

// A model written in a script with no character in the worker's id class
// leaves the model segment empty after the repair, so no template can ever be
// stored at the id. The answer is a 4xx, and nothing is sent to the catalog.
func TestCreateDeviceDefinitionAnswersAValidationErrorForAnUnmintableID(t *testing.T) {
	h, mock, catalog := createHandler(t)
	expectDeviceTypeRead(mock)
	// The premise of the finding: this must never reach the worker, whose 422
	// dd-api cannot act on.
	catalog.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	res, err := h.Handle(context.Background(), &CreateDeviceDefinitionCommand{
		Make:  "Toyota",
		Model: "ハイエース",
		Year:  2020,
	})

	require.Error(t, err)
	assert.Nil(t, res)
	var validation *exceptions.ValidationError
	require.ErrorAs(t, err, &validation, "an id the worker can never hold is a bad request, not a 500")
	var internal *exceptions.InternalError
	assert.False(t, errors.As(err, &internal), "it must not be reported as a server fault")
	assert.Contains(t, err.Error(), "ハイエース")
}

// The ordinary create still reaches the catalog with the id it built.
func TestCreateDeviceDefinitionSendsAMintableIDToTheCatalog(t *testing.T) {
	h, mock, catalog := createHandler(t)
	expectDeviceTypeRead(mock)
	catalog.EXPECT().Create(gomock.Any(), "Toyota", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, dd coremodels.DeviceDefinitionTablelandModel) (*coremodels.Template, error) {
			assert.Equal(t, "toyota_camry_2026", dd.ID)
			return &coremodels.Template{ID: dd.ID}, nil
		})

	res, err := h.Handle(context.Background(), &CreateDeviceDefinitionCommand{
		Make:  "Toyota",
		Model: "Camry",
		Year:  2026,
	})

	require.NoError(t, err)
	created, ok := res.(CreateDeviceDefinitionCommandResult)
	require.True(t, ok)
	assert.Equal(t, "toyota_camry_2026", created.ID)
}
