package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
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

func BuildExternalIDs(externalIDsJSON null.JSON) []*coremodels.ExternalID {
	var externalIDs []*coremodels.ExternalID
	var ei map[string]string
	if err := externalIDsJSON.Unmarshal(&ei); err == nil {
		for vendor, id := range ei {
			externalIDs = append(externalIDs, &coremodels.ExternalID{
				Vendor: vendor,
				ID:     id,
			})
		}
	}
	return externalIDs
}

func ExternalIDsToGRPC(externalIDs []*coremodels.ExternalID) []*grpc.ExternalID {
	externalIDsGRPC := make([]*grpc.ExternalID, len(externalIDs))
	for i, ei := range externalIDs {
		externalIDsGRPC[i] = &grpc.ExternalID{
			Vendor: ei.Vendor,
			Id:     ei.ID,
		}
	}
	return externalIDsGRPC
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

//func BuildFromDeviceDefinitionOnChainToQueryResult(dd *repoModel.DeviceDefinition) (*coremodels.GetDeviceDefinitionQueryResult, error) {
//	if dd.R == nil || dd.R.DeviceMake == nil {
//		return nil, errors.New("DeviceMake relation cannot be nil, must be loaded in relation R.DeviceMake")
//	}
//	rp := &coremodels.GetDeviceDefinitionQueryResult{
//		DefinitionID: dd.ID,
//		ExternalID:         dd.ExternalID.String,
//		Name:               BuildDeviceDefinitionName(dd.Year, dd.R.DeviceMake.Name, dd.Model),
//		Source:             dd.Source.String,
//		HardwareTemplateID: dd.HardwareTemplateID.String,
//		DeviceMake: coremodels.DeviceMake{
//			ID:                 dd.R.DeviceMake.ID,
//			Name:               dd.R.DeviceMake.Name,
//			LogoURL:            dd.R.DeviceMake.LogoURL,
//			OemPlatformName:    dd.R.DeviceMake.OemPlatformName,
//			NameSlug:           dd.R.DeviceMake.NameSlug,
//			ExternalIDs:        JSONOrDefault(dd.R.DeviceMake.ExternalIds),
//			ExternalIDsTyped:   BuildExternalIDs(dd.R.DeviceMake.ExternalIds),
//			HardwareTemplateID: dd.R.DeviceMake.HardwareTemplateID,
//		},
//		Type: coremodels.DeviceType{
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
//	rp.DeviceIntegrations = []coremodels.DeviceIntegration{}
//	rp.DeviceStyles = []coremodels.DeviceStyle{}
//	rp.CompatibleIntegrations = []coremodels.DeviceIntegration{}
//	rp.DeviceAttributes = []coremodels.DeviceTypeAttribute{}
//
//	// pull out the device type device attributes, egGetDev. vehicle information
//	rp.DeviceAttributes = GetDeviceAttributesTyped(dd.Metadata, dd.R.DeviceType.Metadatakey)
//
//	if dd.R.DeviceIntegrations != nil {
//		for _, di := range dd.R.DeviceIntegrations {
//			deviceIntegration := coremodels.DeviceIntegration{
//				ID:     di.R.Integration.ID,
//				Type:   di.R.Integration.Type,
//				Style:  di.R.Integration.Style,
//				Vendor: di.R.Integration.Vendor,
//				Region: di.Region,
//			}
//
//			if di.Features.Valid {
//				var deviceIntegrationFeature []coremodels.DeviceIntegrationFeature
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
//			deviceStyle := coremodels.DeviceStyle{
//				ID:                 ds.ID,
//				DefinitionID: ds.DefinitionID,
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

func GetDeviceAttributesTyped(metadata null.JSON, key string) []coremodels.DeviceTypeAttributeEditor {
	var respAttrs []coremodels.DeviceTypeAttributeEditor
	var ai map[string]any
	if err := metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[key]; ok && a != nil {
				attributes := ai[key].(map[string]any)
				for key, value := range attributes {
					respAttrs = append(respAttrs, coremodels.DeviceTypeAttributeEditor{
						Name:  key,
						Value: fmt.Sprint(value),
					})
				}
			}
		}
	}

	return respAttrs
}

func BuildFromDeviceDefinitionToQueryResult(dd *coremodels.DeviceDefinitionTablelandModel, dm *coremodels.Manufacturer, dss []*repoModel.DeviceStyle, trx []*repoModel.DefinitionTransaction) (*coremodels.GetDeviceDefinitionQueryResult, error) {
	mdBytes := []byte("{}")
	if dd.Metadata != nil {
		mdBytes, _ = json.Marshal(dd.Metadata)
	}
	rp := &coremodels.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: dd.ID,
		NameSlug:           dd.ID,
		Name:               BuildDeviceDefinitionName(int16(dd.Year), dm.Name, dd.Model),
		HardwareTemplateID: DefautlAutoPiTemplate, // used for the autopi template id, which should now always be 130
		MakeName:           dm.Name,
		MakeTokenID:        dm.TokenID,
		Metadata:           mdBytes,
		Verified:           true,
		ImageURL:           dd.ImageURI,
	}

	// build object for integrations that have all the info
	rp.DeviceStyles = []coremodels.DeviceStyle{}
	rp.DeviceAttributes = []coremodels.DeviceTypeAttributeEditor{}

	// pull out the device type device attributes, egGetDev. vehicle information
	rp.DeviceAttributes = GetDeviceAttributesTyped(null.JSONFrom(mdBytes), dd.DeviceType)

	for _, ds := range dss {
		deviceStyle := coremodels.DeviceStyle{
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

	if trx != nil {
		rp.Transactions = make([]string, len(trx))
		for i, transaction := range trx {
			rp.Transactions[i] = transaction.TransactionHash
		}
	}

	return rp, nil
}

func BuildDeviceTypeAttributes(attributes []*coremodels.UpdateDeviceTypeAttribute, dt *repoModel.DeviceType) (null.JSON, error) {
	// attribute info
	if attributes == nil {
		return null.JSON{Valid: false}, nil
	}
	deviceTypeInfo := make(map[string]interface{})
	metaData := make(map[string]interface{})
	var ai map[string][]coremodels.GetDeviceTypeAttributeQueryResult
	if err := dt.Properties.Unmarshal(&ai); err == nil {
		filterProperty := func(name string, items []coremodels.GetDeviceTypeAttributeQueryResult) *coremodels.GetDeviceTypeAttributeQueryResult {
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

func ConvertMetadataToDeviceAttributes(metadata *coremodels.DeviceDefinitionMetadata) []coremodels.DeviceTypeAttributeEditor {
	// Depending on your types, you might have to perform additional conversion logic.
	// Here we're simply returning the metadata as-is for assignment, but if your
	// DeviceAttributes require a specific structure, you'll have to adjust this.
	dta := []coremodels.DeviceTypeAttributeEditor{}
	if metadata == nil {
		return dta
	}
	for _, attribute := range metadata.DeviceAttributes {
		dta = append(dta, coremodels.DeviceTypeAttributeEditor{
			Name:        attribute.Name,
			Label:       attribute.Name,
			Description: attribute.Name,
			Type:        "",
			Required:    false,
			Value:       attribute.Value,
			Option:      nil,
		})
	}
	return dta
}

func ConvertDeviceTypeAttrsToDefinitionMetadata(attributes []*coremodels.UpdateDeviceTypeAttribute) *coremodels.DeviceDefinitionMetadata {
	ddm := &coremodels.DeviceDefinitionMetadata{
		DeviceAttributes: make([]coremodels.DeviceTypeAttribute, len(attributes)),
	}
	if len(attributes) == 0 {
		return nil
	}
	for i, attr := range attributes {
		ddm.DeviceAttributes[i] = coremodels.DeviceTypeAttribute{
			Name:  attr.Name,
			Value: attr.Value,
		}
	}
	return ddm
}

type TxStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		Status string `json:"status"`
	} `json:"result"`
}
