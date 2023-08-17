package main

import (
	"context"
	"flag"
	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/commands"
	"github.com/DIMO-Network/device-definitions-api/internal/core/services"
	"github.com/DIMO-Network/shared/db"
	"github.com/TheFellow/go-mediator/mediator"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type powerTrainTypeCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	force bool
}

func (*powerTrainTypeCmd) Name() string     { return "syncpowertraintype" }
func (*powerTrainTypeCmd) Synopsis() string { return "sync powertraintype" }
func (*powerTrainTypeCmd) Usage() string {
	return `syncpowertraintype [-force]`
}

func (p *powerTrainTypeCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.force, "force", false, "Enable Force Powertraintype")
}

func (p *powerTrainTypeCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	//db
	pdb := db.NewDbConnectionFromSettings(ctx, &p.settings.DB, true)
	pdb.WaitForDB(p.logger)

	powerTrainTypeService := services.NewPowerTrainTypeService(pdb.DBS, &p.logger)

	//commands
	m, _ := mediator.New(
		mediator.WithBehaviour(common.NewLoggingBehavior(&p.logger, &p.settings)),
		mediator.WithBehaviour(common.NewValidationBehavior(&p.logger, &p.settings)),
		mediator.WithHandler(&commands.SyncPowerTrainTypeCommand{},
			commands.NewSyncPowerTrainTypeCommandHandler(pdb.DBS, p.logger, powerTrainTypeService)),
	)

	_, _ = m.Send(ctx, &commands.SyncPowerTrainTypeCommand{ForceUpdate: p.force, DeviceTypeID: "vehicle"})

	return subcommands.ExitSuccess
}
