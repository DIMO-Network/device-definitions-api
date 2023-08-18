package services

import (
	"context"
	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"testing"
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

	ptSvc, err := NewPowerTrainTypeService(pdb.DBS, "test_powertrain_type_rule.yaml", logger)
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
		// todo drivly
		// todo vincario
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := ptSvc.ResolvePowerTrainType(ctx, tt.args.makeSlug, tt.args.modelSlug, tt.args.definitionID, tt.args.drivlyData, tt.args.vincarioData)
			require.NoError(t, err)

			assert.Equalf(t, tt.want, got, "ResolvePowerTrainType( %v, %v, %v, %v, %v)", tt.args.makeSlug, tt.args.modelSlug, tt.args.definitionID, tt.args.drivlyData, tt.args.vincarioData)
		})
	}
}
