package services_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/furmanp/gitlab-activity-importer/internal/services"
)

func TestGetGitlabUser(t *testing.T) {
	os.Setenv("BASE_URL", "http://test-url.com")
	os.Setenv("GITLAB_TOKEN", "test-token")
	defer os.Unsetenv("BASE_URL")
	defer os.Unsetenv("GITLAB_TOKEN")

	expectedResponse := `{"username":"testuser", "id":"1"}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/api/v4/user" {
			t.Errorf("Expected URL '/api/v4/user', got '%s'", r.URL.String())
		}
		if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
			t.Errorf("Expected PRIVATE-TOKEN 'test-token', got '%s'", r.Header.Get("PRIVATE-TOKEN"))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponse)
	}))
	defer mockServer.Close()

	os.Setenv("BASE_URL", mockServer.URL)

	result := services.GetGitlabUser()

	if result != expectedResponse {
		t.Errorf("Expected '%s', got '%s'", expectedResponse, result)
	}
}
