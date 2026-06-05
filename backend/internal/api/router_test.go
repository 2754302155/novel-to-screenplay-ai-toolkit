package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/config"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{
		Environment: "test",
		Port:        "8080",
		Version:     "0.1.0-test",
	})

	request := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}

	if body := response.Body.String(); body == "" || !contains(body, "novel-to-screenplay-api") {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func contains(value string, needle string) bool {
	return len(value) >= len(needle) && (value == needle || len(needle) == 0 || index(value, needle) >= 0)
}

func index(value string, needle string) int {
	for i := 0; i+len(needle) <= len(value); i++ {
		if value[i:i+len(needle)] == needle {
			return i
		}
	}

	return -1
}
