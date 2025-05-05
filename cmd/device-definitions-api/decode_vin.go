package main

import (
	"context"
	"flag"
	"fmt"

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
}

func (*decodeVINCmd) Name() string { return "decodevin" }
func (*decodeVINCmd) Synopsis() string {
	return "tries decoding a vin with chosen provider - does not insert in our db"
}
func (*decodeVINCmd) Usage() string {
	return `decodevin [-dat|-drivly|-vincario|-japan17vin] <vin 17 chars> <country two letter iso>`
}

func (p *decodeVINCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.datGroup, "dat", false, "use dat group vin decoder")
	f.BoolVar(&p.drivly, "drivly", false, "use drivly vin decoder")
	f.BoolVar(&p.vincario, "vincario", false, "use vincario vin decoder")
	f.BoolVar(&p.japan17vin, "japan17vin", false, "use japan17vin vin decoder")
}

func (p *decodeVINCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) == 0 {
		fmt.Println("missing vin parameter")
		return subcommands.ExitUsageError
	}
	vin := f.Args()[0]

	country := "USA"
	if len(f.Args()) == 2 {
		country = f.Args()[1]
	}
	fmt.Printf("VIN: %s\n", vin)
	fmt.Printf("Country: %s\n", country)

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
	return subcommands.ExitSuccess
}
