package gateways

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/device-definitions-api/datgroup"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/hooklift/gowsdl/soap"
)

//go:generate mockgen -source datgroup_api_service.go -destination mocks/datgroup_api_service_mock.go -package mocks
type DATGroupAPIService interface {
	GetVIN(ctx context.Context, vin, country string) (*datgroup.GetVehicleIdentificationByVinResponse, error)
}

type datGroupAPIService struct {
	Settings *config.Settings
	log      *zerolog.Logger
}

func NewDATGroupAPIService(settings *config.Settings, logger *zerolog.Logger) DATGroupAPIService {
	return &datGroupAPIService{
		Settings: settings,
		log:      logger,
	}
}

func (ai *datGroupAPIService) GetVIN(ctx context.Context, vin, country string) (*datgroup.GetVehicleIdentificationByVinResponse, error) {
	tokenResponse, err := ai.getToken(ctx)
	if err != nil {
		return nil, err
	}

	soapClient := &soap.Client{}
	soapClient.SetHeaders(
		fmt.Sprintf("DAT-AuthorizationToken: %s", *tokenResponse),
	)

	client := datgroup.NewVehicleIdentificationService(soapClient)
	response, err := client.GetVehicleIdentificationByVinContext(ctx, &datgroup.GetVehicleIdentificationByVin{
		Request: &datgroup.VinSelectionRequest{
			Vin: vin,
			AbstractSelectionRequest: &datgroup.AbstractSelectionRequest{
				Locale: &datgroup.Locale{
					Country:             "US",
					DatCountryIndicator: country,
					Language:            "us",
				},
			},
		}})

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (ai *datGroupAPIService) getToken(ctx context.Context) (*string, error) {
	soapClient := &soap.Client{}
	client := datgroup.NewMyClaimExternalAuthenticationService(soapClient)

	ai.log.Info().Msgf("Client %s", client)

	response, err := client.GenerateTokenContext(ctx, &datgroup.GenerateToken{
		Request: &datgroup.GenerateTokenRequest{
			CustomerLogin:             ai.Settings.DatGroupCustomerLogin,
			CustomerNumber:            ai.Settings.DatGroupCustomerNumber,
			CustomerPassword:          ai.Settings.DatGroupCustomerPassword,
			InterfacePartnerNumber:    ai.Settings.DatGroupInterfacePartnerNumber,
			InterfacePartnerSignature: ai.Settings.DatGroupInterfacePartnerSignature,
		},
	})

	ai.log.Info().Msgf("Response %s", response)
	ai.log.Info().Msgf("Error %s", err)

	if err != nil {
		return nil, err
	}

	return &response.Token, nil
}
