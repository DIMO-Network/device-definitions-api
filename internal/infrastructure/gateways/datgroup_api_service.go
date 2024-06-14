package gateways

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"

	"github.com/rs/zerolog"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
)

//go:generate mockgen -source datgroup_api_service.go -destination mocks/datgroup_api_service_mock.go -package mocks
type DATGroupAPIService interface {
	GetVINv2(vin string, country string) (*DATGroupInfoResponse, error)
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

func (ai *datGroupAPIService) GetVINv2(vin, userCountryISO2 string) (*DATGroupInfoResponse, error) {
	if userCountryISO2 == "" || len(userCountryISO2) != 2 {
		userCountryISO2 = "US"
	}
	customerLogin := ai.Settings.DatGroupCustomerLogin
	customerNumber := ai.Settings.DatGroupCustomerNumber
	customerSignature := ai.Settings.DatGroupCustomerSignature
	interfacePartnerSignature := ai.Settings.DatGroupInterfacePartnerSignature

	soapReq := `
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:veh="http://sphinx.dat.de/services/VehicleIdentificationService">
<soapenv:Header>
<customerLogin>%s</customerLogin>
<customerNumber>%s</customerNumber>
<interfacePartnerNumber>%s</interfacePartnerNumber>
<customerSignature>%s</customerSignature>
<interfacePartnerSignature>%s</interfacePartnerSignature>
</soapenv:Header>
<soapenv:Body>
<veh:getVehicleIdentificationByVin>
<!-- Optional: -->
<request>
<locale country="%s" datCountryIndicator="TR" language="EN"/>
<!-- Optional: -->
<!-- Zero or more repetitions: -->
<coverage>ALL</coverage>
<restriction>ALL</restriction>
<vin>%s</vin>
</request>
<templateId>157011</templateId>
</veh:getVehicleIdentificationByVin>
</soapenv:Body>
</soapenv:Envelope>
`
	soapReqWParams := fmt.Sprintf(soapReq, customerLogin, customerNumber, customerNumber, customerSignature, interfacePartnerSignature,
		userCountryISO2, vin)

	ai.log.Debug().Msg(soapReqWParams)

	timeout := 30 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", ai.Settings.DatGroupURL, bytes.NewBuffer([]byte(soapReqWParams)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Accept", "text/xml, multipart/related")
	req.Header.Set("SOAPAction", "getVehicleIdentificationByVin")
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	response, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	ai.log.Debug().Msg(string(bodyBytes))

	if response.StatusCode != http.StatusOK {
		ai.log.Error().Str("vin", vin).Msgf("error response status code: %d. request: %s",
			response.StatusCode, soapReqWParams)
		return nil, fmt.Errorf("error response status code: %d", response.StatusCode)
	}

	infoResponse, err := parseXML(ai.log, string(bodyBytes), vin)
	if err != nil {
		return nil, err
	}

	return infoResponse, err
}

func parseXML(logger *zerolog.Logger, datgroupRespXML, vin string) (*DATGroupInfoResponse, error) {
	doc, err := xmlquery.Parse(strings.NewReader(datgroupRespXML))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse response XML")
	}
	vehicle := xmlquery.FindOne(doc, "//ns1:Vehicle")
	if vehicle == nil {
		return nil, errors.New("failed to find vehicle xml node")
	}

	response := &DATGroupInfoResponse{
		VIN:               vin,
		DatECode:          getXMLValue(vehicle, "//ns1:DatECode"),
		SalesDescription:  getXMLValue(vehicle, "//ns1:SalesDescription"),
		VehicleTypeName:   getXMLValue(vehicle, "//ns1:VehicleTypeName"),
		ManufacturerName:  getXMLValue(vehicle, "//ns1:ManufacturerName"),
		BaseModelName:     getXMLValue(vehicle, "//ns1:BaseModelName"),
		SubModelName:      getXMLValue(vehicle, "//ns1:SubModelName"),
		MainTypeGroupName: getXMLValue(vehicle, "//ns1:MainTypeGroupName"),
	}
	response.VinAccuracy, _ = strconv.Atoi(getXMLValue(vehicle, "//ns1:MainTypeGroupName"))
	yearLow, yearHigh, err := extractYears(getXMLValue(vehicle, "//ns1:ContainerName"))
	if err != nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to extract year low/year for datgroup vin decode")
	}

	if yearLow > 2000 {
		response.YearLow = yearLow
	}

	if yearHigh > 2000 {
		response.YearHigh = yearHigh
	}

	yr := shared.VIN(response.VIN).Year()
	if yr >= response.YearLow && yr <= response.YearHigh {
		response.Year = yr
	} else {
		response.Year = response.YearHigh
	}

	// series equipment
	seriesEquipment := xmlquery.FindOne(vehicle, "//ns1:SeriesEquipment")

	if seriesEquipment == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find series equipment node")
		return response, nil
	}

	seNodes := xmlquery.Find(seriesEquipment, "//ns1:EquipmentPosition")

	if seNodes == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find series equipment nodes")
		return response, nil
	}

	for _, seNode := range seNodes {
		equipment := DATGroupEquipment{
			DatEquipmentId:          getXMLValue(seNode, "//ns1:DatEquipmentId"),
			ManufacturerEquipmentId: getXMLValue(seNode, "//ns1:ManufacturerEquipmentId"),
			ManufacturerDescription: getXMLValue(seNode, "//ns1:ManufacturerDescription"),
			Description:             getXMLValue(seNode, "//ns1:Description"),
		}
		response.SeriesEquipment = append(response.SeriesEquipment, equipment)
	}

	// special equipment
	specialEquipment := xmlquery.FindOne(vehicle, "//ns1:SpecialEquipment")

	if specialEquipment == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find special equipment node")
		return response, nil
	}

	spNodes := xmlquery.Find(specialEquipment, "//ns1:EquipmentPosition")

	if spNodes == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find special equipment nodes")
		return response, nil
	}

	for _, seNode := range spNodes {
		equipment := DATGroupEquipment{
			DatEquipmentId:          getXMLValue(seNode, "//ns1:DatEquipmentId"),
			ManufacturerEquipmentId: getXMLValue(seNode, "//ns1:ManufacturerEquipmentId"),
			ManufacturerDescription: getXMLValue(seNode, "//ns1:ManufacturerDescription"),
			Description:             getXMLValue(seNode, "//ns1:Description"),
		}
		response.SpecialEquipment = append(response.SpecialEquipment, equipment)
	}

	// DATECode Equipment
	datECodeEquipment := xmlquery.FindOne(vehicle, "//ns1:DATECodeEquipment")

	if datECodeEquipment == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find datECode equipment node")
		return response, nil
	}

	decNodes := xmlquery.Find(datECodeEquipment, "//ns1:EquipmentPosition")

	if decNodes == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find datECode equipment nodes")
		return response, nil
	}

	for _, seNode := range decNodes {
		equipment := DATGroupEquipment{
			DatEquipmentId: getXMLValue(seNode, "//ns1:DatEquipmentId"),
			Description:    getXMLValue(seNode, "//ns1:Description"),
		}
		response.DATECodeEquipment = append(response.DATECodeEquipment, equipment)
	}

	// VIN Equipment
	vinEquipment := xmlquery.FindOne(vehicle, "//ns1:VINEquipments")

	if vinEquipment == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find vin equipment node")
		return response, nil
	}

	vinNodes := xmlquery.Find(vinEquipment, "//ns1:VINEquipment")

	if vinNodes == nil {
		logger.Err(err).Str("vin", vin).Msgf("failed to find vin equipment nodes")
		return response, nil
	}

	for _, seNode := range vinNodes {
		equipment := DATGroupEquipment{
			ManufacturerEquipmentId: getXMLValue(seNode, "//ns1:ManufacturerCode"),
			ManufacturerDescription: getXMLValue(seNode, "//ns1:ShortName"),
		}
		response.VINEquipment = append(response.VINEquipment, equipment)
	}

	return response, nil
}

func getXMLValue(doc *xmlquery.Node, field string) string {
	value := xmlquery.FindOne(doc, field)
	if value != nil && value.InnerText() != "" {
		return value.InnerText()
	}
	return ""
}

func extractYears(s string) (int, int, error) {
	splitString := strings.Split(s, "-")
	if len(splitString) < 3 {
		return 0, 0, fmt.Errorf("input format is incorrect")
	}

	startYearString := strings.TrimSpace(splitString[1])
	endYearString := strings.Split(splitString[2], " ")[1]

	sub := startYearString[len(startYearString)-4:] // need to take the last 4 chars
	startYear, err1 := strconv.Atoi(sub)
	endYear, err2 := strconv.Atoi(endYearString)

	if err1 != nil || err2 != nil {
		return startYear, endYear, fmt.Errorf("unable to convert year string to int")
	}

	return startYear, endYear, nil
}

type DATGroupInfoResponse struct {
	VIN              string `json:"vin"`
	DatECode         string `json:"datecode"`
	SalesDescription string `json:"salesDescription"`
	VehicleTypeName  string `json:"vehicleTypeName"`
	// make
	ManufacturerName string `json:"manufacturerName"`
	BaseModelName    string `json:"baseModelName"`
	SubModelName     string `json:"subModelName"`
	// this is the model name we want to use
	MainTypeGroupName string `json:"mainTypeGroupName"`
	VinAccuracy       int    `json:"vinAccuracy"`

	// when we're unable to get exact year
	YearLow  int `json:"yearLow"`
	YearHigh int `json:"yearHigh"`
	// we don't always get the exact year
	Year int `json:"year"`

	SeriesEquipment   []DATGroupEquipment `json:"seriesEquipment"`
	SpecialEquipment  []DATGroupEquipment `json:"specialEquipment"`
	DATECodeEquipment []DATGroupEquipment `json:"datECodeEquipment"`
	VINEquipment      []DATGroupEquipment `json:"vinEquipments"`
}

type DATGroupEquipment struct {
	DatEquipmentId          string `json:"datEquipmentId"`
	ManufacturerEquipmentId string `json:"manufacturerEquipmentId"`
	// if Vin Equipment, this comes from ShortName
	ManufacturerDescription string `json:"manufacturerDescription"`
	Description             string `json:"description"`
}
