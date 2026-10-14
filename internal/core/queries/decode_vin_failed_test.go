package queries

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	mock_services "github.com/DIMO-Network/device-definitions-api/internal/core/services/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// A decode that fails has to record itself in failed_vin_decodes: that row is
// what stops the next request for the same VIN from paying for the whole
// vendor fan-out again, and it is the only record that the VIN was ever tried.
//
// GetVIN returns (nil, nil, err) on several paths -- the 0SC test-VIN branch
// now reads a template from the catalog and returns its error verbatim, and
// the invalid-VIN guard returns before any vendor is tried -- so the vendor
// extra the failure branch reads can be nil. It used to be read
// unconditionally, so the failure branch panicked, wrote no row, and the same
// VIN failed the same way on every retry forever.
//
// No Postgres: sqlmock answers the reads and records the insert.

const failedTestVIN = "4T1C11AK8NU123456"

func failedDecodeHandler(t *testing.T) (DecodeVINQueryHandler, sqlmock.Sqlmock, *mock_services.MockVINDecodingService) {
	t.Helper()
	dbs, mock := mockDBS(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	vinDecoding := mock_services.NewMockVINDecodingService(ctrl)
	logger := zerolog.Nop()
	return DecodeVINQueryHandler{dbs: dbs, logger: &logger, vinDecodingService: vinDecoding}, mock, vinDecoding
}

// expectDecodeUpToProviders queues every read Handle makes before it calls the
// decoding service: the cached vin_numbers read (in its own transaction), the
// failed_vin_decodes check, the device_types read and the wmis read.
func expectDecodeUpToProviders(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM .*vin_numbers.* WHERE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM .*failed_vin_decodes.*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT .* FROM .*device_types.* WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("vehicle", "Vehicle"))
	mock.ExpectQuery(`SELECT .* FROM .*wmis.* WHERE`).
		WillReturnError(sql.ErrNoRows)
}

// The regression: every vendor failed and the decoding service reported no
// vendor extra at all.
func TestHandleRecordsAFailedDecodeWhenTheVendorExtraIsNil(t *testing.T) {
	dc, mock, vinDecoding := failedDecodeHandler(t)
	expectDecodeUpToProviders(mock)
	vinDecoding.EXPECT().GetVIN(gomock.Any(), failedTestVIN, coremodels.AllProviders, "USA").
		Return(nil, nil, errors.New("invalid vin"))

	mock.ExpectQuery(`INSERT INTO .*failed_vin_decodes.*`).
		WillReturnRows(sqlmock.NewRows([]string{"vin"}).AddRow(failedTestVIN))

	resp, err := dc.Handle(context.Background(), &DecodeVINQuery{VIN: failedTestVIN, Country: "USA"})

	require.Error(t, err, "a decode with no vendor extra must report the decode failure, not panic")
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid vin")
	// The failure has to be recorded, or the next request repeats the whole
	// vendor fan-out and fails identically.
	assert.NoError(t, mock.ExpectationsWereMet(), "the failed_vin_decodes row must still be written")
}

// The ordinary failure: the vendors were tried and said no, so the extra
// carries what each of them answered. That detail must survive into the row.
func TestHandleRecordsTheVendorsTriedOnAFailedDecode(t *testing.T) {
	dc, mock, vinDecoding := failedDecodeHandler(t)
	expectDecodeUpToProviders(mock)
	vinDecoding.EXPECT().GetVIN(gomock.Any(), failedTestVIN, coremodels.AllProviders, "USA").
		Return(nil, &coremodels.VINDecodingVendorExtra{VendorsTried: []string{"drivly", "vincario"}}, errors.New("no vendor could decode"))

	mock.ExpectQuery(`INSERT INTO .*failed_vin_decodes.*`).
		WillReturnRows(sqlmock.NewRows([]string{"vin"}).AddRow(failedTestVIN))

	_, err := dc.Handle(context.Background(), &DecodeVINQuery{VIN: failedTestVIN, Country: "USA"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no vendor could decode")
	assert.NoError(t, mock.ExpectationsWereMet(), "the failed_vin_decodes row must still be written")
}
