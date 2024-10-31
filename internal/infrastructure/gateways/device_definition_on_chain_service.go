//nolint:tagliatelle
package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/tidwall/gjson"

	"github.com/pkg/errors"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/prometheus/client_golang/prometheus"

	common2 "github.com/DIMO-Network/device-definitions-api/internal/core/common"

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
)

//go:generate mockgen -source device_definition_on_chain_service.go -destination mocks/device_definition_on_chain_service_mock.go -package mocks
type DeviceDefinitionOnChainService interface {
	GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*models.DeviceDefinition, error)
	GetDefinitionByID(ctx context.Context, ID string, reader *db.DB) (*DeviceDefinitionTablelandModel, error)
	GetDefinitionTableland(ctx context.Context, manufacturerID *big.Int, ID string) (*DeviceDefinitionTablelandModel, error)
	GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]*models.DeviceDefinition, error)
	Create(ctx context.Context, make models.DeviceMake, dd models.DeviceDefinition) (*string, error)
	Update(ctx context.Context, manufacturerName string, input contracts.DeviceDefinitionUpdateInput) (*string, error)
}

type deviceDefinitionOnChainService struct {
	settings *config.Settings
	logger   *zerolog.Logger
	client   *ethclient.Client
	sender   sender.Sender
	chainID  *big.Int
}

func NewDeviceDefinitionOnChainService(settings *config.Settings, logger *zerolog.Logger, client *ethclient.Client, chainID *big.Int, sender sender.Sender) DeviceDefinitionOnChainService {
	return &deviceDefinitionOnChainService{
		settings: settings,
		logger:   logger,
		client:   client,
		chainID:  chainID,
		sender:   sender,
	}
}

// GetDeviceDefinitionByID gets dd from tableland with a select statement, returning a db model object
func (e *deviceDefinitionOnChainService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID *big.Int, ID string) (*models.DeviceDefinition, error) {
	if manufacturerID.Uint64() == 0 {
		return nil, fmt.Errorf("manufacturerID has not value")
	}

	contractAddress := e.settings.EthereumRegistryAddress
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerID)
	if err != nil {
		e.logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed get GetDeviceDefinitionTableName: %w", err)
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", tableName, ID)
	queryParams := map[string]string{
		"statement": statement,
	}

	e.logger.Info().Msgf("Tableland %s query => %s", tableName, statement)

	var modelTableland []DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, err
	}

	if len(modelTableland) > 0 {
		return transformToDefinition(modelTableland[0]), nil
	}

	return nil, nil
}

func (e *deviceDefinitionOnChainService) GetDefinitionByID(ctx context.Context, ID string, reader *db.DB) (*DeviceDefinitionTablelandModel, error) {
	split := strings.Split(ID, "_")
	if len(split) != 3 {
		return nil, fmt.Errorf("invalid slug")
	}
	manufacturerName := split[0]
	deviceMake, err := models.DeviceMakes(models.DeviceMakeWhere.Name.EQ(manufacturerName)).One(ctx, reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed get DeviceMake: %s", manufacturerName)
	}
	return e.GetDefinitionTableland(ctx, deviceMake.TokenID.Int(new(big.Int)), ID)
}

// GetDeviceDefinitionTableland gets dd from tableland with a select statement and returns tbl object
func (e *deviceDefinitionOnChainService) GetDefinitionTableland(ctx context.Context, manufacturerID *big.Int, ID string) (*DeviceDefinitionTablelandModel, error) {
	if manufacturerID == nil || manufacturerID.Uint64() == 0 {
		return nil, fmt.Errorf("manufacturerID cannot be 0")
	}

	contractAddress := e.settings.EthereumRegistryAddress
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true}, manufacturerID)
	if err != nil {
		e.logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed get GetDeviceDefinitionTableName: %w", err)
	}

	statement := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", tableName, ID)
	queryParams := map[string]string{
		"statement": statement,
	}

	var modelTableland []DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, err
	}
	if len(modelTableland) == 0 {
		return nil, nil
	}

	return &modelTableland[0], nil
}

func transformToDefinition(tblDD DeviceDefinitionTablelandModel) *models.DeviceDefinition {
	data := &models.DeviceDefinition{
		ID:           tblDD.ID,
		Year:         int16(tblDD.Year),
		Model:        tblDD.Model,
		DeviceTypeID: null.StringFrom(tblDD.DeviceType),
	}

	if tblDD.Metadata != nil && len(tblDD.Metadata.DeviceAttributes) > 0 {
		deviceTypeInfo := make(map[string]interface{})
		metaData := make(map[string]interface{})

		for _, attr := range tblDD.Metadata.DeviceAttributes {
			metaData[attr.Name] = attr.Value
		}

		jsonKey := common2.VehicleMetadataKey
		if tblDD.DeviceType == "aftermarket_device" {
			jsonKey = common2.AftermarketMetadataKey
		}

		deviceTypeInfo[jsonKey] = metaData
		j, err := json.Marshal(deviceTypeInfo)
		if err == nil {
			data.Metadata = null.JSONFrom(j)
		}
	}

	return data
}

func (e *deviceDefinitionOnChainService) GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]*models.DeviceDefinition, error) {
	if manufacturerID.IsZero() {
		return nil, fmt.Errorf("manufacturerID has not value")
	}

	contractAddress := e.settings.EthereumRegistryAddress
	fromAddress := e.sender.Address()
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	bigManufID := manufacturerID.Big.Int(new(big.Int))
	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true, From: fromAddress}, bigManufID)
	if err != nil {
		e.logger.Info().Msgf("%s", err)
		return nil, fmt.Errorf("failed get GetDeviceDefinitionTableName: %w", err)
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

	var modelTableland []DeviceDefinitionTablelandModel
	if err := e.QueryTableland(queryParams, &modelTableland); err != nil {
		return nil, err
	}

	result := make([]*models.DeviceDefinition, len(modelTableland))
	for i, item := range modelTableland {
		result[i] = transformToDefinition(item)
	}

	return result, nil
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return errors.Wrapf(err, "resp body: %s", string(body))
	}
	return nil
}

const (
	TablelandRequests = "Tableland_All_Request"
	TablelandFindByID = "Tableland_FindByID_Request"
	TablelandCreated  = "Tableland_Created_Request"
	TablelandUpdated  = "Tableland_Updated_Request"
	TablelandExists   = "Tableland_Exists_Request"
	TablelandErrors   = "Tableland_Error_Request"
)

// Create does a create for tableland, on-chain operation - checks if already exists
func (e *deviceDefinitionOnChainService) Create(ctx context.Context, make models.DeviceMake, dd models.DeviceDefinition) (*string, error) {

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

	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	// Validate if manufacturer exists
	bigManufID, err := queryInstance.GetManufacturerIdByName(&bind.CallOpts{Context: ctx, Pending: true}, make.Name)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed get GetManufacturerIdByName => %s: %w", make.Name, err)
	}
	instance, err := contracts.NewRegistryTransactor(contractAddress, e.client)
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		return nil, fmt.Errorf("failed create NewRegistryTransactor: %w", err)
	}

	if dd.DeviceTypeID.String == "" {
		return nil, fmt.Errorf("dd DeviceTypeId is required")
	}

	deviceInputs := contracts.DeviceDefinitionInput{
		Id:         dd.NameSlug,
		Model:      dd.Model,
		Year:       big.NewInt(int64(dd.Year)),
		Ksuid:      dd.ID,
		DeviceType: dd.DeviceTypeID.String,
		ImageURI:   GetDefaultImageURL(dd),
	}

	mdKey := common2.VehicleMetadataKey
	if dd.DeviceTypeID.String == "aftermarket_device" {
		mdKey = common2.AftermarketMetadataKey
	}

	if dd.Metadata.Valid {
		attributes := GetDeviceAttributesTyped(dd.Metadata, mdKey)
		type deviceAttributes struct {
			DeviceAttributes []DeviceTypeAttribute `json:"device_attributes"`
		}
		deviceAttributesStruct := deviceAttributes{
			DeviceAttributes: attributes,
		}
		jsonData, _ := json.Marshal(deviceAttributesStruct)
		deviceInputs.Metadata = string(jsonData)
	}

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

	return &trx, nil
}

// Update on-chain device definition, only has basic validation that some fields be present. Requires existing tableland record to exist to update
func (e *deviceDefinitionOnChainService) Update(ctx context.Context, manufacturerName string, input contracts.DeviceDefinitionUpdateInput) (*string, error) {

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

	return &trx, nil
}

func validateAttributes(current, new []DeviceTypeAttribute) ([]DeviceTypeAttribute, []DeviceTypeAttribute) {
	currentMap := attributesToMap(current)
	newMap := attributesToMap(new)

	var newOrModifiedAttributes []DeviceTypeAttribute
	var removedAttributes []DeviceTypeAttribute

	// Find new or changed attributes
	for name, newValue := range newMap {
		if currentValue, exists := currentMap[name]; !exists || currentValue != newValue {
			newOrModifiedAttributes = append(newOrModifiedAttributes, DeviceTypeAttribute{
				Name:  name,
				Value: newValue,
			})
		}
	}

	// Find deleted attributes
	for name, currentValue := range currentMap {
		if _, exists := newMap[name]; !exists {
			removedAttributes = append(removedAttributes, DeviceTypeAttribute{
				Name:  name,
				Value: currentValue,
			})
		}
	}

	return newOrModifiedAttributes, removedAttributes
}

func attributesToMap(attributes []DeviceTypeAttribute) map[string]string {
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

func GetDeviceAttributesTyped(metadata null.JSON, key string) []DeviceTypeAttribute {
	var respAttrs []DeviceTypeAttribute
	var ai map[string]any
	if err := metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[key]; ok && a != nil {
				attributes := ai[key].(map[string]any)
				for key, value := range attributes {
					v := fmt.Sprint(value)
					if len(v) > 0 {
						respAttrs = append(respAttrs, DeviceTypeAttribute{
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

func GetDefaultImageURL(dd models.DeviceDefinition) string {
	imgURI := ""
	if dd.R != nil && dd.R.Images != nil {
		w := 0
		for _, image := range dd.R.Images {
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

type DeviceDefinitionTablelandModel struct {
	ID         string                    `json:"id"`
	KSUID      string                    `json:"ksuid"`
	Model      string                    `json:"model"`
	Year       int                       `json:"year"`
	DeviceType string                    `json:"devicetype"`
	ImageURI   string                    `json:"imageuri"`
	Metadata   *DeviceDefinitionMetadata `json:"metadata"`
}

type DeviceDefinitionMetadata struct {
	DeviceAttributes []DeviceTypeAttribute `json:"device_attributes"`
}

type DeviceTypeAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// UnmarshalJSON customizes the unmarshaling of DeviceDefinitionTablelandModel to handle cases where metadata is an empty string.
func (d *DeviceDefinitionTablelandModel) UnmarshalJSON(data []byte) error {
	type Alias DeviceDefinitionTablelandModel // Create an alias to avoid recursion

	aux := &struct {
		Metadata json.RawMessage `json:"metadata"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Metadata) > 0 && string(aux.Metadata) != `""` {
		metadata := new(DeviceDefinitionMetadata)
		if err := json.Unmarshal(aux.Metadata, metadata); err != nil {
			return err
		}
		d.Metadata = metadata
	}

	return nil
}

// BuildDeviceTypeAttributesTbland converts a list of DeviceTypeAttributeRequest to a JSON string for the given device type ID.
// It works the same as BuildDeviceTypeAttributes but the metadatakey is always "device_attributes" and does no attribute name validation
func BuildDeviceTypeAttributesTbland(attributes []*grpc.DeviceTypeAttributeRequest) string {
	if attributes == nil {
		return ""
	}
	deviceTypeInfo := DeviceDefinitionMetadata{}
	metaData := make([]DeviceTypeAttribute, len(attributes))
	for i, prop := range attributes {
		metaData[i].Name = prop.Name
		metaData[i].Value = prop.Value
	}
	deviceTypeInfo.DeviceAttributes = metaData
	j, _ := json.Marshal(deviceTypeInfo)
	return string(j)
}
