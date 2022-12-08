package queries

import (
	"context"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"testing"
)

type DecodeVINQueryHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl             *gomock.Controller
	pdb              db.Store
	container        testcontainers.Container
	ctx              context.Context
	mockDrivlyApiSvc *mock_gateways.MockDrivlyAPIService

	queryHandler DecodeVINQueryHandler
}

func TestDecodeVINQueryHandler(t *testing.T) {
	suite.Run(t, new(DecodeVINQueryHandlerSuite))
}

func (s *DecodeVINQueryHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()

	s.mockDrivlyApiSvc = mock_gateways.NewMockDrivlyAPIService(s.ctrl)
	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)
	// todo inject mock drivly api svc
	s.queryHandler = NewDecodeVINQueryHandler(s.pdb.DBS, &config.Settings{}, dbtesthelper.Logger())
}

func (s *DecodeVINQueryHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}
