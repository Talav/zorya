package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talav/zorya"
)

func TestFiberAdapter_HeadersPropagate(t *testing.T) {
	app := fiber.New()
	api := zorya.NewAPI(NewFiber(app))

	type Out struct {
		Body struct {
			Value string `json:"value"`
		} `body:"structured"`
	}

	zorya.Get(api, "/test", func(ctx context.Context, _ *struct{}) (*Out, error) {
		return &Out{Body: struct {
			Value string `json:"value"`
		}{Value: "ok"}}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	ct := resp.Header.Get("Content-Type")
	assert.NotEmpty(t, ct, "Content-Type must be set by handler via w.Header().Set")
	assert.Contains(t, ct, "json", "response should be JSON")
}
