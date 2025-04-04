package common

import (
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"net/http"
	"sort"
	"strings"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/volatiletech/null/v8"
)

func JSONOrDefault(j null.JSON) json.RawMessage {
	if !j.Valid || len(j.JSON) == 0 {
		return []byte(`{}`)
	}
	return j.JSON
}

// Contains returns true if string exist in slice
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// SubModelsFromStylesDB gets the unique style.SubModel from the styles slice, deduping sub_model
func SubModelsFromStylesDB(styles repoModel.DeviceStyleSlice) []string {
	items := map[string]string{}
	for _, style := range styles {
		if _, ok := items[style.SubModel]; !ok {
			items[style.SubModel] = style.Name
		}
	}

	sm := make([]string, len(items))
	i := 0
	for key := range items {
		sm[i] = key
		i++
	}
	sort.Strings(sm)
	return sm
}

func BuildExternalIDs(externalIDsJSON null.JSON) []*models.ExternalID {
	var externalIDs []*models.ExternalID
	var ei map[string]string
	if err := externalIDsJSON.Unmarshal(&ei); err == nil {
		for vendor, id := range ei {
			externalIDs = append(externalIDs, &models.ExternalID{
				Vendor: vendor,
				ID:     id,
			})
		}
	}
	return externalIDs
}

func BuildDeviceMakeMetadata(metadataJSON null.JSON) *models.DeviceMakeMetadata {
	var dmMetadata *models.DeviceMakeMetadata
	var m map[string]string
	if err := metadataJSON.Unmarshal(&m); err == nil {
		dmMetadata = &models.DeviceMakeMetadata{
			RideGuideLink: m["RideGuideLink"],
		}
	}

	return dmMetadata
}

func ExternalIDsToGRPC(externalIDs []*models.ExternalID) []*grpc.ExternalID {
	externalIDsGRPC := make([]*grpc.ExternalID, len(externalIDs))
	for i, ei := range externalIDs {
		externalIDsGRPC[i] = &grpc.ExternalID{
			Vendor: ei.Vendor,
			Id:     ei.ID,
		}
	}
	return externalIDsGRPC
}

func DeviceMakeMetadataToGRPC(dm *models.DeviceMakeMetadata) *grpc.Metadata {
	dmMetadata := &grpc.Metadata{
		RideGuideLink: dm.RideGuideLink,
	}

	return dmMetadata
}

// GetDefaultImageURL if the images relation is not empty, looks for the best image to use based on some logic
func GetDefaultImageURL(images []*repoModel.Image) string {
	img := ""
	if images != nil {
		w := 0
		for _, image := range images {
			extra := 0
			if !image.NotExactImage {
				extra = 2000 // we want to give preference to exact images
			}
			if image.Width.Int+extra > w {
				w = image.Width.Int + extra
				img = image.SourceURL
			}
		}
	}
	return img
}

//func BuildFromDeviceDefinitionOnChainToQueryResult(dd *repoModel.DeviceDefinition) (*models.GetDeviceDefinitionQueryResult, error) {
//	if dd.R == nil || dd.R.DeviceMake == nil {
//		return nil, errors.New("DeviceMake relation cannot be nil, must be loaded in relation R.DeviceMake")
//	}
//	rp := &models.GetDeviceDefinitionQueryResult{
//		DeviceDefinitionID: dd.ID,
//		ExternalID:         dd.ExternalID.String,
//		Name:               BuildDeviceDefinitionName(dd.Year, dd.R.DeviceMake.Name, dd.Model),
//		Source:             dd.Source.String,
//		HardwareTemplateID: dd.HardwareTemplateID.String,
//		DeviceMake: models.DeviceMake{
//			ID:                 dd.R.DeviceMake.ID,
//			Name:               dd.R.DeviceMake.Name,
//			LogoURL:            dd.R.DeviceMake.LogoURL,
//			OemPlatformName:    dd.R.DeviceMake.OemPlatformName,
//			NameSlug:           dd.R.DeviceMake.NameSlug,
//			ExternalIDs:        JSONOrDefault(dd.R.DeviceMake.ExternalIds),
//			ExternalIDsTyped:   BuildExternalIDs(dd.R.DeviceMake.ExternalIds),
//			HardwareTemplateID: dd.R.DeviceMake.HardwareTemplateID,
//		},
//		Type: models.DeviceType{
//			//Type:      strings.TrimSpace(dd.R.DeviceType.ID),
//			Make:      dd.R.DeviceMake.Name,
//			Model:     dd.Model,
//			Year:      int(dd.Year),
//			MakeSlug:  dd.R.DeviceMake.NameSlug,
//			ModelSlug: dd.ModelSlug,
//		},
//		Metadata:    dd.Metadata.JSON,
//		Verified:    dd.Verified,
//		ExternalIDs: BuildExternalIDs(dd.ExternalIds),
//	}
//
//	if !dd.R.DeviceMake.TokenID.IsZero() {
//		rp.DeviceMake.TokenID = dd.R.DeviceMake.TokenID.Big.Int(new(big.Int))
//	}
//
//	// build object for integrations that have all the info
//	rp.DeviceIntegrations = []models.DeviceIntegration{}
//	rp.DeviceStyles = []models.DeviceStyle{}
//	rp.CompatibleIntegrations = []models.DeviceIntegration{}
//	rp.DeviceAttributes = []models.DeviceTypeAttribute{}
//
//	// pull out the device type device attributes, egGetDev. vehicle information
//	rp.DeviceAttributes = GetDeviceAttributesTyped(dd.Metadata, dd.R.DeviceType.Metadatakey)
//
//	if dd.R.DeviceIntegrations != nil {
//		for _, di := range dd.R.DeviceIntegrations {
//			deviceIntegration := models.DeviceIntegration{
//				ID:     di.R.Integration.ID,
//				Type:   di.R.Integration.Type,
//				Style:  di.R.Integration.Style,
//				Vendor: di.R.Integration.Vendor,
//				Region: di.Region,
//			}
//
//			if di.Features.Valid {
//				var deviceIntegrationFeature []models.DeviceIntegrationFeature
//				if err := di.Features.Unmarshal(&deviceIntegrationFeature); err == nil {
//					//nolint
//					deviceIntegration.Features = deviceIntegrationFeature
//				}
//			}
//			rp.DeviceIntegrations = append(rp.DeviceIntegrations, deviceIntegration)
//
//			rp.CompatibleIntegrations = rp.DeviceIntegrations
//		}
//	}
//
//	if dd.R.DefinitionDeviceStyles != nil {
//		rp.Type.SubModels = SubModelsFromStylesDB(dd.R.DefinitionDeviceStyles)
//
//		for _, ds := range dd.R.DefinitionDeviceStyles {
//			deviceStyle := models.DeviceStyle{
//				ID:                 ds.ID,
//				DeviceDefinitionID: ds.DeviceDefinitionID,
//				ExternalStyleID:    ds.ExternalStyleID,
//				Name:               ds.Name,
//				Source:             ds.Source,
//				SubModel:           ds.SubModel,
//				Metadata:           GetDeviceAttributesTyped(ds.Metadata, dd.R.DeviceType.Metadatakey),
//			}
//
//			if ds.HardwareTemplateID.Valid {
//				deviceStyle.HardwareTemplateID = ds.HardwareTemplateID.String
//			}
//
//			rp.DeviceStyles = append(rp.DeviceStyles, deviceStyle)
//		}
//	}
//	// trying pulling most recent images, pick the biggest one and where not exact image = false
//	rp.imageURL = GetDefaultImageURL(dd)
//
//	return rp, nil
//}

func GetDeviceAttributesTyped(metadata null.JSON, key string) []models.DeviceTypeAttribute {
	var respAttrs []models.DeviceTypeAttribute
	var ai map[string]any
	if err := metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[key]; ok && a != nil {
				attributes := ai[key].(map[string]any)
				for key, value := range attributes {
					respAttrs = append(respAttrs, models.DeviceTypeAttribute{
						Name:  key,
						Value: fmt.Sprint(value),
					})
				}
			}
		}
	}

	return respAttrs
}

func BuildFromDeviceDefinitionToQueryResult(dd *gateways.DeviceDefinitionTablelandModel, dm *models.DeviceMake, dss []repoModel.DeviceStyle, trx []repoModel.DefinitionTransaction) (*models.GetDeviceDefinitionQueryResult, error) {
	mdBytes := []byte("{}")
	if dd.Metadata != nil {
		mdBytes, _ = json.Marshal(dd.Metadata)
	}
	rp := &models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: dd.ID,
		NameSlug:           dd.ID,
		Name:               BuildDeviceDefinitionName(int16(dd.Year), dm.Name, dd.Model),
		HardwareTemplateID: DefautlAutoPiTemplate, // used for the autopi template id, which should now always be 130
		DeviceMake:         *dm,
		Metadata:           mdBytes,
		Verified:           true,
		ImageURL:           dd.ImageURI,
	}

	// build object for integrations that have all the info
	rp.DeviceStyles = []models.DeviceStyle{}
	rp.DeviceAttributes = []models.DeviceTypeAttribute{}

	// pull out the device type device attributes, egGetDev. vehicle information
	rp.DeviceAttributes = GetDeviceAttributesTyped(null.JSONFrom(mdBytes), dd.DeviceType)

	if dss != nil {
		for _, ds := range dss {
			deviceStyle := models.DeviceStyle{
				ID:              ds.ID,
				DefinitionID:    ds.DefinitionID,
				ExternalStyleID: ds.ExternalStyleID,
				Name:            ds.Name,
				Source:          ds.Source,
				SubModel:        ds.SubModel,
				Metadata:        GetDeviceAttributesTyped(ds.Metadata, dd.DeviceType),
			}

			if ds.HardwareTemplateID.Valid {
				deviceStyle.HardwareTemplateID = ds.HardwareTemplateID.String
			}

			rp.DeviceStyles = append(rp.DeviceStyles, deviceStyle)
		}
	}

	if trx != nil {
		rp.Transactions = make([]string, len(trx))
		for i, transaction := range trx {
			rp.Transactions[i] = transaction.TransactionHash
		}
	}

	return rp, nil
}

func BuildFromQueryResultToGRPC(dd *models.GetDeviceDefinitionQueryResult) *grpc.GetDeviceDefinitionItemResponse {
	rp := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: dd.DeviceDefinitionID,
		NameSlug:           dd.NameSlug,
		Name:               dd.Name,
		ImageUrl:           dd.ImageURL,

		HardwareTemplateId: DefautlAutoPiTemplate, //used for the autopi template id, which should always be 130 now
		Make: &grpc.DeviceMake{
			Id:              dd.DeviceMake.ID,
			Name:            dd.DeviceMake.Name,
			LogoUrl:         dd.DeviceMake.LogoURL.String,
			OemPlatformName: dd.DeviceMake.OemPlatformName.String,
			NameSlug:        dd.DeviceMake.NameSlug,
		},
		Verified:     dd.Verified,
		Transactions: dd.Transactions,
	}
	if dd.DeviceMake.TokenID != nil {
		rp.Make.TokenId = dd.DeviceMake.TokenID.Uint64()
	}

	rp.DeviceStyles = []*grpc.DeviceStyle{}
	for _, ds := range dd.DeviceStyles {
		rp.DeviceStyles = append(rp.DeviceStyles, &grpc.DeviceStyle{
			DeviceDefinitionId: dd.DeviceDefinitionID,
			ExternalStyleId:    ds.ExternalStyleID,
			Id:                 ds.ID,
			Name:               ds.Name,
			Source:             ds.Source,
			SubModel:           ds.SubModel,
			HardwareTemplateId: ds.HardwareTemplateID,
		})
	}

	rp.DeviceAttributes = []*grpc.DeviceTypeAttribute{}
	for _, da := range dd.DeviceAttributes {
		rp.DeviceAttributes = append(rp.DeviceAttributes, &grpc.DeviceTypeAttribute{
			Name:        da.Name,
			Label:       da.Label,
			Description: da.Description,
			Value:       da.Value,
			Required:    da.Required,
			Type:        da.Type,
			Options:     da.Option,
		})
	}

	return rp
}

func BuildDeviceTypeAttributes(attributes []*models.UpdateDeviceTypeAttribute, dt *repoModel.DeviceType) (null.JSON, error) {
	// attribute info
	if attributes == nil {
		return null.JSON{Valid: false}, nil
	}
	deviceTypeInfo := make(map[string]interface{})
	metaData := make(map[string]interface{})
	var ai map[string][]models.GetDeviceTypeAttributeQueryResult
	if err := dt.Properties.Unmarshal(&ai); err == nil {
		filterProperty := func(name string, items []models.GetDeviceTypeAttributeQueryResult) *models.GetDeviceTypeAttributeQueryResult {
			for _, attribute := range items {
				if name == attribute.Name {
					return &attribute
				}
			}
			return nil
		}

		for _, prop := range attributes {
			property := filterProperty(prop.Name, ai["properties"])

			if property == nil {
				return null.JSON{Valid: false}, &exceptions.ValidationError{
					Err: fmt.Errorf("invalid property %s", prop.Name),
				}
			}

			if property.Required && len(prop.Value) == 0 {
				return null.JSON{Valid: false}, &exceptions.ValidationError{
					Err: fmt.Errorf("property %s is required", prop.Name),
				}
			}

			metaData[property.Name] = prop.Value
		}
	}
	deviceTypeInfo[dt.Metadatakey] = metaData

	j, err := json.Marshal(deviceTypeInfo)
	if err != nil {
		return null.JSON{Valid: false}, nil
	}
	return null.JSONFrom(j), nil
}

func BuildDeviceDefinitionName(year int16, mk string, model string) string {
	return fmt.Sprintf("%d %s %s", year, mk, model)
}

func DeviceDefinitionSlug(makeSlug, modelSlug string, year int16) string {
	modelSlugCleaned := strings.ReplaceAll(modelSlug, ",", "")
	modelSlugCleaned = strings.ReplaceAll(modelSlugCleaned, "/", "-")
	modelSlugCleaned = strings.ReplaceAll(modelSlugCleaned, ".", "-")
	return fmt.Sprintf("%s_%s_%d", makeSlug, modelSlugCleaned, year)
}

func CheckTransactionStatus(txHash, apiKey string, useAmoy bool) (bool, error) {
	baseURL := "https://api.polygonscan.com"
	if useAmoy {
		baseURL = "https://amoy.polygonscan.com"
	}
	url := fmt.Sprintf("%s/api?module=transaction&action=gettxreceiptstatus&txhash=%s&apikey=%s", baseURL, txHash, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var txStatus TxStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&txStatus); err != nil {
		return false, err
	}

	// Check the transaction status
	if txStatus.Status == "1" && txStatus.Result.Status == "1" {
		return true, nil
	}
	return false, nil
}

type TxStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Status string `json:"status"`
	} `json:"result"`
}
