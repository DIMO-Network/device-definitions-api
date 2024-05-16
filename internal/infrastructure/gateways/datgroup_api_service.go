package gateways

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
)

//go:generate mockgen -source datgroup_api_service.go -destination mocks/datgroup_api_service_mock.go -package mocks
type DATGroupAPIService interface {
	GetVIN(vin, country string) (*GetVehicleIdentificationByVinResponse, error)
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

func (ai *datGroupAPIService) GetVIN(vin, userCountryISO2 string) (*GetVehicleIdentificationByVinResponse, error) {
	token, err := ai.getToken()
	if err != nil {
		return nil, err
	}
	if userCountryISO2 == "" || len(userCountryISO2) != 2 {
		userCountryISO2 = "US"
	}

	request := GenerateVehicleIdentificationByVinRequest{
		Request: GetVehicleIdentificationByVinRequest{
			VIN:         vin,
			Restriction: "ALL",
			Locale: LocaleRequest{
				Country:             userCountryISO2,
				DatCountryIndicator: "TR",
				Language:            "EN",
			},
		},
	}
	// temporary log for debugging
	ai.log.Info().Msgf("datgroup vin request: %+v", request)

	var result GetVehicleIdentificationByVinResponse
	err = soapCallHandleResponse(ai.Settings.DatGroupURL, "getVehicleIdentificationByVin", request, &result, token)
	if err != nil {
		return nil, err
	}

	// temporary log for debugging
	ai.log.Info().Msgf("datgroup vin response: %+v", result)

	if len(result.Body.GetDataVehicleIdentificationByVinResponse.VXS.Dossier) == 0 {
		return nil, fmt.Errorf("datgroup dosier response was empty")
	}

	return &result, nil
}

var tokenCache string
var tokenTimestamp time.Time

func (ai *datGroupAPIService) getToken() (string, error) {
	if tokenCache != "" && time.Since(tokenTimestamp).Minutes() < 15 {
		return tokenCache, nil
	}

	request := GenerateTokenRequest{
		Request: GetTokenRequest{
			CustomerLogin:             ai.Settings.DatGroupCustomerLogin,
			CustomerNumber:            ai.Settings.DatGroupCustomerNumber,
			CustomerPassword:          ai.Settings.DatGroupCustomerPassword,
			InterfacePartnerNumber:    ai.Settings.DatGroupInterfacePartnerNumber,
			InterfacePartnerSignature: ai.Settings.DatGroupInterfacePartnerSignature,
		},
	}

	var result GenerateTokenResponse
	err := soapCallHandleResponse(ai.Settings.DatGroupAUTHURL, "generateToken", request, &result, "")
	if err != nil {
		return "", err
	}

	tokenCache = result.Body.GetTokenResponse.Token
	tokenTimestamp = time.Now()

	return tokenCache, nil
}

func soapCallHandleResponse(ws string, action string, payloadInterface interface{}, result interface{}, token string) error {
	body, err := soapCall(ws, action, payloadInterface, token)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	return nil
}

func soapCall(ws string, action string, payloadInterface interface{}, token string) ([]byte, error) {
	v := soapRQ{
		XMLNsSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		XMLNsXSD:  "http://www.w3.org/2001/XMLSchema",
		XMLNsXSI:  "http://www.w3.org/2001/XMLSchema-instance",
		Body: soapBody{
			Payload: payloadInterface,
		},
	}
	payload, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	timeout := 30 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", ws, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/xml, multipart/related")
	req.Header.Set("SOAPAction", action)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	if len(token) > 0 {
		req.Header.Set("DAT-AuthorizationToken", token)
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return bodyBytes, nil
}

type soapRQ struct {
	XMLName   xml.Name `xml:"soap:Envelope"`
	XMLNsSoap string   `xml:"xmlns:soap,attr"`
	XMLNsXSI  string   `xml:"xmlns:xsi,attr"`
	XMLNsXSD  string   `xml:"xmlns:xsd,attr"`
	Body      soapBody
}

type soapBody struct {
	XMLName xml.Name `xml:"soap:Body"`
	Payload interface{}
}

type GenerateTokenRequest struct {
	XMLName xml.Name        `xml:"generateToken"`
	Request GetTokenRequest `xml:"request"`
}

type GetTokenRequest struct {
	CustomerLogin             string `xml:"customerLogin"`
	CustomerNumber            string `xml:"customerNumber"`
	CustomerPassword          string `xml:"customerPassword"`
	IncludePermissionData     string `xml:"includePermissionData,omitempty"`
	InterfacePartnerNumber    string `xml:"interfacePartnerNumber"`
	InterfacePartnerSignature string `xml:"interfacePartnerSignature"`
	ProductVariant            string `xml:"productVariant,omitempty"`
}

type GenerateTokenResponse struct {
	XMLName xml.Name                  `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    GenerateBodyTokenResponse `xml:"Body"`
}

type GenerateBodyTokenResponse struct {
	GetTokenResponse GetTokenResponse `xml:"generateTokenResponse"`
}

type GetTokenResponse struct {
	Token string `xml:"token"`
}

type GenerateVehicleIdentificationByVinRequest struct {
	XMLName xml.Name                             `xml:"getVehicleIdentificationByVin"`
	Request GetVehicleIdentificationByVinRequest `xml:"request"`
}

type GetVehicleIdentificationByVinRequest struct {
	VIN         string        `xml:"vin"`
	Restriction string        `xml:"restriction"`
	Locale      LocaleRequest `xml:"locale"`
}

type LocaleRequest struct {
	Country             string `xml:"country,attr"`
	DatCountryIndicator string `xml:"datCountryIndicator,attr"`
	Language            string `xml:"language,attr"`
}

type GetVehicleIdentificationByVinResponse struct {
	XMLName xml.Name                                  `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    GetBodyVehicleIdentificationByVinResponse `xml:"Body"`
}

type GetBodyVehicleIdentificationByVinResponse struct {
	GetDataVehicleIdentificationByVinResponse GetDataVehicleIdentificationByVinResponse `xml:"getVehicleIdentificationByVinResponse"`
}

type GetDataVehicleIdentificationByVinResponse struct {
	VXS struct {
		Dossier []*Dossier `xml:"Dossier,omitempty" json:"Dossier,omitempty"`
		Source  string     `xml:"source,attr,omitempty" json:"source,omitempty"`
		Type    string     `xml:"type,attr,omitempty" json:"type,omitempty"`
	} `xml:"VXS,omitempty" json:"VXS,omitempty"`
}

type Dossier struct {
	XMLName       xml.Name   `xml:"http://www.dat.de/vxs Dossier"`
	Name          string     `xml:"Name,omitempty" json:"Name,omitempty"`
	Description   string     `xml:"Description,omitempty" json:"Description,omitempty"`
	UUID          string     `xml:"UUID,omitempty" json:"UUID,omitempty"`
	ExternalID    string     `xml:"ExternalID,omitempty" json:"ExternalId,omitempty"`
	IdSDo         int64      `xml:"IdSDo,omitempty" json:"IdSDo,omitempty"`               // nolint
	IdSD3Local    int64      `xml:"IdSD3Local,omitempty" json:"IdSD3Local,omitempty"`     // nolint
	DossierId     int64      `xml:"DossierId,omitempty" json:"DossierId,omitempty"`       // nolint
	IdSD3Network  int64      `xml:"IdSD3Network,omitempty" json:"IdSD3Network,omitempty"` // nolint
	IdExtern      string     `xml:"IdExtern,omitempty" json:"IdExtern,omitempty"`         // nolint
	Country       string     `xml:"Country,omitempty" json:"Country,omitempty"`
	DossierType   string     `xml:"DossierType,omitempty" json:"DossierType,omitempty"`
	DossierOrigin string     `xml:"DossierOrigin,omitempty" json:"DossierOrigin,omitempty"`
	Vehicle       *Vehicle   `xml:"Vehicle,omitempty" json:"Vehicle,omitempty"`
	Images        *Images    `xml:"Images,omitempty" json:"Images,omitempty"`
	ImageList     *ImageList `xml:"ImageList,omitempty" json:"ImageList,omitempty"`
	VAT           *VAT       `xml:"VAT,omitempty" json:"VAT,omitempty"`
}

type VAT struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VAT"`

	VatType                             string   `xml:"VatType,omitempty" json:"VatType,omitempty"`
	VatAtConstructionTime               *float32 `xml:"VatAtConstructionTime,omitempty" json:"VatAtConstructionTime,omitempty"`
	DatVatAtConstructionTime            *float32 `xml:"DatVatAtConstructionTime,omitempty" json:"DatVatAtConstructionTime,omitempty"`
	BaseVatAtConstructionTime           *float32 `xml:"BaseVatAtConstructionTime,omitempty" json:"BaseVatAtConstructionTime,omitempty"`
	DatBaseVatAtConstructionTime        *float32 `xml:"DatBaseVatAtConstructionTime,omitempty" json:"DatBaseVatAtConstructionTime,omitempty"`
	AddOnTaxAtConstructionTime          *float32 `xml:"AddOnTaxAtConstructionTime,omitempty" json:"AddOnTaxAtConstructionTime,omitempty"`
	AddOnTaxApplication                 string   `xml:"AddOnTaxApplication,omitempty" json:"AddOnTaxApplication,omitempty"`
	PostTaxDifference                   *float32 `xml:"PostTaxDifference,omitempty" json:"PostTaxDifference,omitempty"`
	VatAtValuationTime                  *float32 `xml:"VatAtValuationTime,omitempty" json:"VatAtValuationTime,omitempty"`
	DatVatAtValuationTime               *float32 `xml:"DatVatAtValuationTime,omitempty" json:"DatVatAtValuationTime,omitempty"`
	VatAtCalculationTime                *float32 `xml:"VatAtCalculationTime,omitempty" json:"VatAtCalculationTime,omitempty"`
	VatAtSalesTime                      *float32 `xml:"VatAtSalesTime,omitempty" json:"VatAtSalesTime,omitempty"`
	DatVatAtSalesTime                   *float32 `xml:"DatVatAtSalesTime,omitempty" json:"DatVatAtSalesTime,omitempty"`
	VatAtPurchaseTime                   *float32 `xml:"VatAtPurchaseTime,omitempty" json:"VatAtPurchaseTime,omitempty"`
	DatVatAtPurchaseTime                *float32 `xml:"DatVatAtPurchaseTime,omitempty" json:"DatVatAtPurchaseTime,omitempty"`
	VATReplacementPartAtCalculationTime *float32 `xml:"VATReplacementPartAtCalculationTime,omitempty" json:"VATReplacementPartAtCalculationTime,omitempty"`
}

type ImageList struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs ImageList"`

	Image []*Image `xml:"Image,omitempty" json:"Image,omitempty"`
}

type Image struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Image"`

	Description                 string `xml:"Description,omitempty" json:"Description,omitempty"`
	DefaultImage                *bool  `xml:"DefaultImage,omitempty" json:"DefaultImage,omitempty"`
	ForValuation                *bool  `xml:"ForValuation,omitempty" json:"ForValuation,omitempty"`
	ForRepairCalculation        *bool  `xml:"ForRepairCalculation,omitempty" json:"ForRepairCalculation,omitempty"`
	ForMarketplace              *bool  `xml:"ForMarketplace,omitempty" json:"ForMarketplace,omitempty"`
	ListLabelVariable           string `xml:"ListLabelVariable,omitempty" json:"ListLabelVariable,omitempty"`
	ImageType                   string `xml:"ImageType,omitempty" json:"ImageType,omitempty"`
	Origin                      string `xml:"Origin,omitempty" json:"Origin,omitempty"`
	AssignedApplication         string `xml:"AssignedApplication,omitempty" json:"AssignedApplication,omitempty"`
	BitIndicatorImageAlteration string `xml:"BitIndicatorImageAlteration,omitempty" json:"BitIndicatorImageAlteration,omitempty"`
	ImageNumber                 int64  `xml:"ImageNumber,omitempty" json:"ImageNumber,omitempty"`
	Height                      int64  `xml:"Height,omitempty" json:"Height,omitempty"`
	Width                       int64  `xml:"Width,omitempty" json:"Width,omitempty"`
	RealFilename                string `xml:"RealFilename,omitempty" json:"RealFilename,omitempty"`
	ImageId                     string `xml:"ImageId,omitempty" json:"ImageId,omitempty"` //nolint
	RelativePath                string `xml:"RelativePath,omitempty" json:"RelativePath,omitempty"`
}

type Images struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Images"`

	Description                 string `xml:"Description,omitempty" json:"Description,omitempty"`
	DefaultImage                *bool  `xml:"DefaultImage,omitempty" json:"DefaultImage,omitempty"`
	ForValuation                *bool  `xml:"ForValuation,omitempty" json:"ForValuation,omitempty"`
	ForRepairCalculation        *bool  `xml:"ForRepairCalculation,omitempty" json:"ForRepairCalculation,omitempty"`
	ForMarketplace              *bool  `xml:"ForMarketplace,omitempty" json:"ForMarketplace,omitempty"`
	ListLabelVariable           string `xml:"ListLabelVariable,omitempty" json:"ListLabelVariable,omitempty"`
	ImageType                   string `xml:"ImageType,omitempty" json:"ImageType,omitempty"`
	Origin                      string `xml:"Origin,omitempty" json:"Origin,omitempty"`
	AssignedApplication         string `xml:"AssignedApplication,omitempty" json:"AssignedApplication,omitempty"`
	BitIndicatorImageAlteration string `xml:"BitIndicatorImageAlteration,omitempty" json:"BitIndicatorImageAlteration,omitempty"`
	ImageNumber                 int64  `xml:"ImageNumber,omitempty" json:"ImageNumber,omitempty"`
}

type Vehicle struct {
	XMLName                               xml.Name `xml:"http://www.dat.de/vxs Vehicle"`
	VehicleIdentNumber                    string   `xml:"VehicleIdentNumber,omitempty" json:"VehicleIdentNumber,omitempty"`
	DatECode                              string   `xml:"DatECode,omitempty" json:"DatECode,omitempty"`
	Container                             string   `xml:"Container,omitempty" json:"Container,omitempty"`
	ConstructionYear                      int64    `xml:"ConstructionYear,omitempty" json:"ConstructionYear,omitempty"`
	DatConstructionYear                   int64    `xml:"DatConstructionYear,omitempty" json:"DatConstructionYear,omitempty"`
	ConstructionMonth                     int64    `xml:"ConstructionMonth,omitempty" json:"ConstructionMonth,omitempty"`
	ConstructionTime                      int64    `xml:"ConstructionTime,omitempty" json:"ConstructionTime,omitempty"`
	ConstructionTimeFrom                  int64    `xml:"ConstructionTimeFrom,omitempty" json:"ConstructionTimeFrom,omitempty"`
	ConstructionTimeTo                    int64    `xml:"ConstructionTimeTo,omitempty" json:"ConstructionTimeTo,omitempty"`
	ConstructionTimePriceList             int64    `xml:"ConstructionTimePriceList,omitempty" json:"ConstructionTimePriceList,omitempty"`
	MileageEstimated                      int64    `xml:"MileageEstimated,omitempty" json:"MileageEstimated,omitempty"`
	MileageOdometer                       int64    `xml:"MileageOdometer,omitempty" json:"MileageOdometer,omitempty"`
	MileageAccordingUser                  *float32 `xml:"MileageAccordingUser,omitempty" json:"MileageAccordingUser,omitempty"`
	MileageType                           string   `xml:"MileageType,omitempty" json:"MileageType,omitempty"`
	MileageComment                        string   `xml:"MileageComment,omitempty" json:"MileageComment,omitempty"`
	SalesDescription                      string   `xml:"SalesDescription,omitempty" json:"SalesDescription,omitempty"`
	VehicleTypeName                       string   `xml:"VehicleTypeName,omitempty" json:"VehicleTypeName,omitempty"`
	VehicleTypeNameN                      string   `xml:"VehicleTypeNameN,omitempty" json:"VehicleTypeNameN,omitempty"`
	DatVehicleTypeNameN                   string   `xml:"DatVehicleTypeNameN,omitempty" json:"DatVehicleTypeNameN,omitempty"`
	ManufacturerName                      string   `xml:"ManufacturerName,omitempty" json:"ManufacturerName,omitempty"`
	DatManufacturerName                   string   `xml:"DatManufacturerName,omitempty" json:"DatManufacturerName,omitempty"`
	BaseModelName                         string   `xml:"BaseModelName,omitempty" json:"BaseModelName,omitempty"`
	DatBaseModelName                      string   `xml:"DatBaseModelName,omitempty" json:"DatBaseModelName,omitempty"`
	SubModelName                          string   `xml:"SubModelName,omitempty" json:"SubModelName,omitempty"`
	DatSubModelName                       string   `xml:"DatSubModelName,omitempty" json:"DatSubModelName,omitempty"`
	EngineNameManual                      string   `xml:"EngineNameManual,omitempty" json:"EngineNameManual,omitempty"`
	BodyNameManual                        string   `xml:"BodyNameManual,omitempty" json:"BodyNameManual,omitempty"`
	WheelbaseNameManual                   string   `xml:"WheelbaseNameManual,omitempty" json:"WheelbaseNameManual,omitempty"`
	PropulsionNameManual                  string   `xml:"PropulsionNameManual,omitempty" json:"PropulsionNameManual,omitempty"`
	DrivingCabNameManual                  string   `xml:"DrivingCabNameManual,omitempty" json:"DrivingCabNameManual,omitempty"`
	TonnageNameManual                     string   `xml:"TonnageNameManual,omitempty" json:"TonnageNameManual,omitempty"`
	ConstructionNameManual                string   `xml:"ConstructionNameManual,omitempty" json:"ConstructionNameManual,omitempty"`
	SuspensionNameManual                  string   `xml:"SuspensionNameManual,omitempty" json:"SuspensionNameManual,omitempty"`
	AxleCountNameManual                   string   `xml:"AxleCountNameManual,omitempty" json:"AxleCountNameManual,omitempty"`
	EquipmentLineNameManual               string   `xml:"EquipmentLineNameManual,omitempty" json:"EquipmentLineNameManual,omitempty"`
	GearboxNameManual                     string   `xml:"GearboxNameManual,omitempty" json:"GearboxNameManual,omitempty"`
	ContainerName                         string   `xml:"ContainerName,omitempty" json:"ContainerName,omitempty"`
	ContainerNameN                        string   `xml:"ContainerNameN,omitempty" json:"ContainerNameN,omitempty"`
	DatContainerNameN                     string   `xml:"DatContainerNameN,omitempty" json:"DatContainerNameN,omitempty"`
	MainTypeGroupName                     string   `xml:"MainTypeGroupName,omitempty" json:"MainTypeGroupName,omitempty"`
	VehicleType                           int64    `xml:"VehicleType,omitempty" json:"VehicleType,omitempty"`
	Manufacturer                          int64    `xml:"Manufacturer,omitempty" json:"Manufacturer,omitempty"`
	BaseModel                             int64    `xml:"BaseModel,omitempty" json:"BaseModel,omitempty"`
	AlternativeVehicleType                int64    `xml:"AlternativeVehicleType,omitempty" json:"AlternativeVehicleType,omitempty"`
	AlternativeManufacturer               int64    `xml:"AlternativeManufacturer,omitempty" json:"AlternativeManufacturer,omitempty"`
	AlternativeBaseModel                  int64    `xml:"AlternativeBaseModel,omitempty" json:"AlternativeBaseModel,omitempty"`
	SubModel                              int64    `xml:"SubModel,omitempty" json:"SubModel,omitempty"`
	AlternativeSubModel                   int64    `xml:"AlternativeSubModel,omitempty" json:"AlternativeSubModel,omitempty"`
	MainTypeGroup                         string   `xml:"MainTypeGroup,omitempty" json:"MainTypeGroup,omitempty"`
	IdentificationSource                  string   `xml:"IdentificationSource,omitempty" json:"IdentificationSource,omitempty"`
	Country                               string   `xml:"Country,omitempty" json:"Country,omitempty"`
	CountryTarget                         string   `xml:"CountryTarget,omitempty" json:"CountryTarget,omitempty"`
	IsDisengaged                          bool     `xml:"isDisengaged,omitempty" json:"isDisengaged,omitempty"`
	WithoutDistinctionEquStandardSpecial  bool     `xml:"withoutDistinctionEquStandardSpecial,omitempty" json:"withoutDistinctionEquStandardSpecial,omitempty"`
	IsWithManualTypeNames                 *bool    `xml:"IsWithManualTypeNames,omitempty" json:"IsWithManualTypeNames,omitempty"`
	IsDisengagedN                         *bool    `xml:"IsDisengagedN,omitempty" json:"IsDisengagedN,omitempty"`
	WithoutDistinctionEquStandardSpecialN *bool    `xml:"WithoutDistinctionEquStandardSpecialN,omitempty" json:"WithoutDistinctionEquStandardSpecialN,omitempty"`
	IsUniversalSubModel                   *bool    `xml:"IsUniversalSubModel,omitempty" json:"IsUniversalSubModel,omitempty"`
	VinAccuracy                           int64    `xml:"VinAccuracy,omitempty" json:"VinAccuracy,omitempty"`
	VinActive                             *bool    `xml:"VinActive,omitempty" json:"VinActive,omitempty"`
	ReleaseIndicator                      string   `xml:"ReleaseIndicator,omitempty" json:"ReleaseIndicator,omitempty"`
	KbaNumbersN                           struct {
		KbaNumber []string `xml:"KbaNumber,omitempty" json:"KbaNumber,omitempty"`
	} `xml:"KbaNumbersN,omitempty" json:"KbaNumbersN,omitempty"`
	NationalCodeAustria struct {
		NationalCodeAustria []string `xml:"NationalCodeAustria,omitempty" json:"NationalCodeAustria,omitempty"`
	} `xml:"NationalCodeAustria,omitempty" json:"NationalCodeAustria,omitempty"`
	TypeOfConstruction     string `xml:"TypeOfConstruction,omitempty" json:"TypeOfConstruction,omitempty"`
	ConstructionYearManual string `xml:"ConstructionYearManual,omitempty" json:"ConstructionYearManual,omitempty"`
	ColorScheme            string `xml:"ColorScheme,omitempty" json:"ColorScheme,omitempty"`
	ColorSchemeManual      string `xml:"ColorSchemeManual,omitempty" json:"ColorSchemeManual,omitempty"`
	ColorVariant           string `xml:"ColorVariant,omitempty" json:"ColorVariant,omitempty"`
	PaintTypes             struct {
		PaintType []string `xml:"PaintType,omitempty" json:"PaintType,omitempty"`
	} `xml:"PaintTypes,omitempty" json:"PaintTypes,omitempty"`
	GeneralInspectionDate       string             `xml:"GeneralInspectionDate,omitempty" json:"GeneralInspectionDate,omitempty"`
	ManufacturerOrderKey        string             `xml:"ManufacturerOrderKey,omitempty" json:"ManufacturerOrderKey,omitempty"`
	SubModelVariant             int64              `xml:"SubModelVariant,omitempty" json:"SubModelVariant,omitempty"`
	TokenColorScheme            string             `xml:"TokenColorScheme,omitempty" json:"TokenColorScheme,omitempty"`
	VehicleTypeAUFromKba        string             `xml:"VehicleTypeAUFromKba,omitempty" json:"VehicleTypeAUFromKba,omitempty"`
	VehicleTypeFromKba          string             `xml:"VehicleTypeFromKba,omitempty" json:"VehicleTypeFromKba,omitempty"`
	VehicleTypeFromManufacturer string             `xml:"VehicleTypeFromManufacturer,omitempty" json:"VehicleTypeFromManufacturer,omitempty"`
	Colorcode                   int64              `xml:"Colorcode,omitempty" json:"Colorcode,omitempty"`
	SubTypeSubstitution         int64              `xml:"SubTypeSubstitution,omitempty" json:"SubTypeSubstitution,omitempty"`
	OriginalPrice               *float32           `xml:"OriginalPrice,omitempty" json:"OriginalPrice,omitempty"`
	DatOriginalPrice            *float32           `xml:"DatOriginalPrice,omitempty" json:"DatOriginalPrice,omitempty"`
	OriginalPriceGross          *float32           `xml:"OriginalPriceGross,omitempty" json:"OriginalPriceGross,omitempty"`
	DatOriginalPriceGross       *float32           `xml:"DatOriginalPriceGross,omitempty" json:"DatOriginalPriceGross,omitempty"`
	RentalCarClass              int64              `xml:"RentalCarClass,omitempty" json:"RentalCarClass,omitempty"`
	OriginalPriceInfo           *OriginalPriceInfo `xml:"OriginalPriceInfo,omitempty" json:"OriginalPriceInfo,omitempty"`
	Engine                      *Engine            `xml:"Engine,omitempty" json:"Engine,omitempty"`
	TechInfo                    *TechInfo          `xml:"TechInfo,omitempty" json:"TechInfo,omitempty"`
	Equipment                   *Equipment         `xml:"Equipment,omitempty" json:"Equipment,omitempty"`
	Tires                       *Tires             `xml:"Tires,omitempty" json:"Tires,omitempty"`
	// DATECodeEquipment           *EquipSequence     `xml:"DATECodeEquipment,omitempty" json:"DATECodeEquipment,omitempty"`
	VINResult        *VINResult `xml:"VINResult,omitempty" json:"VINResult,omitempty"`
	TokenOfVinResult string     `xml:"TokenOfVinResult,omitempty" json:"TokenOfVinResult,omitempty"`
	BuildYear        int64      `xml:"BuildYear,omitempty" json:"BuildYear,omitempty"`
	OperatingHours   int64      `xml:"OperatingHours,omitempty" json:"OperatingHours,omitempty"`
	MileageInMiles   int64      `xml:"MileageInMiles,omitempty" json:"MileageInMiles,omitempty"`
	VehicleCondition string     `xml:"VehicleCondition,omitempty" json:"VehicleCondition,omitempty"`
}

type FillingQuantities struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs FillingQuantities"`

	Fluid []*Fluid `xml:"Fluid,omitempty" json:"Fluid,omitempty"`
}

type Fluid struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Fluid"`

	Capacity       []*Capacity       `xml:"Capacity,omitempty" json:"Capacity,omitempty"`
	Recommendation []*Recommendation `xml:"Recommendation,omitempty" json:"Recommendation,omitempty"`

	Type int32  `xml:"type,attr,omitempty" json:"type,omitempty"`
	Desc string `xml:"desc,attr,omitempty" json:"desc,omitempty"`
	Code string `xml:"code,attr,omitempty" json:"code,omitempty"`
}

type Usage struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Usage"`

	Value string `xml:",chardata" json:"-,"`
	Type  int32  `xml:"type,attr,omitempty" json:"type,omitempty"`
}

type Recommendation struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Recommendation"`

	Usage []*Usage `xml:"Usage,omitempty" json:"Usage,omitempty"`

	Interval []*Interval `xml:"Interval,omitempty" json:"Interval,omitempty"`

	Product []string `xml:"Product,omitempty" json:"Product,omitempty"`
}

type Interval struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Interval"`

	Value string `xml:",chardata" json:"-,"`
	Type  int32  `xml:"type,attr,omitempty" json:"type,omitempty"`
}

type Capacity struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Capacity"`

	Type      int32   `xml:"type,attr,omitempty" json:"type,omitempty"`
	Desc      string  `xml:"desc,attr,omitempty" json:"desc,omitempty"`
	Min       float64 `xml:"min,attr,omitempty" json:"min,omitempty"`
	Max       float64 `xml:"max,attr,omitempty" json:"max,omitempty"`
	Unit      string  `xml:"unit,attr,omitempty" json:"unit,omitempty"`
	Condition string  `xml:"condition,attr,omitempty" json:"condition,omitempty"`
}

type TechInfo struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs TechInfo"`

	FillingQuantities                     *FillingQuantities `xml:"FillingQuantities,omitempty" json:"FillingQuantities,omitempty"`
	StructureType                         string             `xml:"StructureType,omitempty" json:"StructureType,omitempty"`
	StructureDescription                  string             `xml:"StructureDescription,omitempty" json:"StructureDescription,omitempty"`
	CabineStructureType                   string             `xml:"CabineStructureType,omitempty" json:"CabineStructureType,omitempty"`
	CabineStructureDescription            string             `xml:"CabineStructureDescription,omitempty" json:"CabineStructureDescription,omitempty"`
	UpperBodyStructureType                string             `xml:"UpperBodyStructureType,omitempty" json:"UpperBodyStructureType,omitempty"`
	UpperBodyStructureDescription         string             `xml:"UpperBodyStructureDescription,omitempty" json:"UpperBodyStructureDescription,omitempty"`
	UpperBodyStructureDescriptionUser     string             `xml:"UpperBodyStructureDescriptionUser,omitempty" json:"UpperBodyStructureDescriptionUser,omitempty"`
	UpperBodyStructureAndVersion          string             `xml:"UpperBodyStructureAndVersion,omitempty" json:"UpperBodyStructureAndVersion,omitempty"`
	CountOfAxles                          int64              `xml:"CountOfAxles,omitempty" json:"CountOfAxles,omitempty"`
	DatCountOfAxles                       int64              `xml:"DatCountOfAxles,omitempty" json:"DatCountOfAxles,omitempty"`
	CountOfDrivedAxles                    int64              `xml:"CountOfDrivedAxles,omitempty" json:"CountOfDrivedAxles,omitempty"`
	DatCountOfDrivedAxles                 int64              `xml:"DatCountOfDrivedAxles,omitempty" json:"DatCountOfDrivedAxles,omitempty"`
	WheelBase                             int64              `xml:"WheelBase,omitempty" json:"WheelBase,omitempty"`
	DatWheelBase                          int64              `xml:"DatWheelBase,omitempty" json:"DatWheelBase,omitempty"`
	WheelBase2                            int64              `xml:"WheelBase2,omitempty" json:"WheelBase2,omitempty"`
	AxleLoadFront                         int64              `xml:"AxleLoadFront,omitempty" json:"AxleLoadFront,omitempty"`
	AxleLoadMiddle                        int64              `xml:"AxleLoadMiddle,omitempty" json:"AxleLoadMiddle,omitempty"`
	AxleLoadBack                          int64              `xml:"AxleLoadBack,omitempty" json:"AxleLoadBack,omitempty"`
	TonnageClass                          string             `xml:"TonnageClass,omitempty" json:"TonnageClass,omitempty"`
	Length                                int64              `xml:"Length,omitempty" json:"Length,omitempty"`
	DatLength                             int64              `xml:"DatLength,omitempty" json:"DatLength,omitempty"`
	Width                                 int64              `xml:"Width,omitempty" json:"Width,omitempty"`
	DatWidth                              int64              `xml:"DatWidth,omitempty" json:"DatWidth,omitempty"`
	Height                                int64              `xml:"Height,omitempty" json:"Height,omitempty"`
	DatHeight                             int64              `xml:"DatHeight,omitempty" json:"DatHeight,omitempty"`
	RoofLoad                              int64              `xml:"RoofLoad,omitempty" json:"RoofLoad,omitempty"`
	DatRoofLoad                           int64              `xml:"DatRoofLoad,omitempty" json:"DatRoofLoad,omitempty"`
	TrailerLoadBraked                     int64              `xml:"TrailerLoadBraked,omitempty" json:"TrailerLoadBraked,omitempty"`
	DatTrailerLoadBraked                  int64              `xml:"DatTrailerLoadBraked,omitempty" json:"DatTrailerLoadBraked,omitempty"`
	TrailerLoadUnbraked                   int64              `xml:"TrailerLoadUnbraked,omitempty" json:"TrailerLoadUnbraked,omitempty"`
	DatTrailerLoadUnbraked                int64              `xml:"DatTrailerLoadUnbraked,omitempty" json:"DatTrailerLoadUnbraked,omitempty"`
	VehicleSeats                          int64              `xml:"VehicleSeats,omitempty" json:"VehicleSeats,omitempty"`
	DatVehicleSeats                       int64              `xml:"DatVehicleSeats,omitempty" json:"DatVehicleSeats,omitempty"`
	VehicleDoors                          int64              `xml:"VehicleDoors,omitempty" json:"VehicleDoors,omitempty"`
	DatVehicleDoors                       int64              `xml:"DatVehicleDoors,omitempty" json:"DatVehicleDoors,omitempty"`
	CountOfAirbags                        int64              `xml:"CountOfAirbags,omitempty" json:"CountOfAirbags,omitempty"`
	DatCountOfAirbags                     int64              `xml:"DatCountOfAirbags,omitempty" json:"DatCountOfAirbags,omitempty"`
	Acceleration                          *float32           `xml:"Acceleration,omitempty" json:"Acceleration,omitempty"`
	DatAcceleration                       *float32           `xml:"DatAcceleration,omitempty" json:"DatAcceleration,omitempty"`
	SpeedMax                              int64              `xml:"SpeedMax,omitempty" json:"SpeedMax,omitempty"`
	DatSpeedMax                           int64              `xml:"DatSpeedMax,omitempty" json:"DatSpeedMax,omitempty"`
	PowerHp                               int64              `xml:"PowerHp,omitempty" json:"PowerHp,omitempty"`
	DatPowerHp                            int64              `xml:"DatPowerHp,omitempty" json:"DatPowerHp,omitempty"`
	PowerKw                               *float32           `xml:"PowerKw,omitempty" json:"PowerKw,omitempty"`
	DatPowerKw                            *float32           `xml:"DatPowerKw,omitempty" json:"DatPowerKw,omitempty"`
	Capacity                              int64              `xml:"Capacity,omitempty" json:"Capacity,omitempty"`
	DatCapacity                           int64              `xml:"DatCapacity,omitempty" json:"DatCapacity,omitempty"`
	Cylinder                              int64              `xml:"Cylinder,omitempty" json:"Cylinder,omitempty"`
	DatCylinder                           int64              `xml:"DatCylinder,omitempty" json:"DatCylinder,omitempty"`
	CylinderArrangement                   string             `xml:"CylinderArrangement,omitempty" json:"CylinderArrangement,omitempty"`
	DatCylinderArrangement                string             `xml:"DatCylinderArrangement,omitempty" json:"DatCylinderArrangement,omitempty"`
	RotationsOnMaxPower                   int64              `xml:"RotationsOnMaxPower,omitempty" json:"RotationsOnMaxPower,omitempty"`
	DatRotationsOnMaxPower                int64              `xml:"DatRotationsOnMaxPower,omitempty" json:"DatRotationsOnMaxPower,omitempty"`
	RotationsOnMaxTorque                  int64              `xml:"RotationsOnMaxTorque,omitempty" json:"RotationsOnMaxTorque,omitempty"`
	DatRotationsOnMaxTorque               int64              `xml:"DatRotationsOnMaxTorque,omitempty" json:"DatRotationsOnMaxTorque,omitempty"`
	Torque                                int64              `xml:"Torque,omitempty" json:"Torque,omitempty"`
	DatTorque                             int64              `xml:"DatTorque,omitempty" json:"DatTorque,omitempty"`
	GearboxType                           string             `xml:"GearboxType,omitempty" json:"GearboxType,omitempty"`
	NrOfGears                             string             `xml:"NrOfGears,omitempty" json:"NrOfGears,omitempty"`
	OriginalTireSizeAxle1                 string             `xml:"OriginalTireSizeAxle1,omitempty" json:"OriginalTireSizeAxle1,omitempty"`
	OriginalTireSizeAxle2                 string             `xml:"OriginalTireSizeAxle2,omitempty" json:"OriginalTireSizeAxle2,omitempty"`
	TankVolume                            int64              `xml:"TankVolume,omitempty" json:"TankVolume,omitempty"`
	DatTankVolume                         int64              `xml:"DatTankVolume,omitempty" json:"DatTankVolume,omitempty"`
	TankVolumeAlternative                 int64              `xml:"TankVolumeAlternative,omitempty" json:"TankVolumeAlternative,omitempty"`
	DatTankVolumeAlternative              int64              `xml:"DatTankVolumeAlternative,omitempty" json:"DatTankVolumeAlternative,omitempty"`
	ConsumptionInTown                     *float32           `xml:"ConsumptionInTown,omitempty" json:"ConsumptionInTown,omitempty"`
	DatConsumptionInTown                  *float32           `xml:"DatConsumptionInTown,omitempty" json:"DatConsumptionInTown,omitempty"`
	ConsumptionOutOfTown                  *float32           `xml:"ConsumptionOutOfTown,omitempty" json:"ConsumptionOutOfTown,omitempty"`
	DatConsumptionOutOfTown               *float32           `xml:"DatConsumptionOutOfTown,omitempty" json:"DatConsumptionOutOfTown,omitempty"`
	Consumption                           *float32           `xml:"Consumption,omitempty" json:"Consumption,omitempty"`
	DatConsumption                        *float32           `xml:"DatConsumption,omitempty" json:"DatConsumption,omitempty"`
	WltpConsumptionMixedMin               *float32           `xml:"WltpConsumptionMixedMin,omitempty" json:"WltpConsumptionMixedMin,omitempty"`
	DatWltpConsumptionMixedMin            *float32           `xml:"DatWltpConsumptionMixedMin,omitempty" json:"DatWltpConsumptionMixedMin,omitempty"`
	WltpConsumptionMixedMax               *float32           `xml:"WltpConsumptionMixedMax,omitempty" json:"WltpConsumptionMixedMax,omitempty"`
	DatWltpConsumptionMixedMax            *float32           `xml:"DatWltpConsumptionMixedMax,omitempty" json:"DatWltpConsumptionMixedMax,omitempty"`
	ConsumptionInnerCng                   *float32           `xml:"ConsumptionInnerCng,omitempty" json:"ConsumptionInnerCng,omitempty"`
	DatConsumptionInnerCng                *float32           `xml:"DatConsumptionInnerCng,omitempty" json:"DatConsumptionInnerCng,omitempty"`
	ConsumptionOuterCng                   *float32           `xml:"ConsumptionOuterCng,omitempty" json:"ConsumptionOuterCng,omitempty"`
	DatConsumptionOuterCng                *float32           `xml:"DatConsumptionOuterCng,omitempty" json:"DatConsumptionOuterCng,omitempty"`
	ConsumptionMixCng                     *float32           `xml:"ConsumptionMixCng,omitempty" json:"ConsumptionMixCng,omitempty"`
	DatConsumptionMixCng                  *float32           `xml:"DatConsumptionMixCng,omitempty" json:"DatConsumptionMixCng,omitempty"`
	WltpConsumptionBivalentMixedCngMin    *float32           `xml:"WltpConsumptionBivalentMixedCngMin,omitempty" json:"WltpConsumptionBivalentMixedCngMin,omitempty"`
	DatWltpConsumptionBivalentMixedCngMin *float32           `xml:"DatWltpConsumptionBivalentMixedCngMin,omitempty" json:"DatWltpConsumptionBivalentMixedCngMin,omitempty"`
	WltpConsumptionBivalentMixedCngMax    *float32           `xml:"WltpConsumptionBivalentMixedCngMax,omitempty" json:"WltpConsumptionBivalentMixedCngMax,omitempty"`
	DatWltpConsumptionBivalentMixedCngMax *float32           `xml:"DatWltpConsumptionBivalentMixedCngMax,omitempty" json:"DatWltpConsumptionBivalentMixedCngMax,omitempty"`
	ConsumptionInnerLpg                   *float32           `xml:"ConsumptionInnerLpg,omitempty" json:"ConsumptionInnerLpg,omitempty"`
	DatConsumptionInnerLpg                *float32           `xml:"DatConsumptionInnerLpg,omitempty" json:"DatConsumptionInnerLpg,omitempty"`
	ConsumptionOuterLpg                   *float32           `xml:"ConsumptionOuterLpg,omitempty" json:"ConsumptionOuterLpg,omitempty"`
	DatConsumptionOuterLpg                *float32           `xml:"DatConsumptionOuterLpg,omitempty" json:"DatConsumptionOuterLpg,omitempty"`
	ConsumptionMixLpg                     *float32           `xml:"ConsumptionMixLpg,omitempty" json:"ConsumptionMixLpg,omitempty"`
	DatConsumptionMixLpg                  *float32           `xml:"DatConsumptionMixLpg,omitempty" json:"DatConsumptionMixLpg,omitempty"`
	WltpConsumptionBivalentMixedLpgMin    *float32           `xml:"WltpConsumptionBivalentMixedLpgMin,omitempty" json:"WltpConsumptionBivalentMixedLpgMin,omitempty"`
	DatWltpConsumptionBivalentMixedLpgMin *float32           `xml:"DatWltpConsumptionBivalentMixedLpgMin,omitempty" json:"DatWltpConsumptionBivalentMixedLpgMin,omitempty"`
	WltpConsumptionBivalentMixedLpgMax    *float32           `xml:"WltpConsumptionBivalentMixedLpgMax,omitempty" json:"WltpConsumptionBivalentMixedLpgMax,omitempty"`
	DatWltpConsumptionBivalentMixedLpgMax *float32           `xml:"DatWltpConsumptionBivalentMixedLpgMax,omitempty" json:"DatWltpConsumptionBivalentMixedLpgMax,omitempty"`
	ConsumptionInnerH                     *float32           `xml:"ConsumptionInnerH,omitempty" json:"ConsumptionInnerH,omitempty"`
	DatConsumptionInnerH                  *float32           `xml:"DatConsumptionInnerH,omitempty" json:"DatConsumptionInnerH,omitempty"`
	ConsumptionOuterH                     *float32           `xml:"ConsumptionOuterH,omitempty" json:"ConsumptionOuterH,omitempty"`
	DatConsumptionOuterH                  *float32           `xml:"DatConsumptionOuterH,omitempty" json:"DatConsumptionOuterH,omitempty"`
	ConsumptionMixH                       *float32           `xml:"ConsumptionMixH,omitempty" json:"ConsumptionMixH,omitempty"`
	DatConsumptionMixH                    *float32           `xml:"DatConsumptionMixH,omitempty" json:"DatConsumptionMixH,omitempty"`
	WltpConsumptionBivalentMixedHMin      *float32           `xml:"WltpConsumptionBivalentMixedHMin,omitempty" json:"WltpConsumptionBivalentMixedHMin,omitempty"`
	DatWltpConsumptionBivalentMixedHMin   *float32           `xml:"DatWltpConsumptionBivalentMixedHMin,omitempty" json:"DatWltpConsumptionBivalentMixedHMin,omitempty"`
	WltpConsumptionBivalentMixedHMax      *float32           `xml:"WltpConsumptionBivalentMixedHMax,omitempty" json:"WltpConsumptionBivalentMixedHMax,omitempty"`
	DatWltpConsumptionBivalentMixedHMax   *float32           `xml:"DatWltpConsumptionBivalentMixedHMax,omitempty" json:"DatWltpConsumptionBivalentMixedHMax,omitempty"`
	Co2Emission                           *float32           `xml:"Co2Emission,omitempty" json:"Co2Emission,omitempty"`
	DatCo2Emission                        *float32           `xml:"DatCo2Emission,omitempty" json:"DatCo2Emission,omitempty"`
	WltpCo2EmissionMin                    *float32           `xml:"WltpCo2EmissionMin,omitempty" json:"WltpCo2EmissionMin,omitempty"`
	DatWltpCo2EmissionMin                 *float32           `xml:"DatWltpCo2EmissionMin,omitempty" json:"DatWltpCo2EmissionMin,omitempty"`
	WltpCo2EmissionMax                    *float32           `xml:"WltpCo2EmissionMax,omitempty" json:"WltpCo2EmissionMax,omitempty"`
	DatWltpCo2EmissionMax                 *float32           `xml:"DatWltpCo2EmissionMax,omitempty" json:"DatWltpCo2EmissionMax,omitempty"`
	EmissionClass                         string             `xml:"EmissionClass,omitempty" json:"EmissionClass,omitempty"`
	DatEmissionClass                      string             `xml:"DatEmissionClass,omitempty" json:"DatEmissionClass,omitempty"`
	Drive                                 string             `xml:"Drive,omitempty" json:"Drive,omitempty"`
	DatDrive                              string             `xml:"DatDrive,omitempty" json:"DatDrive,omitempty"`
	DriveN                                string             `xml:"DriveN,omitempty" json:"DriveN,omitempty"`
	DatDriveN                             string             `xml:"DatDriveN,omitempty" json:"DatDriveN,omitempty"`
	DriveCode                             string             `xml:"DriveCode,omitempty" json:"DriveCode,omitempty"`
	EngineCycle                           int64              `xml:"EngineCycle,omitempty" json:"EngineCycle,omitempty"`
	DatEngineCycle                        int64              `xml:"DatEngineCycle,omitempty" json:"DatEngineCycle,omitempty"`
	FuelMethod                            string             `xml:"FuelMethod,omitempty" json:"FuelMethod,omitempty"`
	DatFuelMethod                         string             `xml:"DatFuelMethod,omitempty" json:"DatFuelMethod,omitempty"`
	FuelMethodCode                        string             `xml:"FuelMethodCode,omitempty" json:"FuelMethodCode,omitempty"`
	FuelMethodType                        string             `xml:"FuelMethodType,omitempty" json:"FuelMethodType,omitempty"`
	DatFuelMethodType                     string             `xml:"DatFuelMethodType,omitempty" json:"DatFuelMethodType,omitempty"`
	UnloadedWeight                        int64              `xml:"UnloadedWeight,omitempty" json:"UnloadedWeight,omitempty"`
	DatUnloadedWeight                     int64              `xml:"DatUnloadedWeight,omitempty" json:"DatUnloadedWeight,omitempty"`
	PermissableTotalWeight                int64              `xml:"PermissableTotalWeight,omitempty" json:"PermissableTotalWeight,omitempty"`
	DatPermissableTotalWeight             int64              `xml:"DatPermissableTotalWeight,omitempty" json:"DatPermissableTotalWeight,omitempty"`
	Payload                               int64              `xml:"Payload,omitempty" json:"Payload,omitempty"`
	DatPayload                            int64              `xml:"DatPayload,omitempty" json:"DatPayload,omitempty"`
	LoadingLength                         int64              `xml:"LoadingLength,omitempty" json:"LoadingLength,omitempty"`
	DatLoadingLength                      int64              `xml:"DatLoadingLength,omitempty" json:"DatLoadingLength,omitempty"`
	LoadingWidth                          int64              `xml:"LoadingWidth,omitempty" json:"LoadingWidth,omitempty"`
	DatLoadingWidth                       int64              `xml:"DatLoadingWidth,omitempty" json:"DatLoadingWidth,omitempty"`
	LoadingHeight                         int64              `xml:"LoadingHeight,omitempty" json:"LoadingHeight,omitempty"`
	DatLoadingHeight                      int64              `xml:"DatLoadingHeight,omitempty" json:"DatLoadingHeight,omitempty"`
	LoadingSpace                          int64              `xml:"LoadingSpace,omitempty" json:"LoadingSpace,omitempty"`
	DatLoadingSpace                       int64              `xml:"DatLoadingSpace,omitempty" json:"DatLoadingSpace,omitempty"`
	LoadingSpaceMax                       int64              `xml:"LoadingSpaceMax,omitempty" json:"LoadingSpaceMax,omitempty"`
	DatLoadingSpaceMax                    int64              `xml:"DatLoadingSpaceMax,omitempty" json:"DatLoadingSpaceMax,omitempty"`
	UpperBodyMaterial                     string             `xml:"UpperBodyMaterial,omitempty" json:"UpperBodyMaterial,omitempty"`
	InsuranceTypeClassLiability           string             `xml:"InsuranceTypeClassLiability,omitempty" json:"InsuranceTypeClassLiability,omitempty"`
	InsuranceTypeClassCascoPartial        string             `xml:"InsuranceTypeClassCascoPartial,omitempty" json:"InsuranceTypeClassCascoPartial,omitempty"`
	InsuranceTypeClassCascoComplete       string             `xml:"InsuranceTypeClassCascoComplete,omitempty" json:"InsuranceTypeClassCascoComplete,omitempty"`
	DustBadge                             string             `xml:"DustBadge,omitempty" json:"DustBadge,omitempty"`
	ProductGroupName                      string             `xml:"ProductGroupName,omitempty" json:"ProductGroupName,omitempty"`
	EmissionKey                           string             `xml:"EmissionKey,omitempty" json:"EmissionKey,omitempty"`
	Built                                 string             `xml:"Built,omitempty" json:"Built,omitempty"`
	AllowedLoadCapacity                   int64              `xml:"AllowedLoadCapacity,omitempty" json:"AllowedLoadCapacity,omitempty"`
	CabinStructureAltDescription          string             `xml:"CabinStructureAltDescription,omitempty" json:"CabinStructureAltDescription,omitempty"`
	CushionColorId                        string             `xml:"CushionColorId,omitempty" json:"CushionColorId,omitempty"` //nolint
	FuelmethodAbbr                        string             `xml:"FuelmethodAbbr,omitempty" json:"FuelmethodAbbr,omitempty"`
	InsuranceTypeClassCascoCompleteNeu    string             `xml:"InsuranceTypeClassCascoCompleteNeu,omitempty" json:"InsuranceTypeClassCascoCompleteNeu,omitempty"`
	InsuranceTypeClassCascoPartialNeu     string             `xml:"InsuranceTypeClassCascoPartialNeu,omitempty" json:"InsuranceTypeClassCascoPartialNeu,omitempty"`
	InsuranceTypeClassLiabilityNew        string             `xml:"InsuranceTypeClassLiabilityNew,omitempty" json:"InsuranceTypeClassLiabilityNew,omitempty"`
	PayloadAlternative                    int64              `xml:"PayloadAlternative,omitempty" json:"PayloadAlternative,omitempty"`
	PowerKwSae                            *float32           `xml:"PowerKwSae,omitempty" json:"PowerKwSae,omitempty"`
	SommerSmogBadge                       string             `xml:"SommerSmogBadge,omitempty" json:"SommerSmogBadge,omitempty"`
	StowageMassFormat                     string             `xml:"StowageMassFormat,omitempty" json:"StowageMassFormat,omitempty"`
	TokenChangedCapacity                  string             `xml:"TokenChangedCapacity,omitempty" json:"TokenChangedCapacity,omitempty"`
	TokenTurboEngine                      string             `xml:"TokenTurboEngine,omitempty" json:"TokenTurboEngine,omitempty"`
	TypeOfTaxation                        string             `xml:"TypeOfTaxation,omitempty" json:"TypeOfTaxation,omitempty"`
	TypeSheetNumber                       string             `xml:"TypeSheetNumber,omitempty" json:"TypeSheetNumber,omitempty"`
	WhelBaseAlternative                   int64              `xml:"WhelBaseAlternative,omitempty" json:"WhelBaseAlternative,omitempty"`
	SuitableForE10                        *bool              `xml:"SuitableForE10,omitempty" json:"SuitableForE10,omitempty"`
	DatSuitableForE10                     *bool              `xml:"DatSuitableForE10,omitempty" json:"DatSuitableForE10,omitempty"`
	WeightTotalCombination                int64              `xml:"WeightTotalCombination,omitempty" json:"WeightTotalCombination,omitempty"`
	DatWeightTotalCombination             int64              `xml:"DatWeightTotalCombination,omitempty" json:"DatWeightTotalCombination,omitempty"`
	WidthForGarage                        int64              `xml:"WidthForGarage,omitempty" json:"WidthForGarage,omitempty"`
	DatWidthForGarage                     int64              `xml:"DatWidthForGarage,omitempty" json:"DatWidthForGarage,omitempty"`
	PowerKwSystem                         *float32           `xml:"PowerKwSystem,omitempty" json:"PowerKwSystem,omitempty"`
	DatPowerKwSystem                      *float32           `xml:"DatPowerKwSystem,omitempty" json:"DatPowerKwSystem,omitempty"`
	PowerHpSystem                         *float32           `xml:"PowerHpSystem,omitempty" json:"PowerHpSystem,omitempty"`
	DatPowerHpSystem                      *float32           `xml:"DatPowerHpSystem,omitempty" json:"DatPowerHpSystem,omitempty"`
	PowerKwPermanent                      *float32           `xml:"PowerKwPermanent,omitempty" json:"PowerKwPermanent,omitempty"`
	DatPowerKwPermanent                   *float32           `xml:"DatPowerKwPermanent,omitempty" json:"DatPowerKwPermanent,omitempty"`
	PowerHpPermanent                      *float32           `xml:"PowerHpPermanent,omitempty" json:"PowerHpPermanent,omitempty"`
	DatPowerHpPermanent                   *float32           `xml:"DatPowerHpPermanent,omitempty" json:"DatPowerHpPermanent,omitempty"`
	PowerKwMax                            *float32           `xml:"PowerKwMax,omitempty" json:"PowerKwMax,omitempty"`
	DatPowerKwMax                         *float32           `xml:"DatPowerKwMax,omitempty" json:"DatPowerKwMax,omitempty"`
	PowerHpMax                            *float32           `xml:"PowerHpMax,omitempty" json:"PowerHpMax,omitempty"`
	DatPowerHpMax                         *float32           `xml:"DatPowerHpMax,omitempty" json:"DatPowerHpMax,omitempty"`
	PowerKwPermanentSecondary             *float32           `xml:"PowerKwPermanentSecondary,omitempty" json:"PowerKwPermanentSecondary,omitempty"`
	DatPowerKwPermanentSecondary          *float32           `xml:"DatPowerKwPermanentSecondary,omitempty" json:"DatPowerKwPermanentSecondary,omitempty"`
	PowerHpPermanentSecondary             *float32           `xml:"PowerHpPermanentSecondary,omitempty" json:"PowerHpPermanentSecondary,omitempty"`
	DatPowerHpPermanentSecondary          *float32           `xml:"DatPowerHpPermanentSecondary,omitempty" json:"DatPowerHpPermanentSecondary,omitempty"`
	PowerKwMaxSecondary                   *float32           `xml:"PowerKwMaxSecondary,omitempty" json:"PowerKwMaxSecondary,omitempty"`
	DatPowerKwMaxSecondary                *float32           `xml:"DatPowerKwMaxSecondary,omitempty" json:"DatPowerKwMaxSecondary,omitempty"`
	PowerHpMaxSecondary                   *float32           `xml:"PowerHpMaxSecondary,omitempty" json:"PowerHpMaxSecondary,omitempty"`
	DatPowerHpMaxSecondary                *float32           `xml:"DatPowerHpMaxSecondary,omitempty" json:"DatPowerHpMaxSecondary,omitempty"`
	BatteryVoltage                        *float32           `xml:"BatteryVoltage,omitempty" json:"BatteryVoltage,omitempty"`
	DatBatteryVoltage                     *float32           `xml:"DatBatteryVoltage,omitempty" json:"DatBatteryVoltage,omitempty"`
	BatteryCapacity                       *float32           `xml:"BatteryCapacity,omitempty" json:"BatteryCapacity,omitempty"`
	DatBatteryCapacity                    *float32           `xml:"DatBatteryCapacity,omitempty" json:"DatBatteryCapacity,omitempty"`
	BatteryWeight                         *float32           `xml:"BatteryWeight,omitempty" json:"BatteryWeight,omitempty"`
	DatBatteryWeight                      *float32           `xml:"DatBatteryWeight,omitempty" json:"DatBatteryWeight,omitempty"`
	BatteryConstructionType               string             `xml:"BatteryConstructionType,omitempty" json:"BatteryConstructionType,omitempty"`
	DatBatteryConstructionType            string             `xml:"DatBatteryConstructionType,omitempty" json:"DatBatteryConstructionType,omitempty"`
	ChargingCurrentPlugType               string             `xml:"ChargingCurrentPlugType,omitempty" json:"ChargingCurrentPlugType,omitempty"`
	DatChargingCurrentPlugType            string             `xml:"DatChargingCurrentPlugType,omitempty" json:"DatChargingCurrentPlugType,omitempty"`
	PluginSystem                          *bool              `xml:"PluginSystem,omitempty" json:"PluginSystem,omitempty"`
	DatPluginSystem                       *bool              `xml:"DatPluginSystem,omitempty" json:"DatPluginSystem,omitempty"`
	QuickdropSystem                       *bool              `xml:"QuickdropSystem,omitempty" json:"QuickdropSystem,omitempty"`
	DatQuickdropSystem                    *bool              `xml:"DatQuickdropSystem,omitempty" json:"DatQuickdropSystem,omitempty"`
	NormalChargeVoltage                   int64              `xml:"NormalChargeVoltage,omitempty" json:"NormalChargeVoltage,omitempty"`
	DatNormalChargeVoltage                int64              `xml:"DatNormalChargeVoltage,omitempty" json:"DatNormalChargeVoltage,omitempty"`
	NormalChargeDuration                  *float32           `xml:"NormalChargeDuration,omitempty" json:"NormalChargeDuration,omitempty"`
	DatNormalChargeDuration               *float32           `xml:"DatNormalChargeDuration,omitempty" json:"DatNormalChargeDuration,omitempty"`
	QuickChargeVoltage                    int64              `xml:"QuickChargeVoltage,omitempty" json:"QuickChargeVoltage,omitempty"`
	DatQuickChargeVoltage                 int64              `xml:"DatQuickChargeVoltage,omitempty" json:"DatQuickChargeVoltage,omitempty"`
	QuickChargeDuration                   *float32           `xml:"QuickChargeDuration,omitempty" json:"QuickChargeDuration,omitempty"`
	DatQuickChargeDuration                *float32           `xml:"DatQuickChargeDuration,omitempty" json:"DatQuickChargeDuration,omitempty"`
	ConsumptionElectricalCurrent          *float32           `xml:"ConsumptionElectricalCurrent,omitempty" json:"ConsumptionElectricalCurrent,omitempty"`
	DatConsumptionElectricalCurrent       *float32           `xml:"DatConsumptionElectricalCurrent,omitempty" json:"DatConsumptionElectricalCurrent,omitempty"`
	WltpConsumptionElectricalMin          *float32           `xml:"WltpConsumptionElectricalMin,omitempty" json:"WltpConsumptionElectricalMin,omitempty"`
	DatWltpConsumptionElectricalMin       *float32           `xml:"DatWltpConsumptionElectricalMin,omitempty" json:"DatWltpConsumptionElectricalMin,omitempty"`
	WltpConsumptionElectricalMax          *float32           `xml:"WltpConsumptionElectricalMax,omitempty" json:"WltpConsumptionElectricalMax,omitempty"`
	DatWltpConsumptionElectricalMax       *float32           `xml:"DatWltpConsumptionElectricalMax,omitempty" json:"DatWltpConsumptionElectricalMax,omitempty"`
	RangeOfElectricMotor                  int64              `xml:"RangeOfElectricMotor,omitempty" json:"RangeOfElectricMotor,omitempty"`
	DatRangeOfElectricMotor               int64              `xml:"DatRangeOfElectricMotor,omitempty" json:"DatRangeOfElectricMotor,omitempty"`
	RangeTotal                            int64              `xml:"RangeTotal,omitempty" json:"RangeTotal,omitempty"`
	DatRangeTotal                         int64              `xml:"DatRangeTotal,omitempty" json:"DatRangeTotal,omitempty"`
	WltpRangeElectricalMin                int64              `xml:"WltpRangeElectricalMin,omitempty" json:"WltpRangeElectricalMin,omitempty"`
	DatWltpRangeElectricalMin             int64              `xml:"DatWltpRangeElectricalMin,omitempty" json:"DatWltpRangeElectricalMin,omitempty"`
	WltpRangeElectricalMax                int64              `xml:"WltpRangeElectricalMax,omitempty" json:"WltpRangeElectricalMax,omitempty"`
	DatWltpRangeElectricalMax             int64              `xml:"DatWltpRangeElectricalMax,omitempty" json:"DatWltpRangeElectricalMax,omitempty"`
	WltpRangeTotalMin                     int64              `xml:"WltpRangeTotalMin,omitempty" json:"WltpRangeTotalMin,omitempty"`
	DatWltpRangeTotalMin                  int64              `xml:"DatWltpRangeTotalMin,omitempty" json:"DatWltpRangeTotalMin,omitempty"`
	WltpRangeTotalMax                     int64              `xml:"WltpRangeTotalMax,omitempty" json:"WltpRangeTotalMax,omitempty"`
	DatWltpRangeTotalMax                  int64              `xml:"DatWltpRangeTotalMax,omitempty" json:"DatWltpRangeTotalMax,omitempty"`
	EnergyEfficiencyClass                 string             `xml:"EnergyEfficiencyClass,omitempty" json:"EnergyEfficiencyClass,omitempty"`
	DatEnergyEfficiencyClass              string             `xml:"DatEnergyEfficiencyClass,omitempty" json:"DatEnergyEfficiencyClass,omitempty"`
	ModelTypecode                         string             `xml:"ModelTypecode,omitempty" json:"ModelTypecode,omitempty"`
	ModelVariant                          string             `xml:"ModelVariant,omitempty" json:"ModelVariant,omitempty"`
	Type                                  string             `xml:"Type,omitempty" json:"Type,omitempty"`
	TypeVariant                           string             `xml:"TypeVariant,omitempty" json:"TypeVariant,omitempty"`
	EngineType                            string             `xml:"EngineType,omitempty" json:"EngineType,omitempty"`
	SpecialModel                          string             `xml:"SpecialModel,omitempty" json:"SpecialModel,omitempty"`
	TechInfoWltp                          *TechInfoWltp      `xml:"TechInfoWltp,omitempty" json:"TechInfoWltp,omitempty"`
}

type TechInfoWltp struct {
	XMLName                                   xml.Name `xml:"http://www.dat.de/vxs TechInfoWltp"`
	WltpConsumptionLowMin                     *float32 `xml:"WltpConsumptionLowMin,omitempty" json:"WltpConsumptionLowMin,omitempty"`
	DatWltpConsumptionLowMin                  *float32 `xml:"DatWltpConsumptionLowMin,omitempty" json:"DatWltpConsumptionLowMin,omitempty"`
	WltpConsumptionLowMax                     *float32 `xml:"WltpConsumptionLowMax,omitempty" json:"WltpConsumptionLowMax,omitempty"`
	DatWltpConsumptionLowMax                  *float32 `xml:"DatWltpConsumptionLowMax,omitempty" json:"DatWltpConsumptionLowMax,omitempty"`
	WltpConsumptionMediumMin                  *float32 `xml:"WltpConsumptionMediumMin,omitempty" json:"WltpConsumptionMediumMin,omitempty"`
	DatWltpConsumptionMediumMin               *float32 `xml:"DatWltpConsumptionMediumMin,omitempty" json:"DatWltpConsumptionMediumMin,omitempty"`
	WltpConsumptionMediumMax                  *float32 `xml:"WltpConsumptionMediumMax,omitempty" json:"WltpConsumptionMediumMax,omitempty"`
	DatWltpConsumptionMediumMax               *float32 `xml:"DatWltpConsumptionMediumMax,omitempty" json:"DatWltpConsumptionMediumMax,omitempty"`
	WltpConsumptionHighMin                    *float32 `xml:"WltpConsumptionHighMin,omitempty" json:"WltpConsumptionHighMin,omitempty"`
	DatWltpConsumptionHighMin                 *float32 `xml:"DatWltpConsumptionHighMin,omitempty" json:"DatWltpConsumptionHighMin,omitempty"`
	WltpConsumptionHighMax                    *float32 `xml:"WltpConsumptionHighMax,omitempty" json:"WltpConsumptionHighMax,omitempty"`
	DatWltpConsumptionHighMax                 *float32 `xml:"DatWltpConsumptionHighMax,omitempty" json:"DatWltpConsumptionHighMax,omitempty"`
	WltpConsumptionExtraHighMin               *float32 `xml:"WltpConsumptionExtraHighMin,omitempty" json:"WltpConsumptionExtraHighMin,omitempty"`
	DatWltpConsumptionExtraHighMin            *float32 `xml:"DatWltpConsumptionExtraHighMin,omitempty" json:"DatWltpConsumptionExtraHighMin,omitempty"`
	WltpConsumptionExtraHighMax               *float32 `xml:"WltpConsumptionExtraHighMax,omitempty" json:"WltpConsumptionExtraHighMax,omitempty"`
	DatWltpConsumptionExtraHighMax            *float32 `xml:"DatWltpConsumptionExtraHighMax,omitempty" json:"DatWltpConsumptionExtraHighMax,omitempty"`
	WltpConsumptionMixedMin                   *float32 `xml:"WltpConsumptionMixedMin,omitempty" json:"WltpConsumptionMixedMin,omitempty"`
	DatWltpConsumptionMixedMin                *float32 `xml:"DatWltpConsumptionMixedMin,omitempty" json:"DatWltpConsumptionMixedMin,omitempty"`
	WltpConsumptionMixedMax                   *float32 `xml:"WltpConsumptionMixedMax,omitempty" json:"WltpConsumptionMixedMax,omitempty"`
	DatWltpConsumptionMixedMax                *float32 `xml:"DatWltpConsumptionMixedMax,omitempty" json:"DatWltpConsumptionMixedMax,omitempty"`
	WltpConsumptionBivalentLowCngMin          *float32 `xml:"WltpConsumptionBivalentLowCngMin,omitempty" json:"WltpConsumptionBivalentLowCngMin,omitempty"`
	DatWltpConsumptionBivalentLowCngMin       *float32 `xml:"DatWltpConsumptionBivalentLowCngMin,omitempty" json:"DatWltpConsumptionBivalentLowCngMin,omitempty"`
	WltpConsumptionBivalentLowCngMax          *float32 `xml:"WltpConsumptionBivalentLowCngMax,omitempty" json:"WltpConsumptionBivalentLowCngMax,omitempty"`
	DatWltpConsumptionBivalentLowCngMax       *float32 `xml:"DatWltpConsumptionBivalentLowCngMax,omitempty" json:"DatWltpConsumptionBivalentLowCngMax,omitempty"`
	WltpConsumptionBivalentMediumCngMin       *float32 `xml:"WltpConsumptionBivalentMediumCngMin,omitempty" json:"WltpConsumptionBivalentMediumCngMin,omitempty"`
	DatWltpConsumptionBivalentMediumCngMin    *float32 `xml:"DatWltpConsumptionBivalentMediumCngMin,omitempty" json:"DatWltpConsumptionBivalentMediumCngMin,omitempty"`
	WltpConsumptionBivalentMediumCngMax       *float32 `xml:"WltpConsumptionBivalentMediumCngMax,omitempty" json:"WltpConsumptionBivalentMediumCngMax,omitempty"`
	DatWltpConsumptionBivalentMediumCngMax    *float32 `xml:"DatWltpConsumptionBivalentMediumCngMax,omitempty" json:"DatWltpConsumptionBivalentMediumCngMax,omitempty"`
	WltpConsumptionBivalentHighCngMin         *float32 `xml:"WltpConsumptionBivalentHighCngMin,omitempty" json:"WltpConsumptionBivalentHighCngMin,omitempty"`
	DatWltpConsumptionBivalentHighCngMin      *float32 `xml:"DatWltpConsumptionBivalentHighCngMin,omitempty" json:"DatWltpConsumptionBivalentHighCngMin,omitempty"`
	WltpConsumptionBivalentHighCngMax         *float32 `xml:"WltpConsumptionBivalentHighCngMax,omitempty" json:"WltpConsumptionBivalentHighCngMax,omitempty"`
	DatWltpConsumptionBivalentHighCngMax      *float32 `xml:"DatWltpConsumptionBivalentHighCngMax,omitempty" json:"DatWltpConsumptionBivalentHighCngMax,omitempty"`
	WltpConsumptionBivalentExtraHighCngMin    *float32 `xml:"WltpConsumptionBivalentExtraHighCngMin,omitempty" json:"WltpConsumptionBivalentExtraHighCngMin,omitempty"`
	DatWltpConsumptionBivalentExtraHighCngMin *float32 `xml:"DatWltpConsumptionBivalentExtraHighCngMin,omitempty" json:"DatWltpConsumptionBivalentExtraHighCngMin,omitempty"`
	WltpConsumptionBivalentExtraHighCngMax    *float32 `xml:"WltpConsumptionBivalentExtraHighCngMax,omitempty" json:"WltpConsumptionBivalentExtraHighCngMax,omitempty"`
	DatWltpConsumptionBivalentExtraHighCngMax *float32 `xml:"DatWltpConsumptionBivalentExtraHighCngMax,omitempty" json:"DatWltpConsumptionBivalentExtraHighCngMax,omitempty"`
	WltpConsumptionBivalentMixedCngMin        *float32 `xml:"WltpConsumptionBivalentMixedCngMin,omitempty" json:"WltpConsumptionBivalentMixedCngMin,omitempty"`
	DatWltpConsumptionBivalentMixedCngMin     *float32 `xml:"DatWltpConsumptionBivalentMixedCngMin,omitempty" json:"DatWltpConsumptionBivalentMixedCngMin,omitempty"`
	WltpConsumptionBivalentMixedCngMax        *float32 `xml:"WltpConsumptionBivalentMixedCngMax,omitempty" json:"WltpConsumptionBivalentMixedCngMax,omitempty"`
	DatWltpConsumptionBivalentMixedCngMax     *float32 `xml:"DatWltpConsumptionBivalentMixedCngMax,omitempty" json:"DatWltpConsumptionBivalentMixedCngMax,omitempty"`
	WltpConsumptionBivalentLowLpgMin          *float32 `xml:"WltpConsumptionBivalentLowLpgMin,omitempty" json:"WltpConsumptionBivalentLowLpgMin,omitempty"`
	DatWltpConsumptionBivalentLowLpgMin       *float32 `xml:"DatWltpConsumptionBivalentLowLpgMin,omitempty" json:"DatWltpConsumptionBivalentLowLpgMin,omitempty"`
	WltpConsumptionBivalentLowLpgMax          *float32 `xml:"WltpConsumptionBivalentLowLpgMax,omitempty" json:"WltpConsumptionBivalentLowLpgMax,omitempty"`
	DatWltpConsumptionBivalentLowLpgMax       *float32 `xml:"DatWltpConsumptionBivalentLowLpgMax,omitempty" json:"DatWltpConsumptionBivalentLowLpgMax,omitempty"`
	WltpConsumptionBivalentMediumLpgMin       *float32 `xml:"WltpConsumptionBivalentMediumLpgMin,omitempty" json:"WltpConsumptionBivalentMediumLpgMin,omitempty"`
	DatWltpConsumptionBivalentMediumLpgMin    *float32 `xml:"DatWltpConsumptionBivalentMediumLpgMin,omitempty" json:"DatWltpConsumptionBivalentMediumLpgMin,omitempty"`
	WltpConsumptionBivalentMediumLpgMax       *float32 `xml:"WltpConsumptionBivalentMediumLpgMax,omitempty" json:"WltpConsumptionBivalentMediumLpgMax,omitempty"`
	DatWltpConsumptionBivalentMediumLpgMax    *float32 `xml:"DatWltpConsumptionBivalentMediumLpgMax,omitempty" json:"DatWltpConsumptionBivalentMediumLpgMax,omitempty"`
	WltpConsumptionBivalentHighLpgMin         *float32 `xml:"WltpConsumptionBivalentHighLpgMin,omitempty" json:"WltpConsumptionBivalentHighLpgMin,omitempty"`
	DatWltpConsumptionBivalentHighLpgMin      *float32 `xml:"DatWltpConsumptionBivalentHighLpgMin,omitempty" json:"DatWltpConsumptionBivalentHighLpgMin,omitempty"`
	WltpConsumptionBivalentHighLpgMax         *float32 `xml:"WltpConsumptionBivalentHighLpgMax,omitempty" json:"WltpConsumptionBivalentHighLpgMax,omitempty"`
	DatWltpConsumptionBivalentHighLpgMax      *float32 `xml:"DatWltpConsumptionBivalentHighLpgMax,omitempty" json:"DatWltpConsumptionBivalentHighLpgMax,omitempty"`
	WltpConsumptionBivalentExtraHighLpgMin    *float32 `xml:"WltpConsumptionBivalentExtraHighLpgMin,omitempty" json:"WltpConsumptionBivalentExtraHighLpgMin,omitempty"`
	DatWltpConsumptionBivalentExtraHighLpgMin *float32 `xml:"DatWltpConsumptionBivalentExtraHighLpgMin,omitempty" json:"DatWltpConsumptionBivalentExtraHighLpgMin,omitempty"`
	WltpConsumptionBivalentExtraHighLpgMax    *float32 `xml:"WltpConsumptionBivalentExtraHighLpgMax,omitempty" json:"WltpConsumptionBivalentExtraHighLpgMax,omitempty"`
	DatWltpConsumptionBivalentExtraHighLpgMax *float32 `xml:"DatWltpConsumptionBivalentExtraHighLpgMax,omitempty" json:"DatWltpConsumptionBivalentExtraHighLpgMax,omitempty"`
	WltpConsumptionBivalentMixedLpgMin        *float32 `xml:"WltpConsumptionBivalentMixedLpgMin,omitempty" json:"WltpConsumptionBivalentMixedLpgMin,omitempty"`
	DatWltpConsumptionBivalentMixedLpgMin     *float32 `xml:"DatWltpConsumptionBivalentMixedLpgMin,omitempty" json:"DatWltpConsumptionBivalentMixedLpgMin,omitempty"`
	WltpConsumptionBivalentMixedLpgMax        *float32 `xml:"WltpConsumptionBivalentMixedLpgMax,omitempty" json:"WltpConsumptionBivalentMixedLpgMax,omitempty"`
	DatWltpConsumptionBivalentMixedLpgMax     *float32 `xml:"DatWltpConsumptionBivalentMixedLpgMax,omitempty" json:"DatWltpConsumptionBivalentMixedLpgMax,omitempty"`
	WltpConsumptionBivalentLowHMin            *float32 `xml:"WltpConsumptionBivalentLowHMin,omitempty" json:"WltpConsumptionBivalentLowHMin,omitempty"`
	DatWltpConsumptionBivalentLowHMin         *float32 `xml:"DatWltpConsumptionBivalentLowHMin,omitempty" json:"DatWltpConsumptionBivalentLowHMin,omitempty"`
	WltpConsumptionBivalentLowHMax            *float32 `xml:"WltpConsumptionBivalentLowHMax,omitempty" json:"WltpConsumptionBivalentLowHMax,omitempty"`
	DatWltpConsumptionBivalentLowHMax         *float32 `xml:"DatWltpConsumptionBivalentLowHMax,omitempty" json:"DatWltpConsumptionBivalentLowHMax,omitempty"`
	WltpConsumptionBivalentMediumHMin         *float32 `xml:"WltpConsumptionBivalentMediumHMin,omitempty" json:"WltpConsumptionBivalentMediumHMin,omitempty"`
	DatWltpConsumptionBivalentMediumHMin      *float32 `xml:"DatWltpConsumptionBivalentMediumHMin,omitempty" json:"DatWltpConsumptionBivalentMediumHMin,omitempty"`
	WltpConsumptionBivalentMediumHMax         *float32 `xml:"WltpConsumptionBivalentMediumHMax,omitempty" json:"WltpConsumptionBivalentMediumHMax,omitempty"`
	DatWltpConsumptionBivalentMediumHMax      *float32 `xml:"DatWltpConsumptionBivalentMediumHMax,omitempty" json:"DatWltpConsumptionBivalentMediumHMax,omitempty"`
	WltpConsumptionBivalentHighHMin           *float32 `xml:"WltpConsumptionBivalentHighHMin,omitempty" json:"WltpConsumptionBivalentHighHMin,omitempty"`
	DatWltpConsumptionBivalentHighHMin        *float32 `xml:"DatWltpConsumptionBivalentHighHMin,omitempty" json:"DatWltpConsumptionBivalentHighHMin,omitempty"`
	WltpConsumptionBivalentHighHMax           *float32 `xml:"WltpConsumptionBivalentHighHMax,omitempty" json:"WltpConsumptionBivalentHighHMax,omitempty"`
	DatWltpConsumptionBivalentHighHMax        *float32 `xml:"DatWltpConsumptionBivalentHighHMax,omitempty" json:"DatWltpConsumptionBivalentHighHMax,omitempty"`
	WltpConsumptionBivalentExtraHighHMin      *float32 `xml:"WltpConsumptionBivalentExtraHighHMin,omitempty" json:"WltpConsumptionBivalentExtraHighHMin,omitempty"`
	DatWltpConsumptionBivalentExtraHighHMin   *float32 `xml:"DatWltpConsumptionBivalentExtraHighHMin,omitempty" json:"DatWltpConsumptionBivalentExtraHighHMin,omitempty"`
	WltpConsumptionBivalentExtraHighHMax      *float32 `xml:"WltpConsumptionBivalentExtraHighHMax,omitempty" json:"WltpConsumptionBivalentExtraHighHMax,omitempty"`
	DatWltpConsumptionBivalentExtraHighHMax   *float32 `xml:"DatWltpConsumptionBivalentExtraHighHMax,omitempty" json:"DatWltpConsumptionBivalentExtraHighHMax,omitempty"`
	WltpConsumptionBivalentMixedHMin          *float32 `xml:"WltpConsumptionBivalentMixedHMin,omitempty" json:"WltpConsumptionBivalentMixedHMin,omitempty"`
	DatWltpConsumptionBivalentMixedHMin       *float32 `xml:"DatWltpConsumptionBivalentMixedHMin,omitempty" json:"DatWltpConsumptionBivalentMixedHMin,omitempty"`
	WltpConsumptionBivalentMixedHMax          *float32 `xml:"WltpConsumptionBivalentMixedHMax,omitempty" json:"WltpConsumptionBivalentMixedHMax,omitempty"`
	DatWltpConsumptionBivalentMixedHMax       *float32 `xml:"DatWltpConsumptionBivalentMixedHMax,omitempty" json:"DatWltpConsumptionBivalentMixedHMax,omitempty"`
	WltpCo2EmissionMin                        *float32 `xml:"WltpCo2EmissionMin,omitempty" json:"WltpCo2EmissionMin,omitempty"`
	DatWltpCo2EmissionMin                     *float32 `xml:"DatWltpCo2EmissionMin,omitempty" json:"DatWltpCo2EmissionMin,omitempty"`
	WltpCo2EmissionMax                        *float32 `xml:"WltpCo2EmissionMax,omitempty" json:"WltpCo2EmissionMax,omitempty"`
	DatWltpCo2EmissionMax                     *float32 `xml:"DatWltpCo2EmissionMax,omitempty" json:"DatWltpCo2EmissionMax,omitempty"`
	WltpConsumptionElectricalMin              *float32 `xml:"WltpConsumptionElectricalMin,omitempty" json:"WltpConsumptionElectricalMin,omitempty"`
	DatWltpConsumptionElectricalMin           *float32 `xml:"DatWltpConsumptionElectricalMin,omitempty" json:"DatWltpConsumptionElectricalMin,omitempty"`
	WltpConsumptionElectricalMax              *float32 `xml:"WltpConsumptionElectricalMax,omitempty" json:"WltpConsumptionElectricalMax,omitempty"`
	DatWltpConsumptionElectricalMax           *float32 `xml:"DatWltpConsumptionElectricalMax,omitempty" json:"DatWltpConsumptionElectricalMax,omitempty"`
	WltpRangeElectricalMin                    int64    `xml:"WltpRangeElectricalMin,omitempty" json:"WltpRangeElectricalMin,omitempty"`
	DatWltpRangeElectricalMin                 int64    `xml:"DatWltpRangeElectricalMin,omitempty" json:"DatWltpRangeElectricalMin,omitempty"`
	WltpRangeElectricalMax                    int64    `xml:"WltpRangeElectricalMax,omitempty" json:"WltpRangeElectricalMax,omitempty"`
	DatWltpRangeElectricalMax                 int64    `xml:"DatWltpRangeElectricalMax,omitempty" json:"DatWltpRangeElectricalMax,omitempty"`
	WltpRangeTotalMin                         int64    `xml:"WltpRangeTotalMin,omitempty" json:"WltpRangeTotalMin,omitempty"`
	DatWltpRangeTotalMin                      int64    `xml:"DatWltpRangeTotalMin,omitempty" json:"DatWltpRangeTotalMin,omitempty"`
	WltpRangeTotalMax                         int64    `xml:"WltpRangeTotalMax,omitempty" json:"WltpRangeTotalMax,omitempty"`
	DatWltpRangeTotalMax                      int64    `xml:"DatWltpRangeTotalMax,omitempty" json:"DatWltpRangeTotalMax,omitempty"`
}

type Engine struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Engine"`

	EngingeType                          string   `xml:"EngingeType,omitempty" json:"EngingeType,omitempty"`
	EngineType                           string   `xml:"EngineType,omitempty" json:"EngineType,omitempty"`
	CatalyticConverterType               string   `xml:"CatalyticConverterType,omitempty" json:"CatalyticConverterType,omitempty"`
	GearType                             string   `xml:"GearType,omitempty" json:"GearType,omitempty"`
	FuelMethod                           string   `xml:"FuelMethod,omitempty" json:"FuelMethod,omitempty"`
	DatFuelMethod                        string   `xml:"DatFuelMethod,omitempty" json:"DatFuelMethod,omitempty"`
	EnginePowerKw                        int64    `xml:"EnginePowerKw,omitempty" json:"EnginePowerKw,omitempty"`
	DatEnginePowerKw                     int64    `xml:"DatEnginePowerKw,omitempty" json:"DatEnginePowerKw,omitempty"`
	EnginePowerHp                        int64    `xml:"EnginePowerHp,omitempty" json:"EnginePowerHp,omitempty"`
	DatEnginePowerHp                     int64    `xml:"DatEnginePowerHp,omitempty" json:"DatEnginePowerHp,omitempty"`
	Cylinders                            int64    `xml:"Cylinders,omitempty" json:"Cylinders,omitempty"`
	DatCylinders                         int64    `xml:"DatCylinders,omitempty" json:"DatCylinders,omitempty"`
	Capacity                             int64    `xml:"Capacity,omitempty" json:"Capacity,omitempty"`
	DatCapacity                          int64    `xml:"DatCapacity,omitempty" json:"DatCapacity,omitempty"`
	PollutionClass                       string   `xml:"PollutionClass,omitempty" json:"PollutionClass,omitempty"`
	Consumption                          *float32 `xml:"Consumption,omitempty" json:"Consumption,omitempty"`
	ConsumptionInTown                    *float32 `xml:"ConsumptionInTown,omitempty" json:"ConsumptionInTown,omitempty"`
	ConsumptionOutOfTown                 *float32 `xml:"ConsumptionOutOfTown,omitempty" json:"ConsumptionOutOfTown,omitempty"`
	Co2Emission                          *float32 `xml:"Co2Emission,omitempty" json:"Co2Emission,omitempty"`
	DirectInjection                      string   `xml:"DirectInjection,omitempty" json:"DirectInjection,omitempty"`
	EngineClass                          string   `xml:"EngineClass,omitempty" json:"EngineClass,omitempty"`
	EnginePowerHpManufacturerInformation *float32 `xml:"EnginePowerHpManufacturerInformation,omitempty" json:"EnginePowerHpManufacturerInformation,omitempty"`
	PowerKwPsManual                      string   `xml:"PowerKwPsManual,omitempty" json:"PowerKwPsManual,omitempty"`
}

type OriginalPriceInfo struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs OriginalPriceInfo"`

	OriginalPriceNet         *float32 `xml:"OriginalPriceNet,omitempty" json:"OriginalPriceNet,omitempty"`
	OriginalPriceVATRate     *float32 `xml:"OriginalPriceVATRate,omitempty" json:"OriginalPriceVATRate,omitempty"`
	OriginalPriceNoVA        *float32 `xml:"OriginalPriceNoVA,omitempty" json:"OriginalPriceNoVA,omitempty"`
	OriginalPriceNoVARate    *float32 `xml:"OriginalPriceNoVARate,omitempty" json:"OriginalPriceNoVARate,omitempty"`
	DatOriginalPriceNoVARate *float32 `xml:"DatOriginalPriceNoVARate,omitempty" json:"DatOriginalPriceNoVARate,omitempty"`
	OriginalPriceBonus       *float32 `xml:"OriginalPriceBonus,omitempty" json:"OriginalPriceBonus,omitempty"`
	OriginalPriceMalus       *float32 `xml:"OriginalPriceMalus,omitempty" json:"OriginalPriceMalus,omitempty"`
	RegistrationTaxRate      *float32 `xml:"RegistrationTaxRate,omitempty" json:"RegistrationTaxRate,omitempty"`
	RegistrationTax          *float32 `xml:"RegistrationTax,omitempty" json:"RegistrationTax,omitempty"`
	TransportationCosts      *float32 `xml:"TransportationCosts,omitempty" json:"TransportationCosts,omitempty"`
	OriginalPriceGross       *float32 `xml:"OriginalPriceGross,omitempty" json:"OriginalPriceGross,omitempty"`
}

type Equipment struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Equipment"`

	ColorType                      string   `xml:"ColorType,omitempty" json:"ColorType,omitempty"`
	DatColorType                   string   `xml:"DatColorType,omitempty" json:"DatColorType,omitempty"`
	Color                          string   `xml:"Color,omitempty" json:"Color,omitempty"`
	DatColor                       string   `xml:"DatColor,omitempty" json:"DatColor,omitempty"`
	ColorCodeFromVin               string   `xml:"ColorCodeFromVin,omitempty" json:"ColorCodeFromVin,omitempty"`
	ColorVariant                   string   `xml:"ColorVariant,omitempty" json:"ColorVariant,omitempty"`
	DatColorVariant                string   `xml:"DatColorVariant,omitempty" json:"DatColorVariant,omitempty"`
	LacquerType                    string   `xml:"LacquerType,omitempty" json:"LacquerType,omitempty"`
	DatLacquerType                 string   `xml:"DatLacquerType,omitempty" json:"DatLacquerType,omitempty"`
	CushionType                    string   `xml:"CushionType,omitempty" json:"CushionType,omitempty"`
	DatCushionType                 string   `xml:"DatCushionType,omitempty" json:"DatCushionType,omitempty"`
	CushionTypeName                string   `xml:"CushionTypeName,omitempty" json:"CushionTypeName,omitempty"`
	DatCushionTypeName             string   `xml:"DatCushionTypeName,omitempty" json:"DatCushionTypeName,omitempty"`
	CushionColorType               string   `xml:"CushionColorType,omitempty" json:"CushionColorType,omitempty"`
	DatCushionColorType            string   `xml:"DatCushionColorType,omitempty" json:"DatCushionColorType,omitempty"`
	CushionColor                   string   `xml:"CushionColor,omitempty" json:"CushionColor,omitempty"`
	DatCushionColor                string   `xml:"DatCushionColor,omitempty" json:"DatCushionColor,omitempty"`
	EquipmentValue                 *float32 `xml:"EquipmentValue,omitempty" json:"EquipmentValue,omitempty"`
	EquipmentValueGross            *float32 `xml:"EquipmentValueGross,omitempty" json:"EquipmentValueGross,omitempty"`
	DatEquipmentValue              *float32 `xml:"DatEquipmentValue,omitempty" json:"DatEquipmentValue,omitempty"`
	DatEquipmentValueGross         *float32 `xml:"DatEquipmentValueGross,omitempty" json:"DatEquipmentValueGross,omitempty"`
	OriginalEquipmentValue         *float32 `xml:"OriginalEquipmentValue,omitempty" json:"OriginalEquipmentValue,omitempty"`
	OriginalEquipmentValueGross    *float32 `xml:"OriginalEquipmentValueGross,omitempty" json:"OriginalEquipmentValueGross,omitempty"`
	DatOriginalEquipmentValue      *float32 `xml:"DatOriginalEquipmentValue,omitempty" json:"DatOriginalEquipmentValue,omitempty"`
	DatOriginalEquipmentValueGross *float32 `xml:"DatOriginalEquipmentValueGross,omitempty" json:"DatOriginalEquipmentValueGross,omitempty"`
	EquipmentValueType             string   `xml:"EquipmentValueType,omitempty" json:"EquipmentValueType,omitempty"`
	SpecialEditionPackageId        int64    `xml:"SpecialEditionPackageId,omitempty" json:"SpecialEditionPackageId,omitempty"`
	SpecialEditionPackageName      string   `xml:"SpecialEditionPackageName,omitempty" json:"SpecialEditionPackageName,omitempty"`
	SpecialEditionPackageNameN     string   `xml:"SpecialEditionPackageNameN,omitempty" json:"SpecialEditionPackageNameN,omitempty"`
	SpecialEditionPackageDetails1  string   `xml:"SpecialEditionPackageDetails1,omitempty" json:"SpecialEditionPackageDetails1,omitempty"`
	SpecialEditionPackageDetails2  string   `xml:"SpecialEditionPackageDetails2,omitempty" json:"SpecialEditionPackageDetails2,omitempty"`
	// SeriesEquipment                *EquipSequence `xml:"SeriesEquipment,omitempty" json:"SeriesEquipment,omitempty"`
	// DeselectedSeriesEquipment      *EquipSequence `xml:"DeselectedSeriesEquipment,omitempty" json:"DeselectedSeriesEquipment,omitempty"`
	// SpecialModelEquipment          *EquipSequence `xml:"SpecialModelEquipment,omitempty" json:"SpecialModelEquipment,omitempty"`
	// SpecialEquipment               *EquipSequence `xml:"SpecialEquipment,omitempty" json:"SpecialEquipment,omitempty"`
	// SeriesOrSpecialEquipment       *EquipSequence `xml:"SeriesOrSpecialEquipment,omitempty" json:"SeriesOrSpecialEquipment,omitempty"`
	// FreeSpecialEquipment           *EquipSequence `xml:"FreeSpecialEquipment,omitempty" json:"FreeSpecialEquipment,omitempty"`
	// AdditionalEquipment            *EquipSequence `xml:"AdditionalEquipment,omitempty" json:"AdditionalEquipment,omitempty"`
	// FlatRateEquipment              *EquipSequence `xml:"FlatRateEquipment,omitempty" json:"FlatRateEquipment,omitempty"`
	// DenialCaseEquipment            *EquipSequence `xml:"DenialCaseEquipment,omitempty" json:"DenialCaseEquipment,omitempty"`
}

type EquipSequence struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs equipSequence"`

	EquipmentPosition []*EquipmentPosition `xml:"EquipmentPosition,omitempty" json:"EquipmentPosition,omitempty"`
}

type EquipmentPosition struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs EquipmentPosition"`

	AgeInMonths                 int64    `xml:"AgeInMonths,omitempty" json:"AgeInMonths,omitempty"`
	DatAgeInMonths              int64    `xml:"DatAgeInMonths,omitempty" json:"DatAgeInMonths,omitempty"`
	Deselected                  *bool    `xml:"Deselected,omitempty" json:"Deselected,omitempty"`
	DatEquipmentId              int64    `xml:"DatEquipmentId,omitempty" json:"DatEquipmentId,omitempty"` //nolint
	ManufacturerEquipmentId     string   `xml:"ManufacturerEquipmentId,omitempty" json:"ManufacturerEquipmentId,omitempty"`
	ManufacturerDescription     string   `xml:"ManufacturerDescription,omitempty" json:"ManufacturerDescription,omitempty"`
	ValuationControlType        string   `xml:"ValuationControlType,omitempty" json:"ValuationControlType,omitempty"`
	Description                 string   `xml:"Description,omitempty" json:"Description,omitempty"`
	LongDescription             string   `xml:"LongDescription,omitempty" json:"LongDescription,omitempty"`
	FootnoteType                string   `xml:"FootnoteType,omitempty" json:"FootnoteType,omitempty"`
	FootnotePerc                *float32 `xml:"FootnotePerc,omitempty" json:"FootnotePerc,omitempty"`
	DatFootnotePerc             *float32 `xml:"DatFootnotePerc,omitempty" json:"DatFootnotePerc,omitempty"`
	DecreaseType                string   `xml:"DecreaseType,omitempty" json:"DecreaseType,omitempty"`
	DatDecreaseType             string   `xml:"DatDecreaseType,omitempty" json:"DatDecreaseType,omitempty"`
	PercentageOfBasePrice       int64    `xml:"PercentageOfBasePrice,omitempty" json:"PercentageOfBasePrice,omitempty"`
	OriginalPrice               *float32 `xml:"OriginalPrice,omitempty" json:"OriginalPrice,omitempty"`
	OriginalPriceGross          *float32 `xml:"OriginalPriceGross,omitempty" json:"OriginalPriceGross,omitempty"`
	OriginalPriceUser           *float32 `xml:"OriginalPriceUser,omitempty" json:"OriginalPriceUser,omitempty"`
	OriginalPriceGrossUser      *float32 `xml:"OriginalPriceGrossUser,omitempty" json:"OriginalPriceGrossUser,omitempty"`
	DatResidualValue            *float32 `xml:"DatResidualValue,omitempty" json:"DatResidualValue,omitempty"`
	DatResidualValueGross       *float32 `xml:"DatResidualValueGross,omitempty" json:"DatResidualValueGross,omitempty"`
	ResidualValue               *float32 `xml:"ResidualValue,omitempty" json:"ResidualValue,omitempty"`
	ResidualValueGross          *float32 `xml:"ResidualValueGross,omitempty" json:"ResidualValueGross,omitempty"`
	Amount                      int64    `xml:"Amount,omitempty" json:"Amount,omitempty"`
	EquipmentGroup              string   `xml:"EquipmentGroup,omitempty" json:"EquipmentGroup,omitempty"`
	EquipmentType               string   `xml:"EquipmentType,omitempty" json:"EquipmentType,omitempty"`
	Category                    string   `xml:"Category,omitempty" json:"Category,omitempty"`
	ManualEntry                 *bool    `xml:"ManualEntry,omitempty" json:"ManualEntry,omitempty"`
	ManualAgeEntry              *bool    `xml:"ManualAgeEntry,omitempty" json:"ManualAgeEntry,omitempty"`
	EquipmentClass              int64    `xml:"EquipmentClass,omitempty" json:"EquipmentClass,omitempty"`
	ConstructionTimeFrom        int64    `xml:"ConstructionTimeFrom,omitempty" json:"ConstructionTimeFrom,omitempty"`
	SeriesEquipmentMissing      *bool    `xml:"SeriesEquipmentMissing,omitempty" json:"SeriesEquipmentMissing,omitempty"`
	PackageEquipmentId          int64    `xml:"PackageEquipmentId,omitempty" json:"PackageEquipmentId,omitempty"`
	GearBoxType                 string   `xml:"GearBoxType,omitempty" json:"GearBoxType,omitempty"`
	NrOfGears                   string   `xml:"NrOfGears,omitempty" json:"NrOfGears,omitempty"`
	AddedByLogikCheck           *bool    `xml:"AddedByLogikCheck,omitempty" json:"AddedByLogikCheck,omitempty"`
	ContainedEquipmentPositions struct {
		EquipmentPosition []*EquipmentPosition `xml:"EquipmentPosition,omitempty" json:"EquipmentPosition,omitempty"`
	} `xml:"ContainedEquipmentPositions,omitempty" json:"ContainedEquipmentPositions,omitempty"`
	DatEquipmentIdReason    int64  `xml:"DatEquipmentIdReason,omitempty" json:"DatEquipmentIdReason,omitempty"`
	DatEquipmentIdReason2   int64  `xml:"DatEquipmentIdReason2,omitempty" json:"DatEquipmentIdReason2,omitempty"`
	EquipmentClassification int64  `xml:"EquipmentClassification,omitempty" json:"EquipmentClassification,omitempty"`
	ManualDecreaseType      string `xml:"ManualDecreaseType,omitempty" json:"ManualDecreaseType,omitempty"`
	VersionAccording1       int64  `xml:"VersionAccording1,omitempty" json:"VersionAccording1,omitempty"`
	VersionAccording2       int64  `xml:"VersionAccording2,omitempty" json:"VersionAccording2,omitempty"`
	VersionAccording3       int64  `xml:"VersionAccording3,omitempty" json:"VersionAccording3,omitempty"`
	VersionAccording4       int64  `xml:"VersionAccording4,omitempty" json:"VersionAccording4,omitempty"`
	VersionAccording5       int64  `xml:"VersionAccording5,omitempty" json:"VersionAccording5,omitempty"`
}

type Tires struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Tires"`

	TireRepairType string `xml:"TireRepairType,omitempty" json:"TireRepairType,omitempty"`

	TireValuationType string `xml:"TireValuationType,omitempty" json:"TireValuationType,omitempty"`

	Axles struct {
		Axle []*Axle `xml:"Axle,omitempty" json:"Axle,omitempty"`
	} `xml:"Axles,omitempty" json:"Axles,omitempty"`
}

type Axle struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs Axle"`

	AxleNo                      int64    `xml:"AxleNo,omitempty" json:"AxleNo,omitempty"`
	TireId                      int64    `xml:"TireId,omitempty" json:"TireId,omitempty"` //nolint
	TireState                   string   `xml:"TireState,omitempty" json:"TireState,omitempty"`
	NrOfTires                   int64    `xml:"NrOfTires,omitempty" json:"NrOfTires,omitempty"`
	TireType                    string   `xml:"TireType,omitempty" json:"TireType,omitempty"`
	TireTypeTextId              string   `xml:"TireTypeTextId,omitempty" json:"TireTypeTextId,omitempty"` //nolint
	TireOriginalPrice           *float32 `xml:"TireOriginalPrice,omitempty" json:"TireOriginalPrice,omitempty"`
	TireSpeedIndex              string   `xml:"TireSpeedIndex,omitempty" json:"TireSpeedIndex,omitempty"`
	TireSize                    string   `xml:"TireSize,omitempty" json:"TireSize,omitempty"`
	TireSafetySystem            string   `xml:"TireSafetySystem,omitempty" json:"TireSafetySystem,omitempty"`
	TireManufacturer            int64    `xml:"TireManufacturer,omitempty" json:"TireManufacturer,omitempty"`
	TireManufacturerName        string   `xml:"TireManufacturerName,omitempty" json:"TireManufacturerName,omitempty"`
	PrintLoadCapacityIdx        *bool    `xml:"PrintLoadCapacityIdx,omitempty" json:"PrintLoadCapacityIdx,omitempty"` //nolint
	TireOriginalTreadDepth      int64    `xml:"TireOriginalTreadDepth,omitempty" json:"TireOriginalTreadDepth,omitempty"`
	TireOriginalTreadDepthUser  int64    `xml:"TireOriginalTreadDepthUser,omitempty" json:"TireOriginalTreadDepthUser,omitempty"`
	TireOriginalTreadDepthN     *float32 `xml:"TireOriginalTreadDepthN,omitempty" json:"TireOriginalTreadDepthN,omitempty"`
	TireOriginalTreadDepthNUser *float32 `xml:"TireOriginalTreadDepthNUser,omitempty" json:"TireOriginalTreadDepthNUser,omitempty"`
	TireLoadCapacityIndex       int64    `xml:"TireLoadCapacityIndex,omitempty" json:"TireLoadCapacityIndex,omitempty"`
	TireLoadCapacityIndex2      int64    `xml:"TireLoadCapacityIndex2,omitempty" json:"TireLoadCapacityIndex2,omitempty"`
	TreadDepthLeftOuterPerc     *float32 `xml:"TreadDepthLeftOuterPerc,omitempty" json:"TreadDepthLeftOuterPerc,omitempty"`
	TreadDepthLeftInnerPerc     *float32 `xml:"TreadDepthLeftInnerPerc,omitempty" json:"TreadDepthLeftInnerPerc,omitempty"`
	TreadDepthRightInnerPerc    *float32 `xml:"TreadDepthRightInnerPerc,omitempty" json:"TreadDepthRightInnerPerc,omitempty"`
	TreadDepthRightOuterPerc    *float32 `xml:"TreadDepthRightOuterPerc,omitempty" json:"TreadDepthRightOuterPerc,omitempty"`
	TreadDepthLeftOuterMm       *float32 `xml:"TreadDepthLeftOuterMm,omitempty" json:"TreadDepthLeftOuterMm,omitempty"`
	TreadDepthLeftInnerMm       *float32 `xml:"TreadDepthLeftInnerMm,omitempty" json:"TreadDepthLeftInnerMm,omitempty"`
	TreadDepthRightInnerMm      *float32 `xml:"TreadDepthRightInnerMm,omitempty" json:"TreadDepthRightInnerMm,omitempty"`
	TreadDepthRightOuterMm      *float32 `xml:"TreadDepthRightOuterMm,omitempty" json:"TreadDepthRightOuterMm,omitempty"`
	ManualEntry                 *bool    `xml:"ManualEntry,omitempty" json:"ManualEntry,omitempty"`
	RetreadedLeftOuter          *bool    `xml:"RetreadedLeftOuter,omitempty" json:"RetreadedLeftOuter,omitempty"`
	RetreadedLeftInner          *bool    `xml:"RetreadedLeftInner,omitempty" json:"RetreadedLeftInner,omitempty"`
	RetreadedRightInner         *bool    `xml:"RetreadedRightInner,omitempty" json:"RetreadedRightInner,omitempty"`
	RetreadedRightOuter         *bool    `xml:"RetreadedRightOuter,omitempty" json:"RetreadedRightOuter,omitempty"`
	TireAveragePriceUser        *float32 `xml:"TireAveragePriceUser,omitempty" json:"TireAveragePriceUser,omitempty"`
	TireBrandPrice              *float32 `xml:"TireBrandPrice,omitempty" json:"TireBrandPrice,omitempty"`
	TireBrandPriceUser          *float32 `xml:"TireBrandPriceUser,omitempty" json:"TireBrandPriceUser,omitempty"`
	TireManufacturerId          int64    `xml:"TireManufacturerId,omitempty" json:"TireManufacturerId,omitempty"`
	TireManufacturerTextId      int64    `xml:"TireManufacturerTextId,omitempty" json:"TireManufacturerTextId,omitempty"`
	TireBrandId                 int64    `xml:"TireBrandId,omitempty" json:"TireBrandId,omitempty"` //nolint
	TireBrandName               string   `xml:"TireBrandName,omitempty" json:"TireBrandName,omitempty"`
	TireBrandTextId             int64    `xml:"TireBrandTextId,omitempty" json:"TireBrandTextId,omitempty"` //nolint
	TireBrandEanCode            string   `xml:"TireBrandEanCode,omitempty" json:"TireBrandEanCode,omitempty"`
	ProductCodeNumber           int64    `xml:"ProductCodeNumber,omitempty" json:"ProductCodeNumber,omitempty"`
}

type VINVehicle struct {
	XMLName                      xml.Name  `xml:"http://www.dat.de/vxs VINVehicle"`
	VINumber                     *VINumber `xml:"VINumber,omitempty" json:"VINumber,omitempty"`
	ManufacturerCarCode          string    `xml:"ManufacturerCarCode,omitempty" json:"ManufacturerCarCode,omitempty"`
	ManufacturerEngineCode       string    `xml:"ManufacturerEngineCode,omitempty" json:"ManufacturerEngineCode,omitempty"`
	ManufacturerTransmissionCode string    `xml:"ManufacturerTransmissionCode,omitempty" json:"ManufacturerTransmissionCode,omitempty"`
}

type VINResult struct {
	XMLName             xml.Name       `xml:"http://www.dat.de/vxs VINResult"`
	VinInterfaceVersion string         `xml:"VinInterfaceVersion,omitempty" json:"VinInterfaceVersion,omitempty"`
	VinDatProcedure     *bool          `xml:"VinDatProcedure,omitempty" json:"VinDatProcedure,omitempty"`
	CrossBorder         *bool          `xml:"CrossBorder,omitempty" json:"CrossBorder,omitempty"`
	VINECodes           *VINECodes     `xml:"VINECodes,omitempty" json:"VINECodes,omitempty"`
	VINEquipments       *VINEquipments `xml:"VINEquipments,omitempty" json:"VINEquipments,omitempty"`
	VINColors           *VINColors     `xml:"VINColors,omitempty" json:"VINColors,omitempty"`
	VINVehicle          *VINVehicle    `xml:"VINVehicle,omitempty" json:"VINVehicle,omitempty"`
}

type VINColors struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VINColors"`

	VINColor []*VINColor `xml:"VINColor,omitempty" json:"VINColor,omitempty"`
}

type VINColor struct {
	XMLName       xml.Name `xml:"http://www.dat.de/vxs VINColor"`
	ColorID       string   `xml:"ColorID,omitempty" json:"ColorID,omitempty"`
	Code          string   `xml:"Code,omitempty" json:"Code,omitempty"`
	Description   string   `xml:"Description,omitempty" json:"Description,omitempty"`
	StandardColor string   `xml:"StandardColor,omitempty" json:"StandardColor,omitempty"`
	PaintType     string   `xml:"PaintType,omitempty" json:"PaintType,omitempty"`
	CountCoat     string   `xml:"CountCoat,omitempty" json:"CountCoat,omitempty"`
}

type VINEquipments struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VINEquipments"`

	VINEquipment []*VINEquipment `xml:"VINEquipment,omitempty" json:"VINEquipment,omitempty"`
}

type VINEquipment struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VINEquipment"`

	AvNumberDat      int64  `xml:"AvNumberDat,omitempty" json:"AvNumberDat,omitempty"`
	ManufacturerCode string `xml:"ManufacturerCode,omitempty" json:"ManufacturerCode,omitempty"`
	ShortName        string `xml:"ShortName,omitempty" json:"ShortName,omitempty"`
}

type VINECode struct {
	XMLName                   xml.Name       `xml:"http://www.dat.de/vxs VINECode"`
	Sign                      int64          `xml:"Sign,omitempty" json:"Sign,omitempty"`
	Country                   string         `xml:"Country,omitempty" json:"Country,omitempty"`
	VehicleTypeKey            int64          `xml:"VehicleTypeKey,omitempty" json:"VehicleTypeKey,omitempty"`
	ManufacturerKey           int64          `xml:"ManufacturerKey,omitempty" json:"ManufacturerKey,omitempty"`
	VehicleMainTypeKey        int64          `xml:"VehicleMainTypeKey,omitempty" json:"VehicleMainTypeKey,omitempty"`
	VehicleSubTypeKey         int64          `xml:"VehicleSubTypeKey,omitempty" json:"VehicleSubTypeKey,omitempty"`
	VehicleSubTypeVariantKey  int64          `xml:"VehicleSubTypeVariantKey,omitempty" json:"VehicleSubTypeVariantKey,omitempty"`
	ConstructionTimeMin       int64          `xml:"ConstructionTimeMin,omitempty" json:"ConstructionTimeMin,omitempty"`
	ConstructionTime          int64          `xml:"ConstructionTime,omitempty" json:"ConstructionTime,omitempty"`
	ConstructionTimeEdge      int64          `xml:"ConstructionTimeEdge,omitempty" json:"ConstructionTimeEdge,omitempty"`
	ConstructionTimeProd      int64          `xml:"ConstructionTimeProd,omitempty" json:"ConstructionTimeProd,omitempty"`
	ConstructionTimePriceList int64          `xml:"ConstructionTimePriceList,omitempty" json:"ConstructionTimePriceList,omitempty"`
	VINContainers             *VINContainers `xml:"VINContainers,omitempty" json:"VINContainers,omitempty"`
}

type VINContainers struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VINContainers"`

	VINContainer []*VINContainer `xml:"VINContainer,omitempty" json:"VINContainer,omitempty"`
}

type VINContainer struct {
	XMLName                 xml.Name `xml:"http://www.dat.de/vxs VINContainer"`
	Container               string   `xml:"Container,omitempty" json:"Container,omitempty"`
	VehicleTypeKey          int64    `xml:"VehicleTypeKey,omitempty" json:"VehicleTypeKey,omitempty"`
	ManufacturerKey         int64    `xml:"ManufacturerKey,omitempty" json:"ManufacturerKey,omitempty"`
	VehicleMainTypeKey      int64    `xml:"VehicleMainTypeKey,omitempty" json:"VehicleMainTypeKey,omitempty"`
	VehicleSubTypeKey       int64    `xml:"VehicleSubTypeKey,omitempty" json:"VehicleSubTypeKey,omitempty"`
	VehicleConstructionTime int64    `xml:"VehicleConstructionTime,omitempty" json:"VehicleConstructionTime,omitempty"`
}

type VINECodes struct {
	XMLName xml.Name `xml:"http://www.dat.de/vxs VINECodes"`

	VINECode []*VINECode `xml:"VINECode,omitempty" json:"VINECode,omitempty"`
}

type VINumber struct {
	XMLName      xml.Name `xml:"http://www.dat.de/vxs VINumber"`
	VinCode      string   `xml:"VinCode,omitempty" json:"VinCode,omitempty"`
	OrderCode    string   `xml:"OrderCode,omitempty" json:"OrderCode,omitempty"`
	Manufacturer string   `xml:"Manufacturer,omitempty" json:"Manufacturer,omitempty"`
}
