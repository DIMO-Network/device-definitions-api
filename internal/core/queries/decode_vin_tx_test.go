package queries

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Handle reads the cached vin_numbers row inside a SERIALIZABLE transaction on
// the writer, and used to keep that transaction open across
// hydrateResponseFromVinNumber. That function was pure in-memory when the
// transaction was placed around it; it now issues a catalog GET through a
// client with a 30 second timeout and a device_styles query against the
// reader, and returns early on a cache hit -- the path the code itself calls
// "the one most decodes take". So the commonest decode held a writer
// connection and a serializable snapshot across a full CDN round trip, and
// took a reader connection while holding it: a two second catalog stall pins
// every writer connection for two seconds per decode, and a thirty second
// stall exhausts the pool outright while unrelated writes queue behind it.
//
// The fix is to read the row, end the transaction, then hydrate. These pin the
// first half of that: readCachedVinNumber has ended its transaction by the
// time it returns, on every path -- a row, no row, a failed query, a failed
// begin. Nothing slow can then run inside it, because there is no inside.
//
// No Postgres: sqlmock answers the read and records that the rollback happened
// before the call returned.

func mockDBS(t *testing.T) (func() *db.ReaderWriter, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	rw := &db.ReaderWriter{Reader: &db.DB{DB: sqlDB}, Writer: &db.DB{DB: sqlDB}}
	return func() *db.ReaderWriter { return rw }, mock
}

func txHandler(t *testing.T) (DecodeVINQueryHandler, sqlmock.Sqlmock) {
	t.Helper()
	dbs, mock := mockDBS(t)
	logger := zerolog.Nop()
	return DecodeVINQueryHandler{dbs: dbs, logger: &logger}, mock
}

const txTestVIN = "4T1C11AK8NU123456"

func TestReadCachedVinNumberEndsItsTransactionBeforeItReturns(t *testing.T) {
	dc, mock := txHandler(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM .*vin_numbers.* WHERE`).
		WithArgs(txTestVIN).
		WillReturnRows(sqlmock.NewRows([]string{"vin", "definition_id", "year", "manufacturer_name"}).
			AddRow(txTestVIN, "toyota_camry_2026", 2026, "Toyota"))
	mock.ExpectRollback()

	vn, err := dc.readCachedVinNumber(context.Background(), txTestVIN)

	require.NoError(t, err)
	require.NotNil(t, vn)
	assert.Equal(t, "toyota_camry_2026", vn.DefinitionID)
	// The whole point: by the time the caller has the row, the writer
	// connection and the serializable snapshot are already released, so the
	// catalog round trip that follows holds neither.
	assert.NoError(t, mock.ExpectationsWereMet(), "the transaction must be closed before hydration starts")
}

// A VIN that has never been decoded is the path that goes on to do the whole
// decode -- several HTTP calls to vendors -- so leaving the transaction open
// here would be the worst case of all.
func TestReadCachedVinNumberEndsItsTransactionWhenThereIsNoRow(t *testing.T) {
	dc, mock := txHandler(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM .*vin_numbers.* WHERE`).
		WithArgs(txTestVIN).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	vn, err := dc.readCachedVinNumber(context.Background(), txTestVIN)

	require.NoError(t, err, "no cached row is not an error, it is the uncached decode")
	assert.Nil(t, vn)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Not just the happy path: a failed read must not leak the transaction either.
func TestReadCachedVinNumberEndsItsTransactionWhenTheQueryFails(t *testing.T) {
	dc, mock := txHandler(t)
	boom := errors.New("connection reset by peer")
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM .*vin_numbers.* WHERE`).
		WithArgs(txTestVIN).
		WillReturnError(boom)
	mock.ExpectRollback()

	vn, err := dc.readCachedVinNumber(context.Background(), txTestVIN)

	require.Error(t, err)
	assert.Nil(t, vn)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// A transaction that never began has nothing to close, and must not be
// reported as a decode failure of some other kind.
func TestReadCachedVinNumberReportsAFailedBegin(t *testing.T) {
	dc, mock := txHandler(t)
	boom := errors.New("too many connections")
	mock.ExpectBegin().WillReturnError(boom)

	vn, err := dc.readCachedVinNumber(context.Background(), txTestVIN)

	require.Error(t, err)
	assert.Nil(t, vn)
	assert.ErrorIs(t, err, boom)
	assert.NoError(t, mock.ExpectationsWereMet())
}
