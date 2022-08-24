package db

import (
	"context"
	"database/sql"
	"github.com/rs/zerolog"
	"sync"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
)

const databaseDriver = "postgres"

// instance holds a single instance of the database
var instance *ReaderWriter

var ready bool

// once is used to ensure that there is only a single instance of the database
var once sync.Once

// Store holds the database connection and other stuff.
type Store struct {
	db    func() *sql.DB
	dbs   *ReaderWriter
	ready *bool
}

// NewDbConnectionFromSettings sets up a db connection from the settings, only once
func NewDbConnectionFromSettings(ctx context.Context, settings *config.Settings, withSearchPath bool) Store {
	once.Do(func() {
		instance = NewDbConnection(
			ctx,
			&ready,
			ConnectOptions{
				Retries:            5,
				RetryDelay:         time.Second * 10,
				ConnectTimeout:     time.Minute * 5,
				DSN:                settings.GetWriterDSN(withSearchPath),
				MaxOpenConnections: settings.DBMaxOpenConnections,
				MaxIdleConnections: settings.DBMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
				DriverName:         databaseDriver,
			},
			ConnectOptions{
				Retries:            5,
				RetryDelay:         time.Second * 10,
				ConnectTimeout:     time.Minute * 5,
				DSN:                settings.GetWriterDSN(true),
				MaxOpenConnections: settings.DBMaxOpenConnections,
				MaxIdleConnections: settings.DBMaxIdleConnections,
				ConnMaxLifetime:    time.Minute * 5,
				DriverName:         databaseDriver,
			},
		)
	})

	return Store{db: instance.GetWriterConn, dbs: instance, ready: &ready}
}

// IsReady returns if db is ready to connect to
func (store *Store) IsReady() bool {
	return *store.ready
}

// DBS returns the reader and writer databases to connect to
func (store *Store) DBS() *ReaderWriter {
	return store.dbs
}

// NewDbConnectionForTest use this for tests as we have multiple sessions in parallel and don't want synced one
func NewDbConnectionForTest(ctx context.Context, settings config.Settings, withSearchPath bool) Store {
	localReady := false
	dbConnection := NewDbConnection(
		ctx,
		&localReady,
		ConnectOptions{
			Retries:            5,
			RetryDelay:         time.Second * 10,
			ConnectTimeout:     time.Minute * 5,
			DSN:                settings.GetWriterDSN(withSearchPath),
			MaxOpenConnections: settings.DBMaxOpenConnections,
			MaxIdleConnections: settings.DBMaxIdleConnections,
			ConnMaxLifetime:    time.Minute * 5,
			DriverName:         databaseDriver,
		},
		ConnectOptions{
			Retries:            5,
			RetryDelay:         time.Second * 10,
			ConnectTimeout:     time.Minute * 5,
			DSN:                settings.GetWriterDSN(true),
			MaxOpenConnections: settings.DBMaxOpenConnections,
			MaxIdleConnections: settings.DBMaxIdleConnections,
			ConnMaxLifetime:    time.Minute * 5,
			DriverName:         databaseDriver,
		},
	)
	return Store{db: dbConnection.GetWriterConn, dbs: dbConnection, ready: &localReady}
}

// WaitForDB waits 30 seconds for db to become available, and Fatal panic if can't connect.
func (store *Store) WaitForDB(logger zerolog.Logger) {
	totalTime := 0
	for !store.IsReady() {
		if totalTime > 30 {
			logger.Fatal().Msg("could not connect to postgres after 30 seconds")
		}
		time.Sleep(time.Second)
		totalTime++
	}
}
