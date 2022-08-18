package db

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/DIMO-Network/poc-dimo-api/device-definitions-api/internal/config"
)

const databaseDriver = "postgres"

// instance holds a single instance of the database
var instance *DBReaderWriter

var ready bool

// once is used to ensure that there is only a single instance of the database
var once sync.Once

// DbStore holds the database connection and other stuff.
type DbStore struct {
	db    func() *sql.DB
	dbs   *DBReaderWriter
	ready *bool
}

// NewDbConnectionFromSettings sets up a db connection from the settings, only once
func NewDbConnectionFromSettings(ctx context.Context, settings *config.Settings, withSearchPath bool) DbStore {
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

	return DbStore{db: instance.GetWriterConn, dbs: instance, ready: &ready}
}

// IsReady returns if db is ready to connect to
func (store *DbStore) IsReady() bool {
	return *store.ready
}

// DBS returns the reader and writer databases to connect to
func (store *DbStore) DBS() *DBReaderWriter {
	return store.dbs
}

// NewDbConnectionForTest use this for tests as we have multiple sessions in parallel and don't want synced one
func NewDbConnectionForTest(ctx context.Context, settings config.Settings, withSearchPath bool) DbStore {
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
	return DbStore{db: dbConnection.GetWriterConn, dbs: dbConnection, ready: &localReady}
}
