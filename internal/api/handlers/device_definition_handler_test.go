package handlers

import (
	"io"
	"net/http"
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	"github.com/DIMO-Network/device-definitions-api/internal/core/queries"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	mock_gateways "github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// GET /device-definitions/:id end to end over the pieces that decide its
// status code: the mediator panics with the handler's error, fiber's recover
// hands that value to the error handler untouched when it is an error, and
// FiberConfig maps it by type assertion. A wrapped sentinel satisfies none of
// that and falls through to 500.
func definitionByIDApp(t *testing.T, catalogErr error) *fiber.App {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	catalog := mock_gateways.NewMockDeviceDefinitionCatalogService(ctrl)
	catalog.EXPECT().GetTemplateByID(gomock.Any(), "ford_escapexx_2020").Return(nil, nil, catalogErr)

	m, err := mediator.New(
		mediator.WithHandler(&queries.GetDeviceDefinitionByIDQueryV2{}, queries.NewGetDeviceDefinitionByIDQueryV2Handler(catalog, nil)),
	)
	require.NoError(t, err)

	app := fiber.New(common.FiberConfig(true))
	app.Use(recover.New())
	app.Get("/device-definitions/:id", GetDeviceDefinitionByID(*m))
	return app
}

func TestGetDeviceDefinitionByIDIsA404ForAnUnknownID(t *testing.T) {
	app := definitionByIDApp(t, errors.Wrapf(gateways.ErrTemplateNotFound, "template %q", "ford_escapexx_2020"))

	resp, err := app.Test(BuildRequest("GET", "/device-definitions/ford_escapexx_2020", ""))
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body) //nolint

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "an id with no template is not found, not a server error: %s", body)
}

func TestGetDeviceDefinitionByIDIsA500WhenTheCatalogIsUnreachable(t *testing.T) {
	app := definitionByIDApp(t, errors.New("unexpected status 502 from catalog"))

	resp, err := app.Test(BuildRequest("GET", "/device-definitions/ford_escapexx_2020", ""))
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body) //nolint

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "a catalog outage must not read as a missing definition: %s", body)
}
