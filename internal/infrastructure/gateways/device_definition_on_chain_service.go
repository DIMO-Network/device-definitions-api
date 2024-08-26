//nolint:tagliatelle
package gateways

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"path"
	"strings"

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
	GetDeviceDefinitionByID(ctx context.Context, manufacturerID types.NullDecimal, ID string) (*models.DeviceDefinition, error)
	GetDeviceDefinitions(ctx context.Context, manufacturerID types.NullDecimal, ID string, model string, year int, pageIndex, pageSize int32) ([]*models.DeviceDefinition, error)
	CreateOrUpdate(ctx context.Context, make models.DeviceMake, dd models.DeviceDefinition) (*string, error)
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

func (e *deviceDefinitionOnChainService) GetDeviceDefinitionByID(ctx context.Context, manufacturerID types.NullDecimal, ID string) (*models.DeviceDefinition, error) {
	if manufacturerID.IsZero() {
		return nil, fmt.Errorf("manufacturerID has not value")
	}

	contractAddress := common.HexToAddress(e.settings.EthereumRegistryAddress)
	queryInstance, err := contracts.NewRegistry(contractAddress, e.client)
	if err != nil {
		return nil, fmt.Errorf("failed create NewRegistry: %w", err)
	}

	bigManufID := manufacturerID.Big.Int(new(big.Int))
	tableName, err := queryInstance.GetDeviceDefinitionTableName(&bind.CallOpts{Context: ctx, Pending: true}, bigManufID)
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

func transformToDefinition(item DeviceDefinitionTablelandModel) *models.DeviceDefinition {
	data := &models.DeviceDefinition{
		ID:    item.ID,
		Year:  item.Year,
		Model: item.Model,
	}

	if len(item.Metadata.DeviceAttributes) > 0 {
		deviceTypeInfo := make(map[string]interface{})
		metaData := make(map[string]interface{})

		for _, attr := range item.Metadata.DeviceAttributes {
			metaData[attr.Name] = attr.Value
		}

		deviceTypeInfo["vehicle_info"] = metaData
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

	contractAddress := common.HexToAddress(e.settings.EthereumRegistryAddress)
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

	fmt.Print(resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

func (e *deviceDefinitionOnChainService) CreateOrUpdate(ctx context.Context, make models.DeviceMake, dd models.DeviceDefinition) (*string, error) {

	const (
		TablelandRequests  = "Tableland_All_Request"
		TablelandFindByID  = "Tableland_FindByID_Request"
		TablelandCreated   = "Tableland_Created_Request"
		TablelandUpdated   = "Tableland_Updated_Request"
		TablelandNoUpdated = "Tableland_NoUpdated_Request"
		TablelandExists    = "Tableland_Exists_Request"
		TablelandErrors    = "Tableland_Error_Request"
	)

	metrics.Success.With(prometheus.Labels{"method": TablelandRequests}).Inc()
	e.logger.Info().Msgf("OnChain Start CreateOrUpdate for device definition %s. EthereumSendTransaction %t", dd.ID, e.settings.EthereumSendTransaction)

	if !e.settings.EthereumSendTransaction {
		return nil, nil
	}

	contractAddress := common.HexToAddress(e.settings.EthereumRegistryAddress)
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

	deviceInputs := contracts.DeviceDefinitionInput{
		Id:         dd.NameSlug,
		Model:      dd.Model,
		Year:       big.NewInt(int64(dd.Year)),
		Ksuid:      dd.ID,
		DeviceType: "vehicle",
	}

	deviceInputs.ImageURI = GetDefaultImageURL(dd)

	if dd.Metadata.Valid {
		attributes := GetDeviceAttributesTyped(dd.Metadata, common2.VehicleMetadataKey)
		type deviceAttributes struct {
			DeviceAttributes []DeviceTypeAttribute `json:"device_attributes"`
		}
		deviceAttributesStruct := deviceAttributes{
			DeviceAttributes: attributes,
		}
		jsonData, _ := json.Marshal(deviceAttributesStruct)
		deviceInputs.Metadata = string(jsonData)
	}

	// check if any pertinent information changed
	e.logger.Info().Msgf("Validating if device definition %s with tokenID %s exists in tableland", deviceInputs.Id, make.TokenID)
	currentDeviceDefinition, err := e.GetDeviceDefinitionByID(ctx, make.TokenID, deviceInputs.Id)
	if currentDeviceDefinition != nil {
		e.logger.Info().Msgf("DD %s found.", currentDeviceDefinition.ID)
	}
	if err != nil {
		e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
		metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
		e.logger.Err(err).Msgf("Error occurred get device definition %s from tableland.", deviceInputs.Id)
		return nil, err
	}
	metrics.Success.With(prometheus.Labels{"method": TablelandFindByID}).Inc()

	if currentDeviceDefinition != nil {
		metrics.Success.With(prometheus.Labels{"method": TablelandExists}).Inc()
		// validate if attributes was changed
		currentAttributes := GetDeviceAttributesTyped(currentDeviceDefinition.Metadata, common2.VehicleMetadataKey)
		newAttributes := GetDeviceAttributesTyped(dd.Metadata, common2.VehicleMetadataKey)
		newOrModified, removed := validateAttributes(currentAttributes, newAttributes)

		requiereUpdate := false
		if len(newOrModified) > 0 {
			requiereUpdate = true
		}

		if len(removed) > 0 {
			requiereUpdate = true
		}

		e.logger.Info().Msgf("newOrModified => %d and removed %d. Update %t", len(newOrModified), len(removed), requiereUpdate)

		if !requiereUpdate {
			metrics.Success.With(prometheus.Labels{"method": TablelandNoUpdated}).Inc()
			return nil, nil
		}

		// log what we are sending to the chain
		//jsonBytes, err := json.MarshalIndent(deviceInputs, "", "    ")
		//if err != nil {
		//	e.logger.Err(err).Msg("error marshalling device definition inputs")
		//}
		//e.logger.Info().RawJSON("device_definition", jsonBytes).Msg("dd payload sending to chain for CreateOrUpdate")

		e.logger.Info().Msgf("Executing UpdateDeviceDefinition %s with manufacturer ID %s", deviceInputs.Id, bigManufID)

		tx, err := instance.UpdateDeviceDefinition(auth, bigManufID, deviceInputs)
		if err != nil {
			e.logger.Err(err).Msgf("OnChainError - %s", dd.ID)
			metrics.InternalError.With(prometheus.Labels{"method": TablelandErrors}).Inc()
			return nil, fmt.Errorf("failed update UpdateDeviceDefinition: %w", err)
		}

		metrics.Success.With(prometheus.Labels{"method": TablelandUpdated}).Inc()

		trx := tx.Hash().Hex()

		e.logger.Info().Msgf("Executed UpdateDeviceDefinition %s with Trx %s in ManufacturerID %s", deviceInputs.Id, trx, bigManufID)

		return &trx, nil
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

type DeviceTypeAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DeviceDefinitionTablelandModel struct {
	ID       string `json:"id"`
	KSUID    string `json:"ksuid"`
	Model    string `json:"model"`
	Year     int16  `json:"year"`
	Metadata struct {
		DeviceAttributes []struct {
			Name  string `json:"name"`
			Value string `json:"value,omitempty"`
		} `json:"device_attributes"`
	} `json:"metadata"`
}
