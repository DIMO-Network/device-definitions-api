package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetSmartcarManufacturers(t *testing.T) {
	app := fiber.New()

	// Register the handler
	app.Get("/manufacturers", GetSmartcarManufacturers())
	req := BuildRequest("GET", "/manufacturers", "")
	resp, err := app.Test(req)

	assert.NoError(t, err, "Expected no error on request")

	all, _ := io.ReadAll(resp.Body) //nolint
	// Check if the status code is 200 OK
	if !assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK") {
		fmt.Println(string(all))
	}
}

func BuildRequest(method, url, body string) *http.Request {
	req, _ := http.NewRequest(
		method,
		url,
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	return req
}
