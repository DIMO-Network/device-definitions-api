package common

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	repoModel "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	"github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

/* Terminal colors */

var Red = "\033[31m"
var Reset = "\033[0m"
var Green = "\033[32m"
var Purple = "\033[35m"

func PrintMMY(definition *repoModel.DeviceDefinition, color string, includeSource bool) string {
	mk := ""
	if definition.R != nil && definition.R.DeviceMake != nil {
		mk = definition.R.DeviceMake.Name
	}
	if !includeSource {
		return fmt.Sprintf("%s%d %s %s%s", color, definition.Year, mk, definition.Model, Reset)
	}
	return fmt.Sprintf("%s%d %s %s %s(source: %s)%s",
		color, definition.Year, mk, definition.Model, Purple, definition.Source.String, Reset)
}

func SlugString(term string) string {

	lowerCase := cases.Lower(language.English, cases.NoLower)
	lowerTerm := lowerCase.String(term)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	cleaned, _, _ := transform.String(t, lowerTerm)
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = strings.ReplaceAll(cleaned, "_", "-")

	return cleaned

}

func BuildExternalIds(externalIdsJSON null.JSON) []*models.ExternalID {
	var externalIds []*models.ExternalID
	var ei map[string]string
	if err := externalIdsJSON.Unmarshal(&ei); err == nil {
		for vendor, id := range ei {
			externalIds = append(externalIds, &models.ExternalID{
				Vendor: vendor,
				ID:     id,
			})
		}
	}
	return externalIds
}

func ExternalIdsToGRPC(externalIds []*models.ExternalID) []*grpc.ExternalID {
	externalIdsGRPC := make([]*grpc.ExternalID, len(externalIds))
	for i, ei := range externalIds {
		externalIdsGRPC[i] = &grpc.ExternalID{
			Vendor: ei.Vendor,
			Id:     ei.ID,
		}
	}
	return externalIdsGRPC
}

func ExternalIdsFromGRPC(externalIdsGRPC []*grpc.ExternalID) []*models.ExternalID {
	externalIds := make([]*models.ExternalID, len(externalIdsGRPC))
	for i, ei := range externalIdsGRPC {
		externalIds[i] = &models.ExternalID{
			Vendor: ei.Vendor,
			ID:     ei.Id,
		}
	}
	return externalIds
}

func BuildFromDeviceDefinitionToQueryResult(dd *repoModel.DeviceDefinition) (*models.GetDeviceDefinitionQueryResult, error) {
	if dd.R == nil || dd.R.DeviceMake == nil || dd.R.DeviceType == nil {
		return nil, errors.New("DeviceMake relation cannot be nil, must be loaded in relation R.DeviceMake")
	}
	rp := &models.GetDeviceDefinitionQueryResult{
		DeviceDefinitionID: dd.ID,
		ExternalID:         dd.ExternalID.String,
		Name:               fmt.Sprintf("%d %s %s", dd.Year, dd.R.DeviceMake.Name, dd.Model),
		ImageURL:           dd.ImageURL.String,
		Source:             dd.Source.String,
		DeviceMake: models.DeviceMake{
			ID:               dd.R.DeviceMake.ID,
			Name:             dd.R.DeviceMake.Name,
			LogoURL:          dd.R.DeviceMake.LogoURL,
			OemPlatformName:  dd.R.DeviceMake.OemPlatformName,
			NameSlug:         dd.R.DeviceMake.NameSlug,
			ExternalIds:      JSONOrDefault(dd.R.DeviceMake.ExternalIds),
			ExternalIdsTyped: BuildExternalIds(dd.R.DeviceMake.ExternalIds),
		},
		Type: models.DeviceType{
			Type:      strings.TrimSpace(dd.R.DeviceType.ID),
			Make:      dd.R.DeviceMake.Name,
			Model:     dd.Model,
			Year:      int(dd.Year),
			MakeSlug:  dd.R.DeviceMake.NameSlug,
			ModelSlug: dd.ModelSlug,
		},
		Metadata:    string(dd.Metadata.JSON),
		Verified:    dd.Verified,
		ExternalIds: BuildExternalIds(dd.ExternalIds),
	}

	if !dd.R.DeviceMake.TokenID.IsZero() {
		rp.DeviceMake.TokenID = dd.R.DeviceMake.TokenID.Big.Int(new(big.Int))
	}

	// vehicle info
	var vi map[string]models.VehicleInfo
	if err := dd.Metadata.Unmarshal(&vi); err == nil {
		//nolint
		rp.VehicleInfo = vi[dd.R.DeviceType.Metadatakey]
	}

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []models.DeviceIntegration{}
	rp.DeviceStyles = []models.DeviceStyle{}
	rp.CompatibleIntegrations = []models.DeviceIntegration{}
	rp.DeviceAttributes = []models.DeviceTypeAttribute{}

	// pull out the device type device attributes, eg. vehicle information
	var ai map[string]any
	if err := dd.Metadata.Unmarshal(&ai); err == nil {
		if ai != nil {
			if a, ok := ai[dd.R.DeviceType.Metadatakey]; ok && a != nil {
				attributes := ai[dd.R.DeviceType.Metadatakey].(map[string]any)
				for key, value := range attributes {
					rp.DeviceAttributes = append(rp.DeviceAttributes, models.DeviceTypeAttribute{
						Name:  key,
						Value: fmt.Sprint(value),
					})
				}
			}
		}
	}

	if dd.R.DeviceIntegrations != nil {
		for _, di := range dd.R.DeviceIntegrations {
			rp.DeviceIntegrations = append(rp.DeviceIntegrations, models.DeviceIntegration{
				ID:     di.R.Integration.ID,
				Type:   di.R.Integration.Type,
				Style:  di.R.Integration.Style,
				Vendor: di.R.Integration.Vendor,
				Region: di.Region,
			})

			rp.CompatibleIntegrations = rp.DeviceIntegrations
		}
	}

	if dd.R.DeviceStyles != nil {
		rp.Type.SubModels = SubModelsFromStylesDB(dd.R.DeviceStyles)

		for _, ds := range dd.R.DeviceStyles {
			rp.DeviceStyles = append(rp.DeviceStyles, models.DeviceStyle{
				ID:                 ds.ID,
				DeviceDefinitionID: ds.DeviceDefinitionID,
				ExternalStyleID:    ds.ExternalStyleID,
				Name:               ds.Name,
				Source:             ds.Source,
				SubModel:           ds.SubModel,
			})
		}
	}
	// trying pulling fuel images if no image_url, pick the biggest one
	if (dd.ImageURL.IsZero() || dd.ImageURL.String == "") && dd.R.Images != nil {
		w := 0
		for _, image := range dd.R.Images {
			if image.Width.Int > w {
				w = image.Width.Int
				rp.ImageURL = image.SourceURL
			}
		}
	}

	return rp, nil
}

func BuildFromQueryResultToGRPC(dd *models.GetDeviceDefinitionQueryResult) *grpc.GetDeviceDefinitionItemResponse {
	rp := &grpc.GetDeviceDefinitionItemResponse{
		DeviceDefinitionId: dd.DeviceDefinitionID,
		ExternalId:         dd.ExternalID,
		Name:               dd.Name,
		ImageUrl:           dd.ImageURL,
		Source:             dd.Source,
		Make: &grpc.DeviceMake{
			Id:              dd.DeviceMake.ID,
			Name:            dd.DeviceMake.Name,
			LogoUrl:         dd.DeviceMake.LogoURL.String,
			OemPlatformName: dd.DeviceMake.OemPlatformName.String,
			NameSlug:        dd.DeviceMake.NameSlug,
		},
		Type: &grpc.DeviceType{
			Type:      dd.Type.Type,
			Make:      dd.DeviceMake.Name,
			Model:     dd.Type.Model,
			Year:      int32(dd.Type.Year),
			MakeSlug:  dd.Type.MakeSlug,
			ModelSlug: dd.Type.ModelSlug,
		},
		Verified:    dd.Verified,
		ExternalIds: ExternalIdsToGRPC(dd.ExternalIds),
	}

	if dd.DeviceMake.TokenID != nil {
		rp.Make.TokenId = dd.DeviceMake.TokenID.Uint64()
	}

	// todo: remove this in future, now using device_attributes vehicle info
	//nolint
	numberOfDoors, _ := strconv.ParseInt(dd.VehicleInfo.NumberOfDoors, 6, 12)
	//nolint
	mpgHighway, _ := strconv.ParseFloat(dd.VehicleInfo.MPGHighway, 32)
	//nolint
	mpgCity, _ := strconv.ParseFloat(dd.VehicleInfo.MPGCity, 32)
	//nolint
	fuelTankCapacityGal, _ := strconv.ParseFloat(dd.VehicleInfo.FuelTankCapacityGal, 32)
	//nolint
	mpg, _ := strconv.ParseFloat(dd.VehicleInfo.MPG, 32)

	//nolint
	rp.VehicleData = &grpc.VehicleInfo{
		FuelType:            dd.VehicleInfo.FuelType,
		DrivenWheels:        dd.VehicleInfo.DrivenWheels,
		NumberOfDoors:       int32(numberOfDoors),
		Base_MSRP:           int32(dd.VehicleInfo.BaseMSRP),
		EPAClass:            dd.VehicleInfo.EPAClass,
		VehicleType:         dd.VehicleInfo.VehicleType,
		MPGHighway:          float32(mpgHighway),
		MPGCity:             float32(mpgCity),
		FuelTankCapacityGal: float32(fuelTankCapacityGal),
		MPG:                 float32(mpg),
	}

	// sub_models
	rp.Type.SubModels = dd.Type.SubModels

	// build object for integrations that have all the info
	rp.DeviceIntegrations = []*grpc.DeviceIntegration{}
	for _, di := range dd.DeviceIntegrations {
		rp.DeviceIntegrations = append(rp.DeviceIntegrations, &grpc.DeviceIntegration{
			DeviceDefinitionId: dd.DeviceDefinitionID,
			Integration: &grpc.Integration{
				Id:     di.ID,
				Type:   di.Type,
				Style:  di.Style,
				Vendor: di.Vendor,
			},
			Region: di.Region,
		})
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

func BuildDeviceTypeAttributes(attributes []*models.UpdateDeviceTypeAttribute, dt *repoModel.DeviceType) (map[string]interface{}, error) {
	// attribute info
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
				return nil, &exceptions.ValidationError{
					Err: fmt.Errorf("invalid property %s", prop.Name),
				}
			}

			if property.Required && len(prop.Value) == 0 {
				return nil, &exceptions.ValidationError{
					Err: fmt.Errorf("property %s is required", prop.Name),
				}
			}

			metaData[property.Name] = prop.Value
		}
	}

	deviceTypeInfo[dt.Metadatakey] = metaData

	return deviceTypeInfo, nil
}
