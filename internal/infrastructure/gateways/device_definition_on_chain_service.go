package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
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
	client   *ethclient.Client
	sender   sender.Sender
	chainID  *big.Int
}

func NewDeviceDefinitionOnChainService(settings *config.Settings, logger *zerolog.Logger, client *ethclient.Client, chainID *big.Int, sender sender.Sender) DeviceDefinitionOnChainService {
	return &deviceDefinitionOnChainService{
		Settings: settings,
		Logger:   logger,
		client:   client,
		chainID:  chainID,
		sender:   sender,
	}
}

func (e *deviceDefinitionOnChainService) CreateOrUpdate(ctx context.Context, manufacturerID types.NullDecimal, dd models.DeviceDefinition) (*string, error) {

	if !e.Settings.EthereumSendTransaction {
		return nil, nil
	}

	if manufacturerID.IsZero() {
		return nil, fmt.Errorf("manufacturerID has not value")
	}

	contractAddress := common.HexToAddress(e.Settings.EthereumRegistryAddress)
	fromAddress := e.sender.Address()

	nonce, err := e.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed get PendingNonceAt: %w", err)
	}

	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get SuggestGasPrice: %w", err)
	}

	auth, err := NewKeyedTransactorWithChainID(ctx, e.sender, e.chainID)
	//auth.Value = big.NewInt(0)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice
	auth.From = fromAddress

	manufacturerId := manufacturerID.Big.Int(new(big.Int))

	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	// Validate if manufacturer exists
	_, err = queryInstance.GetManufacturerNameById(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerId)
	if err != nil {
		e.Logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed get GetManufacturerNameById: %w", err)
	}

	instance, err := contracts.NewRegistryTransactor(contractAddress, e.client)
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

func NewKeyedTransactorWithChainID(context context.Context, send sender.Sender, chainID *big.Int) (*bind.TransactOpts, error) {
	signer := eth_types.LatestSignerForChainID(chainID)
	return &bind.TransactOpts{
		From: send.Address(),
		Signer: func(address common.Address, tx *eth_types.Transaction) (*eth_types.Transaction, error) {
			signature, err := send.Sign(context, signer.Hash(tx))
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
		Context: context,
	}, nil
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
