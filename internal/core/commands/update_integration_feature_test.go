package commands

import (
	"context"
	_ "embed"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/DIMO-Network/shared/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/mock/gomock"
)

type UpdateIntegrationFeatureCommandHandlerSuite struct {
	suite.Suite
	*require.Assertions

	ctrl      *gomock.Controller
	pdb       db.Store
	container testcontainers.Container
	ctx       context.Context

	commandHandler UpdateIntegrationFeatureCommandHandler
}

func TestUpdateIntegrationFeatureCommandHandler(t *testing.T) {
	suite.Run(t, new(UpdateIntegrationFeatureCommandHandlerSuite))
}

func (s *UpdateIntegrationFeatureCommandHandlerSuite) SetupTest() {

	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)

	s.ctx = context.Background()
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())

	s.pdb, s.container = dbtesthelper.StartContainerDatabase(s.ctx, dbName, s.T(), migrationsDirRelPath)

	s.commandHandler = NewUpdateIntegrationFeatureCommandHandler(s.pdb.DBS)
}

func (s *UpdateIntegrationFeatureCommandHandlerSuite) TearDownTest() {
	dbtesthelper.TruncateTables(s.pdb.DBS().Writer.DB, s.T())
	s.ctrl.Finish()
}

func (s *UpdateIntegrationFeatureCommandHandlerSuite) TestUpdateIntegrationFeatureCommand_Success() {
	ctx := context.Background()

	displayName := "property-display-name"
	css := "css-data"
	elasticProperty := "elastic-property"

	feature := dbtesthelper.SetupIntegrationFeature(s.T(), s.pdb)

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateIntegrationFeatureCommand{
		ID:              feature.FeatureKey,
		DisplayName:     displayName,
		CSSIcon:         css,
		FeatureWeight:   1,
		ElasticProperty: elasticProperty,
	})
	result := commandResult.(UpdateIntegrationFeatureResult)

	s.NoError(err)
	assert.Equal(s.T(), result.ID, feature.FeatureKey)

	featureUpdate, _ := models.IntegrationFeatures(models.IntegrationFeatureWhere.FeatureKey.EQ(feature.FeatureKey)).One(ctx, s.pdb.DBS().Writer)

	assert.Equal(s.T(), featureUpdate.DisplayName, displayName)
	assert.Equal(s.T(), featureUpdate.ElasticProperty, elasticProperty)
	assert.Equal(s.T(), featureUpdate.CSSIcon.String, css)

}

func (s *UpdateIntegrationFeatureCommandHandlerSuite) TestUpdateIntegrationFeatureCommand_Exception() {
	ctx := context.Background()

	commandResult, err := s.commandHandler.Handle(ctx, &UpdateIntegrationFeatureCommand{
		ID: "dd.null",
	})

	s.Nil(commandResult)
	s.Error(err)
}
