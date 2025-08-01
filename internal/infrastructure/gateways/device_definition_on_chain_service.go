//nolint:tagliatelle
package gateways

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/contracts"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/sender"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/DIMO-Network/shared/pkg/db"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

//go:generate mockgen -source device_definition_on_chain_service.go -destination mocks/device_definition_on_chain_service_mock.go -package mocks
type DeviceDefinitionOnChainService interface {
	GetManufacturer(manufacturerSlug string) (*coremodels.Manufacturer, error)
	GetManufacturerNameByID(ctx context.Context, manufacturerID *big.Int) (string, error)
	// GetDeviceDefinitionByID get DD from tableland by slug ID and specifying the manufacturer for the table to lookup in
	GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	// GetDefinitionByID get DD from tableland by slug ID, automatically figures out table by oem portion of slug. returns the manufacturer token id too
	GetDefinitionByID(ctx context.Context, ID string) (*coremodels.DeviceDefinitionTablelandModel, *big.Int, error)
	GetDefinitionTableland(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error)
	GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]coremodels.DeviceDefinitionTablelandModel, error)
	Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error)
	Update(ctx context.Context, manufacturerName string, input contracts.DeviceDefinitionUpdateInput) (*string, error)
	Delete(ctx context.Context, manufacturerName, id string) (*string, error)
	QueryDefinitionsCustom(ctx context.Context, manufacturerID int, whereClause string, pageIndex int) ([]coremodels.DeviceDefinitionTablelandModel, error)
}

type deviceDefinitionOnChainService struct {
	settings    *config.Settings
	logger      *zerolog.Logger
	client      *ethclient.Client
	sender      sender.Sender
	chainID     *big.Int
	identityAPI IdentityAPI
	inmemCache  *cache.Cache
	dbs         func() *db.ReaderWriter
}

func NewDeviceDefinitionOnChainService(settings *config.Settings, logger *zerolog.Logger, client *ethclient.Client,
	chainID *big.Int, sender sender.Sender, dbs func() *db.ReaderWriter) DeviceDefinitionOnChainService {
	return &deviceDefinitionOnChainService{
		settings:    settings,
		logger:      logger,
		client:      client,
		chainID:     chainID,
		sender:      sender,
		identityAPI: NewIdentityAPIService(logger, settings),
		inmemCache:  cache.New(128*time.Hour, 1*time.Hour),
		dbs:         dbs,
	}
}

// GetDeviceDefinitionByID gets dd from tableland with a select statement, returning a db model object
func (e *deviceDefinitionOnChainService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error) {
	tablelandDD, err := e.GetDefinitionTableland(ctx, manufacturerID, ID)
	if err != nil {
		return nil, err
	}

	return tablelandDD, nil
}

func (e *deviceDefinitionOnChainService) getTablelandTableName(ctx context.Context, manufacturerID *big.Int) (string, error) {
	cacheKey := "manufacturer_" + manufacturerID.String()
	value, found := e.inmemCache.Get(cacheKey)
	if found {
		return value.(string), nil
	}

	contractAddress := e.settings.EthereumRegistryAddress
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return "", fmt.Errorf("failed to establish NewRegistry: %w", err)
	}

	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerID)

	if err != nil {
		return "", errors.Wrapf(err, "failed to getTablelandTableName for %d", manufacturerID.Uint64())
	}
	e.inmemCache.Set(cacheKey, tableName, time.Hour*300)

	return tableName, nil
}

func (e *deviceDefinitionOnChainService) GetManufacturerNameByID(ctx context.Context, manufacturerID *big.Int) (string, error) {
	contractAddress := e.settings.EthereumRegistryAddress
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return "", fmt.Errorf("failed to establish NewRegistry: %w", err)
	}
	return queryInstance.GetManufacturerNameById(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerID)
}

// GetDefinitionByID returns the tableland on chain DD model and the manufacturer token id
func (e *deviceDefinitionOnChainService) GetDefinitionByID(ctx context.Context, ID string) (*coremodels.DeviceDefinitionTablelandModel, *big.Int, error) {
	split := strings.Split(ID, "_")
	if len(split) != 3 {
		return nil, nil, fmt.Errorf("get dd by slug - invalid slug: %s", ID)
	}
	manufacturerSlug := split[0]
	// call out to identity-api w/ caching
	manufacturer, err := e.GetManufacturer(manufacturerSlug)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed get DeviceMake: %s", manufacturerSlug)
	}
	manufacturerID := big.NewInt(int64(manufacturer.TokenID))
	tblDD, err := e.GetDefinitionTableland(ctx, manufacturerID, ID)
	return tblDD, manufacturerID, err
}

func (e *deviceDefinitionOnChainService) GetManufacturer(manufacturerSlug string) (*coremodels.Manufacturer, error) {
	value, found := e.inmemCache.Get(manufacturerSlug)
	if found {
		return value.(*coremodels.Manufacturer), nil
	}
	manufacturer, err := e.identityAPI.GetManufacturer(manufacturerSlug)
	if err != nil {
		return nil, err
	}
	e.inmemCache.Set(manufacturerSlug, manufacturer, time.Hour*300)
	return manufacturer, nil
}

// GetDefinitionTableland gets dd from tableland with a select statement and returns tbl object
func (e *deviceDefinitionOnChainService) GetDefinitionTableland(ctx context.Context, manufacturerID *big.Int, ID string) (*coremodels.DeviceDefinitionTablelandModel, error) {
	if manufacturerID == nil || manufacturerID.Uint64() == 0 {
		return nil, fmt.Errorf("GetDefinitionTableland - manufacturerID cannot be 0")
	}

	tableName, err := e.getTablelandTableName(ctx, manufacturerID)
	if err != nil {
		e.logger.Info().Msgf("%s", err)
		return nil, err
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", tableName, ID)
	queryParams := map[string]string{
		"statement": statement,
	}

	var modelTableland []coremodels.DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, errors.Wrapf(err, "failed to query tableland, manufacturer: %d", manufacturerID.Int64())
	}
	if len(modelTableland) == 0 {
		return nil, nil
	}

	return &modelTableland[0], nil
}

func (e *deviceDefinitionOnChainService) GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]coremodels.DeviceDefinitionTablelandModel, error) {
	if manufacturerID.IsZero() {
		return nil, fmt.Errorf("manufacturerID cannot be 0")
	}
	bigManufID := manufacturerID.Int(new(big.Int))
	tableName, err := e.getTablelandTableName(ctx, bigManufID)
	if err != nil {
		return nil, err
	}

	var conditions []string
	if year > 1980 && year < 2999 {
		conditions = append(conditions, fmt.Sprintf("year = %d", year))
	}
	if len(model) > 0 {
		conditions = append(conditions, fmt.Sprintf("model = '%s'", model))
	}
	if len(ID) > 0 {
		conditions = append(conditions, fmt.Sprintf("id = '%s'", ID))
	}

	whereClause := strings.Join(conditions, " AND ")
	if whereClause != "" {
		whereClause = " WHERE " + whereClause
	}

	statement := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d OFFSET %d", tableName, whereClause, pageSize, pageIndex)

	queryParams := map[string]string{
		"statement": statement,
	}

	var modelTableland []coremodels.DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, err
	}

	return modelTableland, nil
}

// QueryDefinitionsCustom queries tableland definitions oem table based on manuf ID. Always page size of 500, but you can alter the page index
func (e *deviceDefinitionOnChainService) QueryDefinitionsCustom(ctx context.Context, manufacturerID int, whereClause string, pageIndex int) ([]coremodels.DeviceDefinitionTablelandModel, error) {
	if manufacturerID == 0 {
		return nil, fmt.Errorf("manufacturerID cannot be 0")
	}

	bigManufID := big.NewInt(int64(manufacturerID))
	tableName, err := e.getTablelandTableName(ctx, bigManufID)
	if err != nil {
		return nil, err
	}
	if tableName == "" {
		return nil, fmt.Errorf("tableName cannot be empty for manufacturer token id %d", manufacturerID)
	}

	statement := fmt.Sprintf("SELECT * FROM %s %s LIMIT %d OFFSET %d", tableName, whereClause, 500, pageIndex)

	queryParams := map[string]string{
		"statement": statement,
	}

	var modelTableland []coremodels.DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, err
	}

	return modelTableland, nil
}

func (e *deviceDefinitionOnChainService) QueryTableland(queryParams map[string]string, result interface{}) error {
	fullURL, err := url.Parse(e.settings.TablelandAPIGateway)
	if err != nil {
		return err
	}

	fullURL.Path = path.Join(fullURL.Path, "api/v1/query")

	if queryParams != nil {
		values := fullURL.Query()
		for key, value := range queryParams {
			values.Set(key, value)
		}
		fullURL.RawQuery = values.Encode()
	}

	resp, err := http.Get(fullURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return errors.Wrapf(err, "resp body: %s. url: %s", string(body), fullURL.String())
	}
	return nil
}

const (
	TablelandRequests = "Tableland_All_Request"
	TablelandFindByID = "Tableland_FindByID_Request"
	TablelandCreated  = "Tableland_Created_Request"
	TablelandUpdated  = "Tableland_Updated_Request"
	TablelandDeleted  = "Tableland_Deleted_Request"
	TablelandExists   = "Tableland_Exists_Request"
	TablelandErrors   = "Tableland_Error_Request"
)

// Create does a create for tableland, on-chain operation - checks if already exists, inserts transaction in db. returns the onchain transaction
func (e *deviceDefinitionOnChainService) Create(ctx context.Context, manufacturerName string, dd coremodels.DeviceDefinitionTablelandModel) (*string, error) {

	metrics.Success.With(prometheus.Labels{"method": TablelandRequests}).Inc()
	e.logger.Info().Msgf("OnChain Start Create for device definition %s. EthereumSendTransaction %t. payload: %+v", dd.ID, e.settings.EthereumSendTransaction, dd)

	if !e.settings.EthereumSendTransaction {
		return nil, nil
	}

	contractAddress := e.settings.EthereumRegistryAddress
	fromAddress := e.sender.Address()

	nonce, err := e.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get PendingNonceAt: %w", err)
	}

	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get SuggestGasPrice: %w", err)
	}

	bump := big.NewInt(20)
	bumpedPrice := getGasPrice(gasPrice, bump)

	e.logger.Info().Msgf("bumped gas price: %d", bumpedPrice)

	auth, err := NewKeyedTransactorWithChainID(ctx, e.sender, e.chainID)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get NewKeyedTransactorWithChainID: %w", err)
	}
	//auth.Value = big.NewInt(0)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = bumpedPrice
	auth.From = fromAddress

	// call identity instead of the chain since more reliable
	manufacturer, err := e.GetManufacturer(stringutils.SlugString(manufacturerName))
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get manufacturer from identity: %w", err)
	}
	if manufacturer == nil {
		return nil, fmt.Errorf("manufacturer not found in identity")
	}
	e.logger.Info().Msgf("manufacturer to create DD with: %+v", manufacturer)

	instance, err := contracts.NewRegistryTransactor(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistryTransactor: %w", err)
	}

	if dd.DeviceType == "" {
		return nil, fmt.Errorf("dd DeviceTypeId is required")
	}
	if dd.ImageURI == "" {
		dd.ImageURI = GetDefaultImageURL(ctx, dd.ID, e.dbs().Reader.DB)
	}

	deviceInputs := contracts.DeviceDefinitionInput{
		Id:         dd.ID,
		Model:      dd.Model,
		Year:       big.NewInt(int64(dd.Year)),
		Ksuid:      dd.ID,
		DeviceType: dd.DeviceType,
		ImageURI:   dd.ImageURI,
	}

	if dd.Metadata != nil {
		jsonData, _ := json.Marshal(dd.Metadata)
		deviceInputs.Metadata = string(jsonData)
	}

	bigManufID := big.NewInt(int64(manufacturer.TokenID))

	// check for duplicate create
	currentDeviceDefinition, err := e.GetDeviceDefinitionByID(ctx, bigManufID, deviceInputs.Id)
	if err != nil {
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		e.logger.Err(err).Msgf("error occurred get device definition %s from tableland when checking for existence.", deviceInputs.Id)
		return nil, err
	}
	if currentDeviceDefinition != nil {
		return nil, fmt.Errorf("cannot create device definition, already exists: %s", deviceInputs.Id)
	}

	tx, err := instance.InsertDeviceDefinition(auth, bigManufID, deviceInputs)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed insert InsertDeviceDefinition: %w", err)
	}

	metrics.Success.With(prometheus.Labels{"method": TablelandCreated}).Inc()

	trx := tx.Hash().Hex()
	e.logger.Info().Msgf("Executed InsertDeviceDefinition %s with Trx %s in ManufacturerID %s", deviceInputs.Id, trx, bigManufID)

	dbTrx := models.DefinitionTransaction{
		TransactionHash: trx,
		DefinitionID:    dd.ID,
		ManufacturerID:  bigManufID.Int64(),
	}
	err = dbTrx.Insert(ctx, e.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &trx, nil
}

// Update on-chain device definition, only has basic validation that some fields be present. Requires existing tableland record to exist to update
func (e *deviceDefinitionOnChainService) Update(ctx context.Context, manufacturerName string, input contracts.DeviceDefinitionUpdateInput) (*string, error) {
	// todo add retry logic to this call
	metrics.Success.With(prometheus.Labels{"method": TablelandRequests}).Inc()
	e.logger.Info().Msgf("OnChain Start Update for device definition %s. EthereumSendTransaction %t. payload: %+v", input.Id, e.settings.EthereumSendTransaction, input)

	if !e.settings.EthereumSendTransaction {
		return nil, nil
	}

	// validations
	if input.DeviceType == "" {
		return nil, fmt.Errorf("dd DeviceType is required")
	}
	if len(input.Metadata) > 4 {
		if !gjson.Get(input.Metadata, "device_attributes").Exists() {
			return nil, fmt.Errorf("device_attributes node is required in metadata if field is set: %s", input.Metadata)
		}
	}

	contractAddress := e.settings.EthereumRegistryAddress
	fromAddress := e.sender.Address()

	nonce, err := e.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get PendingNonceAt: %w", err)
	}

	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get SuggestGasPrice: %w", err)
	}

	bump := big.NewInt(20)
	bumpedPrice := getGasPrice(gasPrice, bump)

	e.logger.Info().Msgf("bumped gas price: %d", bumpedPrice)

	auth, err := NewKeyedTransactorWithChainID(ctx, e.sender, e.chainID)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get NewKeyedTransactorWithChainID: %w", err)
	}
	//auth.Value = big.NewInt(0)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = bumpedPrice
	auth.From = fromAddress

	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	// Validate if manufacturer exists
	bigManufID, err := queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerName)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get GetManufacturerIdByName => %s: %w", manufacturerName, err)
	}
	instance, err := contracts.NewRegistryTransactor(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistryTransactor: %w", err)
	}

	// check if any field changed
	e.logger.Info().Msgf("Validating if device definition %s with tokenID %s exists in tableland", input.Id, bigManufID)
	existingTblDD, err := e.GetDeviceDefinitionByID(ctx, bigManufID, input.Id)
	if existingTblDD != nil {
		e.logger.Info().Msgf("DD %s found.", existingTblDD.ID)
	}
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		e.logger.Err(err).Msgf("Error occurred get device definition %s from tableland.", input.Id)
		return nil, err
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandFindByID}).Inc()
	// change this up if want this method to do update and or create
	if existingTblDD == nil {
		return nil, fmt.Errorf("device definition %s not found in tableland to update", input.Id)
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandExists}).Inc()
	// todo - change GetDeviceDefinition above to return just the tableland object, and compare with our input for any changes.
	// if no changes just return nil trx hash. do not allow changing model or year since it changes the slug id - need to look into these cases better
	// as they have vehicle NFT implications.

	e.logger.Info().Msgf("Executing UpdateDeviceDefinition %s with manufacturer ID %s", input.Id, bigManufID)

	tx, err := instance.UpdateDeviceDefinition(auth, bigManufID, input)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", input.Id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed update UpdateDeviceDefinition: %w", err)
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandUpdated}).Inc()

	trx := tx.Hash().Hex()

	e.logger.Info().Msgf("Executed UpdateDeviceDefinition %s with Trx %s in ManufacturerID %s", input.Id, trx, bigManufID)
	dbTrx := models.DefinitionTransaction{
		TransactionHash: trx,
		DefinitionID:    existingTblDD.ID,
		ManufacturerID:  bigManufID.Int64(),
	}
	err = dbTrx.Insert(ctx, e.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &trx, nil
}

// Delete on-chain device definition by id. Requires existing tableland record to exist to delete
func (e *deviceDefinitionOnChainService) Delete(ctx context.Context, manufacturerName, id string) (*string, error) {

	metrics.Success.With(prometheus.Labels{"method": TablelandRequests}).Inc()
	e.logger.Info().Msgf("OnChain Start Delete for device definition %s. EthereumSendTransaction %t.", id, e.settings.EthereumSendTransaction)

	if !e.settings.EthereumSendTransaction {
		return nil, nil
	}

	// validations
	if len(id) == 0 {
		return nil, fmt.Errorf("id is required")
	}

	contractAddress := e.settings.EthereumRegistryAddress
	fromAddress := e.sender.Address()

	nonce, err := e.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get PendingNonceAt: %w", err)
	}

	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get SuggestGasPrice: %w", err)
	}

	bump := big.NewInt(20)
	bumpedPrice := getGasPrice(gasPrice, bump)

	e.logger.Info().Msgf("bumped gas price: %d", bumpedPrice)

	auth, err := NewKeyedTransactorWithChainID(ctx, e.sender, e.chainID)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get NewKeyedTransactorWithChainID: %w", err)
	}
	//auth.Value = big.NewInt(0)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasLimit = uint64(300000)
	auth.GasPrice = bumpedPrice
	auth.From = fromAddress

	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	// Validate if manufacturer exists
	bigManufID, err := queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerName)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get GetManufacturerIdByName => %s: %w", manufacturerName, err)
	}
	instance, err := contracts.NewRegistryTransactor(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistryTransactor: %w", err)
	}

	// check if any field changed
	existingTblDD, err := e.GetDeviceDefinitionByID(ctx, bigManufID, id)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		e.logger.Err(err).Msgf("Error occurred get device definition %s from tableland, manuf id: %d.", id, bigManufID.Int64())
		return nil, err
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandFindByID}).Inc()
	if existingTblDD == nil {
		return nil, fmt.Errorf("device definition %s not found in tableland to update", id)
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandExists}).Inc()
	// todo - change GetDeviceDefinition above to return just the tableland object, and compare with our input for any changes.
	// if no changes just return nil trx hash. do not allow changing model or year since it changes the slug id - need to look into these cases better
	// as they have vehicle NFT implications.

	e.logger.Info().Msgf("Executing DeleteDeviceDefinition %s with manufacturer ID %d", id, bigManufID.Int64())

	tx, err := instance.DeleteDeviceDefinition(auth, bigManufID, id)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", id)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed delete DeleteDeviceDefinition: %w", err)
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandDeleted}).Inc()

	trx := tx.Hash().Hex()

	e.logger.Info().Msgf("Executed DeleteDeviceDefinition %s with Trx %s in ManufacturerID %s", id, trx, bigManufID)
	dbTrx := models.DefinitionTransaction{
		TransactionHash: trx,
		DefinitionID:    existingTblDD.ID,
		ManufacturerID:  bigManufID.Int64(),
	}
	err = dbTrx.Insert(ctx, e.dbs().Writer, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &trx, nil
}

func validateAttributes(current, newAttrs []coremodels.DeviceTypeAttribute) ([]coremodels.DeviceTypeAttribute, []coremodels.DeviceTypeAttribute) {
	currentMap := attributesToMap(current)
	newMap := attributesToMap(newAttrs)

	var newOrModifiedAttributes []coremodels.DeviceTypeAttribute
	var removedAttributes []coremodels.DeviceTypeAttribute

	// Find new or changed attributes
	for name, newValue := range newMap {
		if currentValue, exists := currentMap[name]; !exists || currentValue != newValue {
			newOrModifiedAttributes = append(newOrModifiedAttributes, coremodels.DeviceTypeAttribute{
				Name:  name,
				Value: newValue,
			})
		}
	}

	// Find deleted attributes
	for name, currentValue := range currentMap {
		if _, exists := newMap[name]; !exists {
			removedAttributes = append(removedAttributes, coremodels.DeviceTypeAttribute{
				Name:  name,
				Value: currentValue,
			})
		}
	}

	return newOrModifiedAttributes, removedAttributes
}

func attributesToMap(attributes []coremodels.DeviceTypeAttribute) map[string]string {
	attrMap := make(map[string]string)
	for _, attr := range attributes {
		attrMap[attr.Name] = attr.Value
	}
	return attrMap
}

func getGasPrice(price *big.Int, bump *big.Int) *big.Int {
	// Calculating the bumped gas price
	bumpAmount := new(big.Int).Mul(price, bump)
	bumpAmount.Div(bumpAmount, big.NewInt(100))
	bumpedPrice := new(big.Int).Add(bumpAmount, price)

	return bumpedPrice
}

func NewKeyedTransactorWithChainID(context context.Context, send sender.Sender, chainID *big.Int) (*bind.TransactOpts, error) {
	signer := eth_types.LatestSignerForChainID(chainID)
	return &bind.TransactOpts{
		From: send.Address(),
		Signer: func(_ common.Address, tx *eth_types.Transaction) (*eth_types.Transaction, error) {
			signature, err := send.Sign(context, signer.Hash(tx))
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
		Context: context,
	}, nil
}

func GetDeviceAttributesTyped(metadata null.JSON, key string) []coremodels.DeviceTypeAttribute {
	var respAttrs []coremodels.DeviceTypeAttribute
	var ai map[string]any
	if err := metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[key]; ok && a != nil {
				attributes := ai[key].(map[string]any)
				for key, value := range attributes {
					v := fmt.Sprint(value)
					if len(v) > 0 {
						respAttrs = append(respAttrs, coremodels.DeviceTypeAttribute{
							Name:  key,
							Value: v,
						})
					}
				}
			}
		}
	}
	return respAttrs
}

func GetDefaultImageURL(ctx context.Context, definitionID string, db2 *sql.DB) string {

	all, err := models.Images(models.ImageWhere.DefinitionID.EQ(definitionID)).All(ctx, db2)
	if err != nil {
		return ""
	}
	imgURI := ""
	if all != nil {
		w := 0
		for _, image := range all {
			extra := 0
			if !image.NotExactImage {
				extra = 2000 // we want to give preference to exact images
			}
			if image.Width.Int+extra > w {
				w = image.Width.Int + extra
				imgURI = image.SourceURL
			}
		}
	}
	return imgURI
}

// note: below code is duplicated in identity-api

// BuildDeviceTypeAttributesTbland converts a list of DeviceTypeAttributeRequest to a JSON string for the given device type ID.
// It works the same as BuildDeviceTypeAttributes but the metadatakey is always "device_attributes" and does no attribute name validation
func BuildDeviceTypeAttributesTbland(attributes []*grpc.DeviceTypeAttributeRequest) string {
	if attributes == nil {
		return ""
	}
	deviceTypeInfo := coremodels.DeviceDefinitionMetadata{}
	metaData := make([]coremodels.DeviceTypeAttribute, len(attributes))
	for i, prop := range attributes {
		metaData[i].Name = prop.Name
		metaData[i].Value = prop.Value
	}
	deviceTypeInfo.DeviceAttributes = metaData
	j, _ := json.Marshal(deviceTypeInfo)
	return string(j)
}
