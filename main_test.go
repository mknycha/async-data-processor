package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageRoute(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		cfg := Config{}
		router, err := setupRouter(cfg)
		if err != nil {
			t.Fatalf("failed to setup router: %s", err.Error())
		}

		w := httptest.NewRecorder()
		reqBody := `{
			"timestamp":"2019-10-12T07:20:50.52Z",
			"value":"B"
		}`
		req, err := http.NewRequest("POST", "/message", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("failed to create request: %s", err.Error())
		}
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, `{"timestamp":"2019-10-12T07:20:50.52Z","value":"B"}`, w.Body.String())
	})

	t.Run("invalid request", func(t *testing.T) {
		cfg := Config{}
		router, err := setupRouter(cfg)
		if err != nil {
			t.Fatalf("failed to setup router: %s", err.Error())
		}

		w := httptest.NewRecorder()
		reqBody := `{
			"timestamp":"2019-10-12T",
			"value":"B"
		}`
		req, err := http.NewRequest("POST", "/message", strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("failed to create request: %s", err.Error())
		}
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"parsing time \"2019-10-12T\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"15\""}`, w.Body.String())
	})
}
