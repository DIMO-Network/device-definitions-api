package dbtest

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"testing"
	"time"

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
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DB.User, settings.DB.Password, port.Port(), settings.DB.Name)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:12.9-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DB.User, "POSTGRES_PASSWORD": settings.DB.Password, "POSTGRES_DB": settings.DB.Name},
		ExposedPorts: []string{pgPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
		WaitingFor:   wait.ForSQL(nat.Port(pgPort), "postgres", dbURL).Timeout(time.Second * 15),
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
	logger.Info().Msgf("set default search_path for user postgres to %s", dbName)
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
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
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

func SetupCreateDeviceDefinition(t *testing.T, dm models.DeviceMake, model string, year int, pdb db.Store) *models.DeviceDefinition {
	dt := SetupCreateDeviceType(t, pdb)
	dd := &models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: dm.ID,
		Model:        model,
		Year:         int16(year),
		Verified:     true,
		DeviceTypeID: null.StringFrom(dt.ID),
		ModelSlug:    common.SlugString(model),
	}
	err := dd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm
	dd.R.DeviceType = dt

	return dd
}

func SetupCreateDeviceDefinitionWithVehicleInfo(t *testing.T, dm models.DeviceMake, model string, year int, pdb db.Store) *models.DeviceDefinition {
	dt := SetupCreateDeviceType(t, pdb)
	dd := &models.DeviceDefinition{
		ID:           ksuid.New().String(),
		DeviceMakeID: dm.ID,
		Model:        model,
		Year:         int16(year),
		Verified:     true,
		DeviceTypeID: null.StringFrom(dt.ID),
		ModelSlug:    common.SlugString(model),
	}

	deviceTypeInfo := make(map[string]interface{})
	metaData := make(map[string]interface{})
	var ai map[string][]interface{}
	defaultValue := "defaultValue"
	if err := dt.Properties.Unmarshal(&ai); err == nil {
		metaData["fuel_type"] = defaultValue
		metaData["driven_wheels"] = "4"
		metaData["number_of_doors"] = "4"
		metaData["MPG"] = defaultValue
	}
	deviceTypeInfo[dt.Metadatakey] = metaData
	_ = dd.Metadata.Marshal(deviceTypeInfo)

	err := dd.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")

	dd.R = dd.R.NewStruct()
	dd.R.DeviceMake = &dm
	dd.R.DeviceType = dt

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
		NameSlug: common.SlugString(mk),
	}
	err := dm.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "no db error expected")
	return dm
}

func SetupCreateStyle(t *testing.T, deviceDefinitionID string, name string, source string, subModel string, pdb db.Store) models.DeviceStyle {
	ds := models.DeviceStyle{
		ID:                 ksuid.New().String(),
		Name:               name,
		DeviceDefinitionID: deviceDefinitionID,
		Source:             source,
		SubModel:           subModel,
		ExternalStyleID:    ksuid.New().String(),
	}
	err := ds.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "no db error expected")
	return ds
}

func SetupCreateAutoPiIntegration(t *testing.T, pdb db.Store) *models.Integration {
	integration := &models.Integration{
		ID:               ksuid.New().String(),
		Type:             models.IntegrationTypeAPI,
		Style:            models.IntegrationStyleWebhook,
		Vendor:           common.AutoPiVendor,
		RefreshLimitSecs: 1800,
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func SetupCreateSmartCarIntegration(t *testing.T, pdb db.Store) *models.Integration {
	integration := &models.Integration{
		ID:               ksuid.New().String(),
		Type:             models.IntegrationTypeAPI,
		Style:            models.IntegrationStyleWebhook,
		Vendor:           common.SmartCarVendor,
		RefreshLimitSecs: 1800,
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func SetupCreateHardwareIntegration(t *testing.T, pdb db.Store) *models.Integration {
	integration := &models.Integration{
		ID:               ksuid.New().String(),
		Type:             models.IntegrationTypeHardware,
		Style:            models.IntegrationStyleAddon,
		Vendor:           "Hardware",
		RefreshLimitSecs: 1800,
	}
	err := integration.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err, "database error")
	return integration
}

func SetupCreateDeviceIntegration(t *testing.T, dd *models.DeviceDefinition, integrationID string, pdb db.Store) *models.DeviceIntegration {
	di := &models.DeviceIntegration{
		DeviceDefinitionID: dd.ID,
		IntegrationID:      integrationID,
		Region:             "Americas",
	}
	err := di.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	return di
}

func SetupIntegrationFeature(t *testing.T, pdb db.Store) *models.IntegrationFeature {
	feature := &models.IntegrationFeature{
		FeatureKey:      ksuid.New().String(),
		DisplayName:     ksuid.New().String(),
		ElasticProperty: ksuid.New().String(),
		FeatureWeight:   null.Float64From(1),
		CSSIcon:         null.StringFrom("css"),
	}
	err := feature.Insert(context.Background(), pdb.DBS().Writer, boil.Infer())
	require.NoError(t, err)
	return feature
}

func Logger() *zerolog.Logger {
	l := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "device-definitions-api").
		Logger()
	return &l
}
