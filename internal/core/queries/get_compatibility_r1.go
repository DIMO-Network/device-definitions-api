package queries

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/api/option"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"google.golang.org/api/sheets/v4"
)

type GetCompatibilityR1SheetQueryHandler struct {
	settings *config.Settings
}

type GetCompatibilityR1SheetQuery struct {
}

func (*GetCompatibilityR1SheetQuery) Key() string { return "GetCompatibilityR1SheetQuery" }

func NewCompatibilityR1SheetQueryHandler(settings *config.Settings) GetCompatibilityR1SheetQueryHandler {
	return GetCompatibilityR1SheetQueryHandler{
		settings: settings,
	}
}

func (crs GetCompatibilityR1SheetQueryHandler) Handle(ctx context.Context, _ mediator.Message) (interface{}, error) {
	return getCompatibilityR1SheetData(ctx, crs.settings)
}

type CompatibilitySheetRow struct {
	DefinitionID string `json:"definitionId"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Compatible   string `json:"compatible"`
	Vin          string `json:"vin"`
	Odometer     string `json:"odometer"`
	ID           int    `json:"id"`
}

//func getClient(config *oauth2.Config) *http.Client {
//	token := &oauth2.Token{}
//	client := config.Client(context.Background(), token)
//	return client
//}

func getCompatibilityR1SheetData(ctx context.Context, settings *config.Settings) ([]CompatibilitySheetRow, error) {
	srv, err := sheets.NewService(ctx,
		option.WithCredentialsJSON([]byte(settings.GoogleSheetsCredentials)),
		option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	spreadsheetID := "1PjUQm84M5xcEGDpykKyjljhslmzGpqaAbRHZQi8q1f0" // Replace with your actual spreadsheet ID
	rangeData := "R1API!A1:E"                                       // Replace with the relevant range
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeData).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	var rows []CompatibilitySheetRow
	for i, row := range resp.Values {
		// Skip the header row
		if i == 0 {
			continue
		}
		if len(row) >= 3 {
			yr, _ := strconv.Atoi(fmt.Sprintf("%v", row[3]))
			rows = append(rows, CompatibilitySheetRow{
				DefinitionID: fmt.Sprintf("%v", row[0]),
				Make:         fmt.Sprintf("%v", row[1]),
				Model:        fmt.Sprintf("%v", row[2]),
				Year:         yr,
				Compatible:   fmt.Sprintf("%v", row[4]),
			})
		}
	}

	return rows, nil
}
