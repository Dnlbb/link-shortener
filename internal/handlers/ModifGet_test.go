package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dnlbb/link-shortener/internal/models"
)

func TestModifFget(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
		owner          string
	}{
		{
			name: "#1 Valid short URL",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://practicum.yandex.ru",
			owner:          "user",
		},
		{
			name:           "#2 Empty request body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Error: empty request body",
			owner:          "user",
		},
		{
			name:           "#3 Invalid JSON structure",
			requestBody:    `{ "Body": "http://localhost:8080/"`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "An anmarshaling error",
			owner:          "user",
		},
		{
			name: "#4 Invalid URL format",
			requestBody: models.RequestModifyGet{
				Body: "not-a-valid-url",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Url parsing error or empty schema or empty host",
			owner:          "user",
		},
		{
			name: "#5 URL without scheme",
			requestBody: models.RequestModifyGet{
				Body: "localhost:8080/some-key",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Url parsing error or empty schema or empty host",
			owner:          "user",
		},
		{
			name: "#6 URL without host",
			requestBody: models.RequestModifyGet{
				Body: "http:///some-key",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Url parsing error or empty schema or empty host",
			owner:          "user",
		},
		{
			name: "#7 Empty key in URL",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
			owner:          "user",
		},
		{
			name: "#8 Key not found in repository",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/non-existing-key",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "The link was not found in the repository.",
			owner:          "user",
		},
		{
			name: "#9 Valid URL but repository returns error",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/error-key",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "The link was not found in the repository.",
			owner:          "user",
		},
		{
			name: "#10 Empty JSON structure",
			requestBody: models.RequestModifyGet{
				Body: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Url parsing error or empty schema or empty host",
			owner:          "user",
		},
		{
			name: "#11 Large valid URL",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru/long/path/to/resource/1234567890"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://practicum.yandex.ru/long/path/to/resource/1234567890",
			owner:          "user",
		},
		{
			name: "#12 Valid URL",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://wergwergwergwergwergwerg.ru"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://wergwergwergwergwergwerg.ru",
			owner:          "user",
		},
		{
			name: "#13 Malformed JSON with extra fields",
			requestBody: `{
				"Body": "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru"),
				"Extra": "extra data"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "An anmarshaling error",
			owner:          "user",
		},
		{
			name: "#14 Valid URL but with query parameters",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru?param=value"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://practicum.yandex.ru?param=value",
			owner:          "user",
		},
		{
			name: "#15 Valid URL with fragment",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru#section"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://practicum.yandex.ru#section",
			owner:          "user",
		},
		{
			name: "#16 URL with spaces encoded as %20",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://example.com/some%20path"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://example.com/some%20path",
			owner:          "user",
		},
		{
			name: "#17 URL with special characters",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://example.com/!@#$%^*()"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://example.com/!@#$%^*()",
			owner:          "user",
		},
		{
			name: "#18 URL with Unicode characters",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/" + GenerateShortURL("https://example.com/路径"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://example.com/路径",
			owner:          "user",
		},
		{
			name: "#19 Valid URL but repository returns not found",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/nonexistent-key",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "The link was not found in the repository.",
			owner:          "user",
		},
		{
			name: "#20 Invalid URL ",
			requestBody: models.RequestModifyGet{
				Body: "HTTP://localhost:8080/" + GenerateShortURL("https://example.com"),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
			owner:          "user",
		},
		{
			name: "#21 Valid URL but with localhost as IP",
			requestBody: models.RequestModifyGet{
				Body: "http://127.0.0.1:8080/" + GenerateShortURL("https://example.com"),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "https://example.com",
			owner:          "user",
		},
		{
			name: "#22 Valid URL but empty path",
			requestBody: models.RequestModifyGet{
				Body: "http://localhost:8080/",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Url parsing error or empty schema or empty host",
			owner:          "user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			if tc.expectedStatus == http.StatusOK {
				mockRepo.Save(GenerateShortURL(tc.expectedBody), tc.expectedBody, tc.owner)
			}
			h := NewHandler(mockRepo)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var requestBody []byte
			var err error
			switch body := tc.requestBody.(type) {
			case models.RequestModifyGet:
				requestBody, err = json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
			case string:
				requestBody = []byte(body)
			}

			req, err := http.NewRequest("GET", "/api/shortenGet", bytes.NewBuffer(requestBody))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			h.ModifFget(ctx, rr, req)
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			if tc.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tc.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedBody)
			}
		})
	}
}
