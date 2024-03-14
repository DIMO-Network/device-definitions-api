package gateways

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"math/big"
)

//go:generate mockgen -source device_definition_on_chain_service.go -destination mocks/device_definition_on_chain_service_mock.go -package mocks
type DeviceDefinitionOnChainService interface {
	CreateOrUpdate(ctx context.Context, manufacturerID types.NullDecimal, dd models.DeviceDefinition) (*string, error)
}

type deviceDefinitionOnChainService struct {
	Settings *config.Settings
	Logger   *zerolog.Logger
}

func NewDeviceDefinitionOnChainService(settings *config.Settings, logger *zerolog.Logger) DeviceDefinitionOnChainService {
	return &deviceDefinitionOnChainService{
		Settings: settings,
		Logger:   logger,
	}
}

func (e *deviceDefinitionOnChainService) CreateOrUpdate(ctx context.Context, manufacturerID types.NullDecimal, dd models.DeviceDefinition) (*string, error) {
	if len(e.Settings.EthereumNetwork) == 0 {
		return nil, nil
	}

	if manufacturerID.IsZero() {
		return nil, fmt.Errorf("manufacturerID has not value")
	}

	client, err := ethclient.Dial(e.Settings.EthereumNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed connect to Etherum Network: %w", err)
	}

	contractAddress := common.HexToAddress(e.Settings.EthereumRegistryAddress)
	privateKey, err := crypto.HexToECDSA(e.Settings.EthereumWalletPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("privateKey: %w", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get ChainID: %w", err)
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed get PendingNonceAt: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get SuggestGasPrice: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	//auth.Value = big.NewInt(0)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice
	auth.From = fromAddress

	manufacturerId := manufacturerID.Big.Int(new(big.Int))

	queryInstance, err := contracts.NewRegistry(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	// Validate if manufacturer exists
	_, err = queryInstance.GetManufacturerNameById(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerId)
	if err != nil {
		e.Logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed get GetManufacturerNameById: %w", err)
	}

	instance, err := contracts.NewRegistryTransactor(contractAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistryTransactor: %w", err)
	}

	deviceInputs := contracts.DeviceDefinitionInput{
		Id:       fmt.Sprintf("%s_%d", dd.ModelSlug, dd.Year),
		Model:    dd.ModelSlug,
		Year:     big.NewInt(int64(dd.Year)),
		Metadata: "",
		Ksuid:    dd.ID,
	}

	if dd.Metadata.Valid {
		attributes := GetDeviceAttributesTyped(dd.Metadata, "vehicle_info")
		type deviceAttributes struct {
			DeviceAttributes []DeviceTypeAttribute `json:"device_attributes"`
		}
		deviceAttributesStruct := deviceAttributes{
			DeviceAttributes: attributes,
		}
		jsonData, _ := json.Marshal(deviceAttributesStruct)
		deviceInputs.Metadata = string(jsonData)
	}

	tx, err := instance.InsertDeviceDefinition(auth, manufacturerId, deviceInputs)

	if err != nil {
		e.Logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed insert InsertDeviceDefinitionBatch: %w", err)
	}

	trx := tx.Hash().Hex()

	return &trx, nil
}

func GetDeviceAttributesTyped(metadata null.JSON, key string) []DeviceTypeAttribute {
	var respAttrs []DeviceTypeAttribute
	var ai map[string]any
	if err := metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[key]; ok && a != nil {
				attributes := ai[key].(map[string]any)
				for key, value := range attributes {
					respAttrs = append(respAttrs, DeviceTypeAttribute{
						Name:  key,
						Value: fmt.Sprint(value),
					})
				}
			}
		}
	}
	return respAttrs
}

type DeviceTypeAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
