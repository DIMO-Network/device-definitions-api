package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/mock/gomock"
)

func TestGetDeviceStyleByIDQueryHandler_Handle(t *testing.T) {
	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)
	ctx := context.Background()
	pdb, container := dbtesthelper.StartContainerDatabase(ctx, dbName, t, migrationsDirRelPath)
	defer container.Terminate(ctx)                      //nolint
	dbtesthelper.TruncateTables(pdb.DBS().Writer.DB, t) // clear setup data for integration features
	mockCtrl := gomock.NewController(t)
	ddCacheSvc := mocks.NewMockDeviceDefinitionCacheService(mockCtrl)

	dm := dbtesthelper.SetupCreateMake(t, "Ford", pdb)
	dd := dbtesthelper.SetupCreateDeviceDefinition(t, dm, "Escape", 2022, pdb)
	dsHybridName := dbtesthelper.SetupCreateStyle(t, dd.NameSlug, "4dr Hatchback (1.8L 4cyl gas/electric hybrid CVT)", "drivly", "1", pdb)
	dsNormal := dbtesthelper.SetupCreateStyle(t, dd.NameSlug, "2.0 vvti", "drivly", "2", pdb)
	dsWithPowertrain := dbtesthelper.SetupCreateStyle(t, dd.NameSlug, "super energiii", "drivly", "3", pdb)
	dsWithPowertrain.Metadata = null.JSONFrom([]byte(fmt.Sprintf(`{"%s": "BEV"}`, common.PowerTrainType)))
	_, err := dsWithPowertrain.Update(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	ddWithPt := dbtesthelper.SetupCreateDeviceDefinition(t, dm, "Focus", 2022, pdb)
	ddWithPt.Metadata = null.JSONFrom([]byte(`{"vehicle_info": {"powertrain_type": "ICE"}}`))
	_, err = ddWithPt.Update(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	dsHybridOverride := dbtesthelper.SetupCreateStyle(t, ddWithPt.NameSlug, "4dr Hatchback (1.8L 4cyl gas/electric hybrid CVT)", "drivly", "1", pdb)

	rp, err := common.BuildFromDeviceDefinitionToQueryResult(dd)
	require.NoError(t, err)
	ddCacheSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), dd.NameSlug).Times(3).Return(rp, nil)       // expect 3 times call since used in 3 tests
	ddCacheSvc.EXPECT().GetDeviceDefinitionByID(gomock.Any(), ddWithPt.NameSlug).Times(1).Return(rp, nil) // expect 1 times call since used in 1 tests

	tests := []struct {
		name           string
		query          *GetDeviceStyleByIDQuery
		wantPowertrain string
	}{
		{name: "powertrain overriden from device definitions",
			query:          &GetDeviceStyleByIDQuery{DeviceStyleID: dsHybridOverride.ID},
			wantPowertrain: coremodels.HEV.String(),
		},
		{name: "powertrain inherited from device definitions",
			query:          &GetDeviceStyleByIDQuery{DeviceStyleID: dsNormal.ID},
			wantPowertrain: coremodels.ICE.String(),
		},
		{name: "powertrain from style naming logic",
			query:          &GetDeviceStyleByIDQuery{DeviceStyleID: dsHybridName.ID},
			wantPowertrain: coremodels.HEV.String(),
		},
		{name: "powertrain from style metadata",
			query:          &GetDeviceStyleByIDQuery{DeviceStyleID: dsWithPowertrain.ID},
			wantPowertrain: coremodels.BEV.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := GetDeviceStyleByIDQueryHandler{
				DBS:     pdb.DBS,
				DDCache: ddCacheSvc,
			}
			got, err := ch.Handle(ctx, tt.query)
			require.NoError(t, err)

			result := got.(coremodels.GetDeviceStyleQueryResult)
			if result.ID == dsHybridOverride.ID {
				assert.Equal(t, ddWithPt.NameSlug, result.DefinitionID)
			} else {
				assert.Equal(t, dd.NameSlug, result.DefinitionID)
			}
			pt := ""
			for _, attribute := range result.DeviceDefinition.DeviceAttributes {
				if attribute.Name == "powertrain_type" {
					pt = attribute.Value
				}
			}
			assert.Equal(t, tt.wantPowertrain, pt)

		})
	}
}
