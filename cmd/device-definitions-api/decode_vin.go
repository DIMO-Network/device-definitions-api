package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-json"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type decodeVINCmd struct {
	logger   *zerolog.Logger
	settings *config.Settings

	datGroup   bool
	drivly     bool
	vincario   bool
	japan17vin bool
	fromFile   bool
}

func (*decodeVINCmd) Name() string { return "decodevin" }
func (*decodeVINCmd) Synopsis() string {
	return "tries decoding a vin with chosen provider - does not insert in our db"
}
func (*decodeVINCmd) Usage() string {
	return `decodevin [-dat|-drivly|-vincario|-japan17vin|-from-file] <vin 17 chars OR filaname in /tmp> <country two letter iso>`
}

func (p *decodeVINCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.datGroup, "dat", false, "use dat group vin decoder")
	f.BoolVar(&p.drivly, "drivly", false, "use drivly vin decoder")
	f.BoolVar(&p.vincario, "vincario", false, "use vincario vin decoder")
	f.BoolVar(&p.japan17vin, "japan17vin", false, "use japan17vin vin decoder")
	f.BoolVar(&p.fromFile, "from-file", false, "read vin from file in /tmp directory")
}

func (p *decodeVINCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) == 0 {
		if p.fromFile {
			fmt.Println("missing filename parameter")
		} else {
			fmt.Println("missing vin parameter")
		}
		return subcommands.ExitUsageError
	}
	vinOrFile := f.Args()[0]

	country := "USA"
	if len(f.Args()) == 2 {
		country = f.Args()[1]
	}
	vins := []string{}
	if p.fromFile {
		fmt.Printf("Filename: %s\n", vinOrFile)
		vins = loadVINsFromFile(vinOrFile)
	} else {
		fmt.Printf("VIN: %s\n", vinOrFile)
		vins = append(vins, vinOrFile)
	}
	fmt.Printf("Country: %s\n", country)

	fmt.Printf("total VINs found: %d\n", len(vins))
	if len(vins) == 0 {
		fmt.Println("no vins found")
		return subcommands.ExitFailure
	}

	for _, vin := range vins {
		if p.datGroup {
			// use the dat group service to decode
			datAPI := gateways.NewDATGroupAPIService(p.settings, p.logger)
			vinInfo, err := datAPI.GetVINv2(vin, country)
			if err != nil {
				fmt.Println(err.Error())
				return subcommands.ExitFailure
			}

			fmt.Printf("\n\nVIN Response: %+v\n", *vinInfo)
		}
		if p.drivly {
			drivlyAPI := gateways.NewDrivlyAPIService(p.settings)
			vinInfo, err := drivlyAPI.GetVINInfo(vin)
			if err != nil {
				fmt.Println(err.Error())
				return subcommands.ExitFailure
			}

			fmt.Printf("VIN Response: %+v\n", vinInfo)
		}
		if p.vincario {
			vincarioAPI := gateways.NewVincarioAPIService(p.settings, p.logger)
			vinInfo, err := vincarioAPI.DecodeVIN(vin)
			if err != nil {
				fmt.Println(err.Error())
				return subcommands.ExitFailure
			}

			fmt.Printf("VIN Response: %+v\n", vinInfo)
		}
		if p.japan17vin {
			jp17vinAPI := gateways.NewJapan17VINAPI(p.logger, p.settings)
			vinInfo, payload, err := jp17vinAPI.GetVINInfo(vin)
			if err != nil {
				fmt.Println(err.Error())
				return subcommands.ExitFailure
			}
			jsonBytes, _ := json.MarshalIndent(vinInfo, "", " ")
			fmt.Println("VIN Info:")
			fmt.Println(string(jsonBytes))
			fmt.Println("Raw JSON Payload:")
			fmt.Println(string(payload))
		}
		fmt.Println()
	}
	return subcommands.ExitSuccess
}

func loadVINsFromFile(file string) []string {
	// files are assumed to be in the tmp directory
	// pull out vins from csv file from column "vin"
	vinFile := "/tmp/" + file
	vinFileContents, err := readVINFile(vinFile)
	if err != nil {
		fmt.Println(err.Error())
		return []string{}
	}
	return vinFileContents
}

func readVINFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header row to find the index of "csv" column
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	csvColumnIndex := -1
	for i, columnName := range header {
		if columnName == "csv" {
			csvColumnIndex = i
			break
		}
	}

	if csvColumnIndex == -1 {
		csvColumnIndex = 0 // default to first column if "csv" column not found
		fmt.Println("defaulting to first column as 'csv' column not found, please ensure your CSV file has a column named 'csv' with VINs in it.")
	}

	// Read all rows and extract values from the "csv" column
	var values []string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if csvColumnIndex < len(row) {
			values = append(values, row[csvColumnIndex])
		}
	}

	return values, nil
}
