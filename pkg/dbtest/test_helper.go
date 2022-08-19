package test_db_helper

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartContainerDatabase starts postgres container with default test settings, and migrates the db. Caller must terminate container.
func StartContainerDatabase(ctx context.Context, dbName string, t *testing.T, migrationsDirRelPath string) (db.DbStore, testcontainers.Container) {
	settings := getTestDbSettings(dbName)
	pgPort := "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", settings.DBUser, settings.DBPassword, port.Port(), settings.DBName)
	}
	cr := testcontainers.ContainerRequest{
		Image:        "postgres:12.9-alpine",
		Env:          map[string]string{"POSTGRES_USER": settings.DBUser, "POSTGRES_PASSWORD": settings.DBPassword, "POSTGRES_DB": settings.DBName},
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

	settings.DBPort = mappedPort.Port()
	pdb := db.NewDbConnectionForTest(ctx, settings, false)
	for !pdb.IsReady() {
		time.Sleep(500 * time.Millisecond)
	}
	// can't connect to db, dsn=user=postgres password=postgres dbname=devices_api host=localhost port=49395 sslmode=disable search_path=devices_api, err=EOF
	// error happens when calling here
	_, err = pdb.DBS().Writer.Exec(`
		grant usage on schema public to public;
		grant create on schema public to public;
		CREATE SCHEMA IF NOT EXISTS devices_api;
		ALTER USER postgres SET search_path = devices_api, public;
		SET search_path = devices_api, public;
		`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrapf(err, "failed to apply schema. session: %s, port: %s",
			pgContainer.SessionID(), mappedPort.Port()), pgContainer, t)
	}
	// add truncate tables func
	_, err = pdb.DBS().Writer.Exec(`
CREATE OR REPLACE FUNCTION truncate_tables() RETURNS void AS $$
DECLARE
    statements CURSOR FOR
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'devices_api' and tablename != 'migrations';
BEGIN
    FOR stmt IN statements LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
`)
	if err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to create truncate func"), pgContainer, t)
	}

	goose.SetTableName("devices_api.migrations")
	if err := goose.Run("up", pdb.DBS().Writer.DB, migrationsDirRelPath); err != nil {
		return handleContainerStartErr(ctx, errors.Wrap(err, "failed to apply goose migrations for test"), pgContainer, t)
	}

	return pdb, pgContainer
}

// getTestDbSettings builds test db config.Settings object
func getTestDbSettings(dbName string) config.Settings {
	settings := config.Settings{
		LogLevel:             "info",
		DBName:               dbName,
		DBHost:               "localhost",
		DBPort:               "6669",
		DBUser:               "postgres",
		DBPassword:           "postgres",
		DBMaxOpenConnections: 2,
		DBMaxIdleConnections: 2,
		ServiceName:          "devices-api",
	}
	return settings
}

func handleContainerStartErr(ctx context.Context, err error, container testcontainers.Container, t *testing.T) (db.DbStore, testcontainers.Container) {
	if err != nil {
		fmt.Println("start container error: " + err.Error())
		if container != nil {
			container.Terminate(ctx) //nolint
		}
		t.Fatal(err)
	}
	return db.DbStore{}, container
}

// TruncateTables truncates tables for the test db, useful to run as teardown at end of each DB dependent test.
func TruncateTables(db *sql.DB, t *testing.T) {
	_, err := db.Exec(`SELECT truncate_tables();`)
	if err != nil {
		fmt.Println("truncating tables failed.")
		t.Fatal(err)
	}
}
