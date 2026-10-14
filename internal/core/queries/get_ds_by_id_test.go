package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

// A catalog outage (a 500, a timeout, a decode failure) must not be
// reclassified as "this vehicle does not exist": that reclassification is
// what would let a caller respond by creating a duplicate definition mid-
// outage, the worst possible moment for it.
type GetDeviceStyleByIDQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl           *gomock.Controller
	pdb            db.Store
	container      testcontainers.Container
	ctx            context.Context
	mockCatalogSvc *mock_gateways.MockDeviceDefinitionCatalogService

	queryHandler GetDeviceStyleByIDQueryHandler
}

func TestGetDeviceStyleByIDQueryHandler(t *testing.T) {
	suite.Run(t, new(GetDeviceStyleByIDQueryHandlerSuite))
}

func (s *GetDeviceStyleByIDQueryHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.mockCatalogSvc = mock_gateways.NewMockDeviceDefinitionCatalogService(s.ctrl)

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.queryHandler = NewGetDeviceStyleByIDQueryHandler(s.pdb.DBS, s.mockCatalogSvc)
}

func (s *GetDeviceStyleByIDQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *GetDeviceStyleByIDQueryHandlerSuite) TearDownSuite() {
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatal(err)
	}
}

// A genuine catalog 404 (ErrTemplateNotFound) is the one case that should
// surface as NotFoundError: the vehicle's parent template really is absent.
func (s *GetDeviceStyleByIDQueryHandlerSuite) TestHandle_CatalogTemplateNotFound_ReturnsNotFoundError() {
	ds := dbtesthelper.SetupCreateStyle(s.T(), "toyota_camry_2020", "LE", "drivly", "", s.pdb)

	s.mockCatalogSvc.EXPECT().GetTemplateByID(gomock.Any(), ds.DefinitionID).Return(
		nil, nil, errors.Wrapf(gateways.ErrTemplateNotFound, "template %s", ds.DefinitionID))

	_, err := s.queryHandler.Handle(s.ctx, &GetDeviceStyleByIDQuery{DeviceStyleID: ds.ID})

	require.Error(s.T(), err)
	var notFound *exceptions.NotFoundError
	assert.ErrorAs(s.T(), err, &notFound)
}

// A catalog outage -- a 500, a timeout, a decode failure -- is a distinct,
// non-sentinel error from GetTemplateByID. It must come back as an internal
// error, never a NotFoundError: conflating the two is what makes a catalog
// outage look like a missing vehicle.
func (s *GetDeviceStyleByIDQueryHandlerSuite) TestHandle_CatalogOutage_ReturnsInternalErrorNotNotFound() {
	ds := dbtesthelper.SetupCreateStyle(s.T(), "toyota_camry_2020", "LE", "drivly", "", s.pdb)

	s.mockCatalogSvc.EXPECT().GetTemplateByID(gomock.Any(), ds.DefinitionID).Return(
		nil, nil, fmt.Errorf("catalog returned 500 for template %s", ds.DefinitionID))

	_, err := s.queryHandler.Handle(s.ctx, &GetDeviceStyleByIDQuery{DeviceStyleID: ds.ID})

	require.Error(s.T(), err)
	var notFound *exceptions.NotFoundError
	assert.False(s.T(), errors.As(err, &notFound), "a catalog outage must not surface as NotFoundError")
	var internal *exceptions.InternalError
	assert.ErrorAs(s.T(), err, &internal)
}
