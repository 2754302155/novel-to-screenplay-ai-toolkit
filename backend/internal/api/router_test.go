package api

import (
	"bytes"
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

func TestParseChapters(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"text":"第一章\n正文。\n\n第二章\n正文。\n\n第三章\n正文。"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/chapters/parse", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}

	if body := response.Body.String(); !contains(body, `"id":"CH001"`) {
		t.Fatalf("expected parsed chapter in response, got %s", body)
	}
}

func TestParseChaptersBlocksInsufficientInput(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"text":"第一章\n正文。"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/chapters/parse", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", response.Code, response.Body.String())
	}

	if body := response.Body.String(); !contains(body, "CHAPTER_COUNT_TOO_LOW") {
		t.Fatalf("expected chapter count error, got %s", body)
	}
}

func TestCreateConversionTask(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"source_text":"正文","ai_config":{"provider":"openai_compatible","base_url":"https://api.example.com/v1","model":"demo","api_key":"secret-key"},"chapters":[{"id":"CH001","title":"第一章","word_count":10},{"id":"CH002","title":"第二章","word_count":10},{"id":"CH003","title":"第三章","word_count":10}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/conversion-tasks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if body := response.Body.String(); !contains(body, `"status":"pending"`) || !contains(body, `"progress":5`) {
		t.Fatalf("expected pending task, got %s", body)
	}
	if body := response.Body.String(); contains(body, "secret-key") {
		t.Fatalf("api key leaked in response: %s", body)
	}
}

func TestListConversionTasks(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"source_text":"正文","ai_config":{"provider":"openai_compatible","base_url":"https://api.example.com/v1","model":"demo","api_key":"secret-key"},"chapters":[{"id":"CH001","title":"第一章","word_count":10},{"id":"CH002","title":"第二章","word_count":10},{"id":"CH003","title":"第三章","word_count":10}]}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/api/conversion-tasks", bytes.NewReader(body))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", createResponse.Code, createResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/conversion-tasks", nil)
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", listResponse.Code, listResponse.Body.String())
	}
	if body := listResponse.Body.String(); !contains(body, `"tasks"`) || !contains(body, `"status":"pending"`) {
		t.Fatalf("expected task list, got %s", body)
	}
	if body := listResponse.Body.String(); contains(body, "secret-key") {
		t.Fatalf("api key leaked in response: %s", body)
	}
}

func TestCreateConversionTaskBlocksInsufficientChapters(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"source_text":"正文","chapters":[{"id":"CH001","title":"第一章","word_count":10}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/conversion-tasks", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAITestRejectsIncompleteConfig(t *testing.T) {
	router := NewRouter(config.Config{Environment: "test", Version: "0.1.0-test"})
	body := []byte(`{"provider":"openai_compatible","base_url":"https://api.example.com/v1"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/ai/test", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", response.Code, response.Body.String())
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
