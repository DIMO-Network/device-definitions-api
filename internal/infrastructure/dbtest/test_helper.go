package dbtest

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"os"
	"testing"

	"github.com/DIMO-Network/shared"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:embed device_type_vehicle_properties.json
var deviceTypeVehiclePropertyDataSample []byte

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, dbName string, t *testing.T, migrationsDirRelPath string) (db.Store, testcontainers.Container) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	settings := getTestDbSettings(dbName)
	pgPort := "5432/tcp"
	dbURL := func(_ string, port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DB.User, settings.DB.Password, port.Port(), settings.DB.Name)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:16.6-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DB.User, "POSTGRES_PASSWORD": settings.DB.Password, "POSTGRES_DB": settings.DB.Name},
		ExposedPorts: []string{pgPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForSQL(nat.Port(pgPort), "postgres", dbURL),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: cr,
		Started:          true,
	})
	if err != nil {
		return handleContainerStartErr(ctx, err, pgContainer, t)
	}
	mappedPort, err := pgContainer.MappedPort(ctx, nat.Port(pgPort))
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to get container external port"), pgContainer, t)
	}
	fmt.Printf("postgres container session %s ready and running at port: %s \n", pgContainer.SessionID(), mappedPort)
	//defer pgContainer.Terminate(ctx) // this should be done by the caller

	settings.DB.Port = mappedPort.Port()
	pdb := db.NewDbConnectionForTest(ctx, &settings.DB, false)
	pdb.WaitForDB(logger)

	_, err = pdb.DBS().Writer.Exec(fmt.Sprintf(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS %s;
		ALTER USER postgres SET search_path = %s, public;
		SET search_path = %s, public;
		`, dbName, dbName, dbName))
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrapf(err, "failed to apply schema. session: %s, port: %s",
			pgContainer.SessionID(), mappedPort.Port()), pgContainer, t)
	}
	// add truncate tables func
	_, err = pdb.DBS().Writer.Exec(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION truncate_tables() RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE schemaname = '%s' and tablename != 'migrations';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
`, dbName))
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to create truncate func"), pgContainer, t)
	}

	goose.SetTableName(dbName + ".migrations")
	if err := goose.RunContext(ctx, "up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to apply goose migrations for test"), pgContainer, t)
	}

	return pdb, pgContainer
}

// getTestDbSettings builds test db config.Settings object
func getTestDbSettings(dbName string) config.Settings {
	settings := config.Settings{
		LogLevel: "info",
		DB: db.Settings{
			Name:               dbName,
			Host:               "localhost",
			Port:               "6669",
			User:               "postgres",
			Password:           "postgres",
			MaxOpenConnections: 2,
			MaxIdleConnections: 2,
		},
		ServiceName: "device-definitions-api",
	}
	return settings
}

func handleContainerStartErr(ctx context.Context, err error, container testcontainers.Container, t *testing.T) (db.Store, testcontainers.Container) {
	if err != nil {
		fmt.Println("start container error: " + err.Error())
		if container != nil {
			container.Terminate(ctx) //nolint
		}
		t.Fatal(err)
	}
	return db.Store{}, container
}

// TruncateTables truncates tables for the test db, useful to run as teardown at end of each DB dependent test.
func TruncateTables(db *sql.DB, t *testing.T) {
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
		fmt.Println("truncating tables failed.")
		t.Fatal(err)
	}
}

func SetupCreateDeviceDefinition(t *testing.T, dm models.DeviceMake, model string, year int, pdb db.Store) *gateways.DeviceDefinitionTablelandModel {
	SetupCreateDeviceType(t, pdb)
	dd := &gateways.DeviceDefinitionTablelandModel{
		ID:         common.DeviceDefinitionSlug(dm.NameSlug, shared.SlugString(model), int16(year)),
		KSUID:      ksuid.New().String(),
		Model:      model,
		Year:       year,
		DeviceType: common.DefaultDeviceType,
		ImageURI:   "",
	}

	return dd
}

func SetupCreateDeviceDefinitionWithVehicleInfo(t *testing.T, dm models.DeviceMake, model string, year int, pdb db.Store) *gateways.DeviceDefinitionTablelandModel {
	dd := SetupCreateDeviceDefinition(t, dm, model, year, pdb)
	dd.Metadata = &gateways.DeviceDefinitionMetadata{
		DeviceAttributes: []gateways.DeviceTypeAttribute{
			{
				Name:  "fuel_type",
				Value: "defaultValue",
			},
			{
				Name:  "driven_wheels",
				Value: "4",
			},
			{
				Name:  "number_of_doors",
				Value: "4",
			},
			{
				Name:  "mpg",
				Value: "defaultValue",
			},
		},
	}

	return dd
}

func SetupCreateDeviceDefinitionWithVehicleInfoIncludePowerTrain(t *testing.T, dm models.DeviceMake, model string, year int, pdb db.Store) *gateways.DeviceDefinitionTablelandModel {
	dd := SetupCreateDeviceDefinition(t, dm, model, year, pdb)
	dd.Metadata = &gateways.DeviceDefinitionMetadata{
		DeviceAttributes: []gateways.DeviceTypeAttribute{
			{
				Name:  "fuel_type",
				Value: "defaultValue",
			},
			{
				Name:  "driven_wheels",
				Value: "4",
			},
			{
				Name:  "number_of_doors",
				Value: "4",
			},
			{
				Name:  "mpg",
				Value: "defaultValue",
			},
			{
				Name:  "powertrain_type",
				Value: "ICE",
			},
		},
	}

	return dd
}

func SetupCreateDeviceType(t *testing.T, pdb db.Store) *models.DeviceType {
	dt := &models.DeviceType{
		ID:          ksuid.New().String(),
		Name:        "vehicle",
		Metadatakey: "vehicle_info",
		Properties:  null.JSONFrom(deviceTypeVehiclePropertyDataSample),
	}
	err := dt.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return dt
}

func SetupCreateMake(t *testing.T, mk string, pdb db.Store) models.DeviceMake {
	dm := models.DeviceMake{
		ID:       ksuid.New().String(),
		Name:     mk,
		NameSlug: shared.SlugString(mk),
	}
	err := dm.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "no db error expected")
	return dm
}

func SetupCreateStyle(t *testing.T, definitionID string, name string, source string, subModel string, pdb db.Store) models.DeviceStyle {
	ds := models.DeviceStyle{
		ID:              ksuid.New().String(),
		Name:            name,
		DefinitionID:    definitionID,
		Source:          source,
		SubModel:        subModel,
		ExternalStyleID: ksuid.New().String(),
	}
	err := ds.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "no db error expected")
	return ds
}

func SetupCreateAutoPiIntegration(t *testing.T, pdb db.Store) *models.Integration {

	dMake := &models.DeviceMake{
		ID:       ksuid.New().String(),
		Name:     "AutoPi",
		NameSlug: "autopi",
	}

	err := dMake.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")

	integration := &models.Integration{
		ID:                  ksuid.New().String(),
		Type:                models.IntegrationTypeAPI,
		Style:               models.IntegrationStyleWebhook,
		Vendor:              common.AutoPiVendor,
		RefreshLimitSecs:    1800,
		Points:              6000,
		ManufacturerTokenID: null.IntFrom(144),
	}
	err = integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func SetupCreateWMI(t *testing.T, id string, deviceMakeID string, pdb db.Store) *models.Wmi {
	wmi := &models.Wmi{
		Wmi:          id,
		DeviceMakeID: deviceMakeID,
	}
	err := wmi.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return wmi
}

func SetupCreateSmartCarIntegration(t *testing.T, pdb db.Store) *models.Integration {
	dMake := &models.DeviceMake{
		ID:       ksuid.New().String(),
		Name:     "Smartcar",
		NameSlug: "smartcar",
	}

	err := dMake.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")

	integration := &models.Integration{
		ID:                  ksuid.New().String(),
		Type:                models.IntegrationTypeAPI,
		Style:               models.IntegrationStyleWebhook,
		Vendor:              common.SmartCarVendor,
		RefreshLimitSecs:    1800,
		Points:              6000,
		ManufacturerTokenID: null.IntFrom(143),
	}
	err = integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func SetupCreateHardwareIntegration(t *testing.T, pdb db.Store) *models.Integration {

	dMake := &models.DeviceMake{
		ID:       ksuid.New().String(),
		Name:     "Macaron",
		NameSlug: "macaron",
	}

	err := dMake.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")

	integration := &models.Integration{
		ID:                  ksuid.New().String(),
		Type:                models.IntegrationTypeHardware,
		Style:               models.IntegrationStyleAddon,
		Vendor:              "Hardware",
		RefreshLimitSecs:    1800,
		Points:              6000,
		ManufacturerTokenID: null.IntFrom(142),
	}
	err = integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func Logger() *zerolog.Logger {
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "device-definitions-api").
		Logger()
	return &l
}
