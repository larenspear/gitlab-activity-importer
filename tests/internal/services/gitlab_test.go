package services_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

func TestGetUsersProjectsIds(t *testing.T) {
	tests := []struct {
		name             string
		userId           int
		statusCode       int
		expectedResponse []map[string]interface{}
		expectedIds      []int
		expectError      bool
	}{
		{
			name:       "projects found",
			userId:     1,
			statusCode: 200,
			expectedResponse: []map[string]interface{}{
				{"id": 1, "name": "Project1"},
				{"id": 2, "name": "Project2"},
			},
			expectedIds: []int{1, 2},
			expectError: false,
		},
		{
			name:             "no projects found",
			userId:           2,
			statusCode:       200,
			expectedResponse: []map[string]interface{}{},
			expectedIds:      nil,
			expectError:      true,
		},
		{
			name:       "user not found",
			userId:     2,
			statusCode: 404,
			expectedResponse: []map[string]interface{}{
				{"message": "404 User Not Found"},
			},
			expectedIds: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BASE_URL", "http://test-url.com")
			os.Setenv("GITLAB_TOKEN", "test-token")
			defer os.Unsetenv("BASE_URL")
			defer os.Unsetenv("GITLAB_TOKEN")

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedURL := fmt.Sprintf("/api/v4/users/%d/contributed_projects", tt.userId)
				if r.URL.Path != expectedURL {
					t.Errorf("Expected URL '%s', got '%s'", expectedURL, r.URL.Path)
				}
				if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
					t.Errorf("Expected PRIVATE-TOKEN 'test-token', got '%s'", r.Header.Get("PRIVATE-TOKEN"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)

				responseData, err := json.Marshal(tt.expectedResponse)
				if err != nil {
					t.Fatalf("Failed to marshal response data: %v", err)
				}
				w.Write(responseData)
			}))
			defer mockServer.Close()

			os.Setenv("BASE_URL", mockServer.URL)

			result, err := services.GetUsersProjectsIds(tt.userId)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUsersProjectsIds returned error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expectedIds) {
				t.Errorf("Expected '%v', got '%v'", tt.expectedIds, result)
			}
		})
	}
}
