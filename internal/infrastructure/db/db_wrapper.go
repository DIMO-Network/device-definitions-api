package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq" // concrete implementation of database
	"github.com/pkg/errors"
)

// ConnectOptions config options for database
type ConnectOptions struct {
	DSN                string
	DriverName         string
	Retries            int
	RetryDelay         time.Duration
	ConnectTimeout     time.Duration
	ConnMaxLifetime    time.Duration
	MaxIdleConnections int
	MaxOpenConnections int
}

// DB type to wrap sql.DB
type DB struct {
	*sql.DB
}

// Tx type to wrap sql.Tx
type Tx struct {
	*sql.Tx
}

// ReaderWriter wraps db reader and writer
type ReaderWriter struct {
	Reader *DB
	Writer *DB
}

var dbs *ReaderWriter

// NewDbConnection connects to the reader and writer per passed in options, with retries, returning a DBReaderWriter object that contains sql.DB connection
func NewDbConnection(ctx context.Context, ready *bool, ro ConnectOptions, wo ConnectOptions) *ReaderWriter {
	dbs = &ReaderWriter{Reader: &DB{}, Writer: &DB{}}

	go func(ctx context.Context, ready *bool, dbs *ReaderWriter) {
		errCh := make(chan error)
		defer close(errCh)

		readyCh := make(chan bool)
		defer close(readyCh)
		// could add open census tracing: https://github.com/opencensus-integrations/ocsql

		var wg sync.WaitGroup

		rCtx, rCancel := context.WithCancel(ctx)
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, ec chan error, rc chan bool) {
			defer wg.Done()
			dbs.Reader.DB = connectWithRetry(rCtx, ec, rc, ro)
		}(rCtx, &wg, errCh, readyCh)

		wCtx, wCancel := context.WithCancel(ctx)
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, ec chan error, rc chan bool) {
			defer wg.Done()
			dbs.Writer.DB = connectWithRetry(wCtx, ec, rc, wo)
		}(wCtx, &wg, errCh, readyCh)

		cancelFunc := func(r bool) {
			rCancel()
			wCancel()
			wg.Wait()
			*ready = r
		}

		rCount := 0

		for {
			select {
			case ir := <-readyCh:
				if ir {
					rCount++
					if rCount >= 2 {
						cancelFunc(true)
						return
					}
					continue
				}
			case <-errCh:
				cancelFunc(false)
				return
			case <-time.After(ro.ConnectTimeout):
				cancelFunc(false)
				return
			case <-time.After(wo.ConnectTimeout):
				cancelFunc(false)
				return
			case <-ctx.Done():
				cancelFunc(false)
				return
			}
		}

	}(ctx, ready, dbs)

	return dbs
}

func connectWithRetry(ctx context.Context, errCh chan error, readyCh chan bool, opts ConnectOptions) (db *sql.DB) {
	var (
		err error
		try = 0
	)

loop:
	for {
		try++
		if opts.Retries > 0 && opts.Retries <= try {
			err = errors.Errorf("could not connect to db, tries=%d", try)
			break
		}

		db, err = connect(opts)
		if err != nil {
			fmt.Printf("can't connect to db, dsn=%s, err=%s, tries=%d", opts.DSN, err, try)

			select {
			case <-ctx.Done():
				break loop
			case <-time.After(opts.RetryDelay):
				continue
			}
		}

		break loop
	}

	if err != nil {
		errCh <- err
		return
	}

	readyCh <- true
	return
}

func connect(opts ConnectOptions) (*sql.DB, error) {
	db, err := sql.Open(opts.DriverName, opts.DSN)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(opts.MaxIdleConnections)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)
	db.SetMaxOpenConns(opts.MaxOpenConnections)

	return db, nil
}

// GetReaderConn returns connection to reader
func (dbs *ReaderWriter) GetReaderConn() *sql.DB {
	return dbs.Reader.DB
}

// GetWriterConn returns connection to writer
func (dbs *ReaderWriter) GetWriterConn() *sql.DB {
	return dbs.Writer.DB
}
