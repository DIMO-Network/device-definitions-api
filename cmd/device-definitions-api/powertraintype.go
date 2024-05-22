package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type powerTrainTypeCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	force bool
}

func (*powerTrainTypeCmd) Name() string { return "syncpowertraintype" }
func (*powerTrainTypeCmd) Synopsis() string {
	return "figures out the right powertrain for a device definition based on rules and logic."
}
func (*powerTrainTypeCmd) Usage() string {
	return `syncpowertraintype [-force]
			force flag will overwrite any powertrain settings already set. can be useful when update rules.
			This script does not set anything on device_styles metadata powertrain.`
}

func (p *powerTrainTypeCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.force, "force", false, "set powertrain even if already set - default is to not modify.")
}

func (p *powerTrainTypeCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	powerTrainTypeService, err := services.NewPowerTrainTypeService(pdb.DBS, "powertrain_type_rule.yaml", &p.logger)
	if err != nil {
		p.logger.Err(err).Stack().Send()
	}

	//commands
	m, _ := mediator.New(
		//mediator.WithBehaviour(common.NewLoggingBehavior(&p.logger, &p.settings)),
		//mediator.WithBehaviour(common.NewValidationBehavior(&p.logger, &p.settings)),
		mediator.WithHandler(&commands.SyncPowerTrainTypeCommand{},
			commands.NewSyncPowerTrainTypeCommandHandler(pdb.DBS, p.logger, powerTrainTypeService)),
	)

	_, _ = m.Send(ctx, &commands.SyncPowerTrainTypeCommand{ForceUpdate: p.force, DeviceTypeID: "vehicle"})

	return subcommands.ExitSuccess
}
