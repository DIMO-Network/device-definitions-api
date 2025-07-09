package services

import (
	"testing"

	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"go.uber.org/mock/gomock"

	dbtesthelper "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/dbtest"
	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_powerTrainTypeService_ResolvePowerTrainType(t *testing.T) {

	// rule data - just use production one
	logger := dbtesthelper.Logger()

	ctrl := gomock.NewController(t)
	onChainSvc := mock_gateways.NewMockDeviceDefinitionOnChainService(ctrl)
	defer ctrl.Finish()

	ptSvc, err := NewPowerTrainTypeService("../../../powertrain_type_rule.yaml", logger, onChainSvc)
	require.NoError(t, err)

	type args struct {
		makeSlug     string
		modelSlug    string
		definitionID *string
		drivlyData   null.JSON
		vincarioData null.JSON
	}
	tests := []struct {
		name   string
		args   args
		want   string
		before func()
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}

			got, err := ptSvc.ResolvePowerTrainType(tt.args.makeSlug, tt.args.modelSlug, tt.args.drivlyData, tt.args.vincarioData)
			assert.NoError(t, err)

			assert.Equalf(t, tt.want, got, "ResolvePowerTrainType( %v, %v, %v, %v, %v)", tt.args.makeSlug, tt.args.modelSlug, tt.args.definitionID, tt.args.drivlyData, tt.args.vincarioData)
		})
	}
}
