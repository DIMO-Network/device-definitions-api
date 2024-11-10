package queries

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"strconv"
)

type CompatibilitySheetRow struct {
	DefinitionId string `json:"definitionId"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Compatible   string `json:"compatible"`
	Vin          string `json:"vin"`
	Odometer     string `json:"odometer"`
	Id           int    `json:"id"`
}

func getClient(config *oauth2.Config) *http.Client {
	token := &oauth2.Token{}
	client := config.Client(context.Background(), token)
	return client
}

func GetCompatibilityR1SheetData() ([]CompatibilitySheetRow, error) {
	ctx := context.Background()
	config, err := google.JWTConfigFromJSON([]byte(os.Getenv("GOOGLE_SHEETS_CREDENTIALS")), sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	srv, err := sheets.New(config.Client(ctx))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := "1PjUQm84M5xcEGDpykKyjljhslmzGpqaAbRHZQi8q1f0" // Replace with your actual spreadsheet ID
	rangeData := "R1API!A1:E"                                       // Replace with the relevant range
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, rangeData).Do()
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
				DefinitionId: fmt.Sprintf("%v", row[0]),
				Make:         fmt.Sprintf("%v", row[1]),
				Model:        fmt.Sprintf("%v", row[2]),
				Year:         yr,
				Compatible:   fmt.Sprintf("%v", row[4]),
			})
		}
	}

	return rows, nil
}
