package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	stringutils "github.com/DIMO-Network/shared/pkg/strings"

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
}

// UnmarshalJSON Custom unmarshaller for CompatibilitySheetRow struct because model can sometimes be interpreted as a number
func (v *CompatibilitySheetRow) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with Model as interface{} to handle both types
	type Alias CompatibilitySheetRow
	temp := &struct {
		Model interface{} `json:"model"`
		*Alias
	}{
		Alias: (*Alias)(v),
	}

	// Unmarshal into the temporary struct
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Handle the model field depending on its type
	switch model := temp.Model.(type) {
	case string:
		v.Model = model
	case float64:
		v.Model = strconv.Itoa(int(model)) // Convert number to string
	default:
		v.Model = "" // Or handle any unexpected type here
	}

	return nil
}

func getCompatibilityR1SheetData(ctx context.Context, settings *config.Settings) ([]CompatibilitySheetRow, error) {
	srv, err := sheets.NewService(ctx,
		option.WithCredentialsJSON([]byte(settings.GoogleSheetsCredentials)),
		option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	spreadsheetID := "1immL2UJb27I2WLBJwQs29HF68sBbLYSLUB9EpnwovTQ" // Replace with your actual spreadsheet ID
	rangeData := "R1 Compatibility Checker!A1:D"                    // Replace with the relevant range
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
		if len(row) >= 4 {
			mk := fmt.Sprintf("%v", row[0])
			model := fmt.Sprintf("%v", row[1])
			yr, _ := strconv.Atoi(fmt.Sprintf("%v", row[2]))
			compat := fmt.Sprintf("%v", row[3])

			rows = append(rows, CompatibilitySheetRow{
				DefinitionID: common.DeviceDefinitionSlug(stringutils.SlugString(mk), stringutils.SlugString(model), int16(yr)),
				Make:         mk,
				Model:        model,
				Year:         yr,
				Compatible:   compat,
			})
		}
	}

	return rows, nil
}
