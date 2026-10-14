package queries

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// processDeviceStyle runs on every decode whose style name is at least two
// characters, which is nearly all of them. sqlboiler's One() returns a nil row
// together with a wrapped error for every failure that is not sql.ErrNoRows --
// a reader-pool blip, a reset connection, a context deadline (the caller's
// context carries one, and a catalog round trip runs just before this) -- so
// branching only on no-rows and then reading style.ID nil-dereferences and
// panics the decode. Handle is already written to tolerate an error from this
// function; it just never got one.
//
// No Postgres: sqlmock answers the device_styles reads.

const styleTestDefinitionID = "toyota_camry_2026"

func styleHandler(t *testing.T) (DecodeVINQueryHandler, sqlmock.Sqlmock) {
	t.Helper()
	return txHandler(t)
}

func styleVINInfo() *coremodels.VINDecodingInfoData {
	return &coremodels.VINDecodingInfoData{
		VIN:       "4T1C11AK8NU123456",
		Make:      "Toyota",
		Model:     "Camry",
		Year:      2026,
		StyleName: "LE",
		Source:    coremodels.DrivlyProvider,
	}
}

// The first lookup is by (definition_id, source, external_style_id). A failure
// that is not no-rows leaves the row nil, and neither no-rows branch fires.
func TestProcessDeviceStyleReportsAFailedExternalIDLookup(t *testing.T) {
	dc, mock := styleHandler(t)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(context.DeadlineExceeded)

	id, err := dc.processDeviceStyle(context.Background(), styleVINInfo(), styleTestDefinitionID, "ICE")

	require.Error(t, err, "a failed device_styles read must be reported, not dereferenced")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Empty(t, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// The second lookup is by name, reached only when the first found no row. It
// fails the same way, and used to reach the same nil dereference.
func TestProcessDeviceStyleReportsAFailedNameLookup(t *testing.T) {
	dc, mock := styleHandler(t)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(sql.ErrConnDone)

	id, err := dc.processDeviceStyle(context.Background(), styleVINInfo(), styleTestDefinitionID, "ICE")

	require.Error(t, err, "a failed device_styles read must be reported, not dereferenced")
	assert.ErrorIs(t, err, sql.ErrConnDone)
	assert.Empty(t, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// The ordinary hit: the style already exists under the external style id.
func TestProcessDeviceStyleReturnsAnExistingStyle(t *testing.T) {
	dc, mock := styleHandler(t)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "definition_id", "name", "external_style_id", "source"}).
			AddRow("26M1cyBpEVbz8Ejw5A7qKcKAaWP", styleTestDefinitionID, "LE", "le", "drivly"))

	id, err := dc.processDeviceStyle(context.Background(), styleVINInfo(), styleTestDefinitionID, "ICE")

	require.NoError(t, err)
	assert.Equal(t, "26M1cyBpEVbz8Ejw5A7qKcKAaWP", id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Found on the second lookup, by name: the first read reported no row.
func TestProcessDeviceStyleReturnsAStyleFoundByName(t *testing.T) {
	dc, mock := styleHandler(t)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "definition_id", "name", "external_style_id", "source"}).
			AddRow("26M1cyBpEVbz8Ejw5A7qKcKAaWQ", styleTestDefinitionID, "LE", "le-other", "vincario"))

	id, err := dc.processDeviceStyle(context.Background(), styleVINInfo(), styleTestDefinitionID, "ICE")

	require.NoError(t, err)
	assert.Equal(t, "26M1cyBpEVbz8Ejw5A7qKcKAaWQ", id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// A failed insert is already reported; keep it that way, and keep the error
// distinguishable from a failed read.
func TestProcessDeviceStyleReportsAFailedInsert(t *testing.T) {
	dc, mock := styleHandler(t)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT .* FROM .*device_styles.* WHERE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO .*device_styles.*`).
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	id, err := dc.processDeviceStyle(context.Background(), styleVINInfo(), styleTestDefinitionID, "ICE")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error creating style")
	assert.Empty(t, id)
}
