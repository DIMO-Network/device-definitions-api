package services

import (
	"context"
	"testing"

	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"go.uber.org/mock/gomock"

	"github.com/volatiletech/sqlboiler/v4/boil"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
)

func Test_powerTrainTypeService_ResolvePowerTrainType(t *testing.T) {
	// database
	const (
		dbName               = "device_definitions_api"
		migrationsDirRelPath = "../../infrastructure/db/migrations"
	)
	ctx := context.Background()
	pdb, container := dbtesthelper.StartContainerDatabase(ctx, dbName, t, migrationsDirRelPath)
	// rule data - just use production one
	logger := dbtesthelper.Logger()
	pdb.WaitForDB(*logger)
	defer container.Terminate(ctx) //nolint

	ctrl := gomock.NewController(t)
	onChainSvc := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	defer ctrl.Finish()

	// used for test case where get powertrain from dd
	dm := dbtesthelper.SetupCreateMake(t, "Ford", pdb)
	ddWithPt := dbtesthelper.SetupCreateDeviceDefinition(t, dm, "super special", 2022, pdb)
	ddWithPt.Metadata = null.JSONFrom([]byte(`{"vehicle_info": {"powertrain_type": "BEV"}}`))
	_, err := ddWithPt.Update(ctx, pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)

	ptSvc, err := NewPowerTrainTypeService(pdb.DBS, "../../../powertrain_type_rule.yaml", logger, onChainSvc)
	require.NoError(t, err)

	type args struct {
		makeSlug     string
		modelSlug    string
		definitionID *string
		drivlyData   null.JSON
		vincarioData null.JSON
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Tesla EV - from rules",
			args: args{
				makeSlug:  "tesla",
				modelSlug: "model-x",
			},
			want: "BEV",
		},
		{
			name: "Toyota rav4 HEV - from rules",
			args: args{
				makeSlug:  "Toyota",
				modelSlug: "rav4-hybrid",
			},
			want: "HEV",
		},
		{
			name: "Random hybrid - inferred from name",
			args: args{
				makeSlug:  "mitsubishi",
				modelSlug: "outlander-hybrid",
			},
			want: "HEV",
		},
		{
			name: "Random plugin - inferred from name",
			args: args{
				makeSlug:  "mitsubishi",
				modelSlug: "outlander-plug-in-hybrid",
			},
			want: "PHEV",
		},
		{
			name: "Inferred from Drivly fuel",
			args: args{
				makeSlug:   "mitsubishi",
				modelSlug:  "cool-car",
				drivlyData: null.JSONFrom([]byte(`{ "fuel": "Hybrid" }`)),
			},
			want: "HEV",
		},
		{
			name: "Inferred from Drivly fuel - ICE",
			args: args{
				makeSlug:   "mitsubishi",
				modelSlug:  "cool-car",
				drivlyData: null.JSONFrom([]byte(`{ "fuel": "Gasoline" }`)),
			},
			want: "ICE",
		},
		{
			name: "Inferred from vincario fuel - BEV",
			args: args{
				makeSlug:     "mitsubishi",
				modelSlug:    "cool-car",
				vincarioData: null.JSONFrom([]byte(`{ "FuelType": "Electric" }`)),
			},
			want: "BEV",
		},
		{
			name: "device definition already has powertrain - BEV",
			args: args{
				definitionID: &ddWithPt.ID,
			},
			want: "BEV",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := ptSvc.ResolvePowerTrainType(ctx, tt.args.makeSlug, tt.args.modelSlug, tt.args.definitionID, tt.args.drivlyData, tt.args.vincarioData)
			require.NoError(t, err)

			assert.Equalf(t, tt.want, got, "ResolvePowerTrainType( %v, %v, %v, %v, %v)", tt.args.makeSlug, tt.args.modelSlug, tt.args.definitionID, tt.args.drivlyData, tt.args.vincarioData)
		})
	}
}
