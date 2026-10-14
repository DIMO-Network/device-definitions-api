package queries

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/db/models"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/exceptions"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/DIMO-Network/shared/pkg/db"
	vinutils "github.com/DIMO-Network/shared/pkg/vin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// vinInfoFromKnown is the smartcar / software-connection fallback: the
// providers all failed, and the decode is rebuilt from the WMI plus a model
// and year the caller supplied. These need no Postgres -- makeFromWMIRows
// takes the rows, and the context test never gets as far as a connection.

const knownVIN = "1FTFW1E502FA00001"

// An unknown WMI is the case this fallback exists for. sqlboiler's .All()
// reports it as an empty slice with a nil error -- unlike .One(), which
// returns sql.ErrNoRows -- so indexing element zero panicked the decode
// handler.
func TestMakeFromWMIRowsReportsAnUnknownWMIAsNotFound(t *testing.T) {
	dc, _ := hydrateHandler(t)

	mk, err := dc.makeFromWMIRows(context.Background(), "1FT", nil, "F-150", 2020)

	require.Error(t, err)
	assert.Empty(t, mk)
	var notFound *exceptions.NotFoundError
	assert.ErrorAs(t, err, &notFound, "an unknown WMI must be a typed not found, not a panic")
}

// One row is the common case and still answers that marque.
func TestMakeFromWMIRowsAnswersTheOnlyManufacturer(t *testing.T) {
	dc, _ := hydrateHandler(t)

	mk, err := dc.makeFromWMIRows(context.Background(), "1FT", models.WmiSlice{{Wmi: "1FT", ManufacturerName: "Ford"}}, "F-150", 2020)

	require.NoError(t, err)
	assert.Equal(t, "Ford", mk)
}

// A WMI shared by several marques of one parent OEM resolves to the marque
// that already has a template for this model-year, and the catalog is asked
// with the caller's context.
func TestMakeFromWMIRowsPicksTheMarqueWithATemplateUsingTheCallersContext(t *testing.T) {
	dc, catalog := hydrateHandler(t)
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "callers-context")

	catalog.EXPECT().GetTemplateByID(gomock.Any(), "chrysler_town-and-country_2012").
		DoAndReturn(func(got context.Context, _ string) (*coremodels.Template, *big.Int, error) {
			assert.Equal(t, "callers-context", got.Value(ctxKey{}), "the catalog must be called with the caller's context")
			return nil, nil, fmt.Errorf("not in the catalog")
		})
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "dodge_town-and-country_2012").
		DoAndReturn(func(got context.Context, _ string) (*coremodels.Template, *big.Int, error) {
			assert.Equal(t, "callers-context", got.Value(ctxKey{}), "the catalog must be called with the caller's context")
			return hydrateTemplate(), nil, nil
		})

	mk, err := dc.makeFromWMIRows(ctx, "2C4", models.WmiSlice{
		{Wmi: "2C4", ManufacturerName: "Chrysler"},
		{Wmi: "2C4", ManufacturerName: "Dodge"},
	}, "Town and Country", 2012)

	require.NoError(t, err)
	assert.Equal(t, "Dodge", mk)
}

// ctxRecordingConnector is a database/sql connector that records the context
// the query was issued with and then refuses to connect, so a query's context
// is observable without a database.
type ctxRecordingConnector struct{ got chan context.Context }

func (c *ctxRecordingConnector) Connect(ctx context.Context) (driver.Conn, error) {
	select {
	case c.got <- ctx:
	default:
	}
	return nil, fmt.Errorf("no database in this test")
}

func (*ctxRecordingConnector) Driver() driver.Driver { return nil }

// A cancelled decode must stop holding a database connection: the wmis query
// ran on context.Background(), so it outlived the request that asked for it.
func TestVinInfoFromKnownQueriesWithTheCallersContext(t *testing.T) {
	conn := &ctxRecordingConnector{got: make(chan context.Context, 1)}
	sqlDB := sql.OpenDB(conn)
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	logger := zerolog.Nop()
	dc := DecodeVINQueryHandler{
		logger:                         &logger,
		deviceDefinitionCatalogService: mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl),
		dbs: func() *db.ReaderWriter {
			return &db.ReaderWriter{Reader: &db.DB{DB: sqlDB}, Writer: &db.DB{DB: sqlDB}}
		},
	}

	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "callers-context")
	_, err := dc.vinInfoFromKnown(ctx, vinutils.VIN(knownVIN), "F-150", 2020)
	require.Error(t, err, "the fake connector refuses, so the query must fail")

	select {
	case got := <-conn.got:
		assert.Equal(t, "callers-context", got.Value(ctxKey{}), "the wmis query must run on the caller's context, not context.Background()")
	default:
		t.Fatal("no database connection was attempted")
	}
}
