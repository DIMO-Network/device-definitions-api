package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/pkg/db"
	vinutil "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/goccy/go-json"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type decodeVINCmd struct {
	logger   *zerolog.Logger
	settings *config.Settings

	datGroup    bool
	drivly      bool
	vincario    bool
	japan17vin  bool
	fromFile    bool
	persistToDB bool
	carvx       bool
}

func (*decodeVINCmd) Name() string { return "decodevin" }
func (*decodeVINCmd) Synopsis() string {
	return "tries decoding a vin with chosen provider - does not insert in our db"
}
func (*decodeVINCmd) Usage() string {
	return `decodevin [-dat|-drivly|-vincario|-japan17vin|carvx|-from-file] <vin 17 chars OR filaname in /tmp> <country two letter iso>`
}

func (p *decodeVINCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.datGroup, "dat", false, "use dat group vin decoder")
	f.BoolVar(&p.drivly, "drivly", false, "use drivly vin decoder")
	f.BoolVar(&p.vincario, "vincario", false, "use vincario vin decoder")
	f.BoolVar(&p.japan17vin, "japan17vin", false, "use japan17vin vin decoder")
	f.BoolVar(&p.carvx, "carvx", false, "use carvx vin decoder")
	f.BoolVar(&p.fromFile, "from-file", false, "read vin from file in /tmp directory")
	f.BoolVar(&p.persistToDB, "persist-to-db", false, "persist successful vin decodings to db, table vin_numbers")
}

func (p *decodeVINCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) == 0 {
		if p.fromFile {
			fmt.Println("missing filename parameter")
		} else {
			fmt.Println("missing vin parameter")
		}
		return subcommands.ExitUsageError
	}
	vinOrFile := f.Args()[0]

	pdb := db.NewDbConnectionFromSettings(context.Background(), &p.settings.DB, true)
	pdb.WaitForDB(*p.logger)

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

	vinDecodingService := instantiateVINDecodingSvc(ctx, p.settings, p.logger, pdb)

	for _, vin := range vins {
		// in case want to insert
		vinObj := vinutil.VIN(vin)
		dbVin := &models.VinNumber{
			Vin:          vin,
			Wmi:          null.StringFrom(vinObj.Wmi()),
			VDS:          null.StringFrom(vinObj.VDS()),
			SerialNumber: vinObj.SerialNumber(),
			CheckDigit:   null.StringFrom(vinObj.CheckDigit()),
			Vis:          null.StringFrom(vinObj.VIS()),
		}
		wmi, _ := models.Wmis(models.WmiWhere.Wmi.EQ(vinObj.Wmi())).One(ctx, pdb.DBS().Reader)
		if wmi != nil {
			dbVin.ManufacturerName = wmi.ManufacturerName
		}
		_, err := models.DeviceTypes(models.DeviceTypeWhere.ID.EQ(common.DefaultDeviceType)).One(ctx, pdb.DBS().Reader)
		if err != nil {
			fmt.Println(err.Error())
			return subcommands.ExitFailure
		}
		vinInfo := &coremodels.VINDecodingInfoData{VIN: vin}

		if p.datGroup {
			vinInfo, _, err = vinDecodingService.GetVIN(ctx, vin, coremodels.DATGroupProvider, country)
			// use the dat group service to decode
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			fmt.Printf("\n\nVIN Response: %+v\n", vinInfo)
		}
		if p.drivly {
			vinInfo, _, err = vinDecodingService.GetVIN(ctx, vin, coremodels.DrivlyProvider, country)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			fmt.Printf("VIN Response: %+v\n", vinInfo)
		}
		if p.vincario {
			vinInfo, _, err = vinDecodingService.GetVIN(ctx, vin, coremodels.VincarioProvider, country)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fmt.Printf("VIN Response: %+v\n", vinInfo)
		}
		if p.carvx {
			vinInfo, _, err = vinDecodingService.GetVIN(ctx, vin, coremodels.CarVXVIN, country)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fmt.Printf("VIN Response: %+v\n", vinInfo)
		}
		if p.japan17vin {
			vinInfo, _, err = vinDecodingService.GetVIN(ctx, vin, coremodels.Japan17VIN, country)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			jsonBytes, _ := json.MarshalIndent(vinInfo, "", " ")
			fmt.Println("VIN Info:")
			fmt.Println(string(jsonBytes))
		}
		fmt.Println()
		if p.persistToDB {
			if vinInfo == nil || vinInfo.Model == "" {
				fmt.Println("no decoding info found, skipping: " + vin)
				continue
			}
			dbVin.Year = int(vinInfo.Year)
			if dbVin.ManufacturerName == "" {
				dbVin.ManufacturerName = vinInfo.Make
			}
			dbVin.DatgroupData = null.JSONFrom(vinInfo.Raw)
			dbVin.DefinitionID = common.DeviceDefinitionSlug(vinInfo.Make, vinInfo.Model, int16(vinInfo.Year))
			dbVin.DecodeProvider = null.StringFrom(string(vinInfo.Source))
			// todo future change to add field with StyleName

			err := dbVin.Insert(ctx, pdb.DBS().Writer, boil.Infer())
			if err != nil {
				fmt.Println(err.Error())
				return subcommands.ExitFailure
			}
		}
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
		if columnName == "vins" {
			csvColumnIndex = i
			break
		}
	}

	if csvColumnIndex == -1 {
		csvColumnIndex = 0 // default to first column if "csv" column not found
		fmt.Println("defaulting to first column as 'vins' column not found, please ensure your CSV file has a column named 'vins' with VINs in it.")
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

func instantiateVINDecodingSvc(ctx context.Context, settings *config.Settings, logger *zerolog.Logger, pdb db.Store) services.VINDecodingService {
	datAPI := gateways.NewDATGroupAPIService(settings, logger)
	drivlyAPI := gateways.NewDrivlyAPIService(settings)
	vincarioAPI := gateways.NewVincarioAPIService(settings, logger)
	jp17vinAPI := gateways.NewJapan17VINAPI(logger, settings)
	carvxAPI := gateways.NewCarVxVINAPI(logger, settings)
	elevaAPI := gateways.NewElevaAPI(settings)

	send, err := createSender(ctx, settings, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create sender.")
	}

	ethClient, err := ethclient.Dial(settings.EthereumRPCURL.String())
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create Ethereum client.")
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Couldn't retrieve chain id.")
	}
	deviceDefinitionOnChainService := gateways.NewDeviceDefinitionOnChainService(settings, logger, ethClient, chainID, send, pdb.DBS)

	vinDecodingService := services.NewVINDecodingService(drivlyAPI, vincarioAPI, nil, logger, deviceDefinitionOnChainService, datAPI, pdb.DBS, jp17vinAPI, carvxAPI, elevaAPI)

	return vinDecodingService
}
