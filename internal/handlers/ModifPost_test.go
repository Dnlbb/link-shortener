package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware "github.com/Dnlbb/link-shortener/internal/Middlewares"
	"github.com/Dnlbb/link-shortener/internal/models"
)

func TestModifPost(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "#1 valid request",
			requestBody: models.RequestModifyPost{
				Body: "https://practicum.yandex.ru",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://practicum.yandex.ru"),
		},
		{
			name:           "#2 empty body",
			requestBody:    models.RequestModifyPost{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "#3 invalid JSON",
			requestBody:    `{"url": "https://practicum.yandex.ru",`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "#4 too long URL",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com/" + string(make([]byte, 2048)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "#5 URL with special characters",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com/path?query=%20",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://example.com/path?query=%20"),
		},
		{
			name: "#6 URL with Russian characters",
			requestBody: models.RequestModifyPost{
				Body: "https://яндекс.рф",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://яндекс.рф"),
		},
		{
			name: "#7 long query parameters in URL",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com/path?query=" + string(make([]byte, 512)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "#8 URL with path parameters",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com/path/param",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://example.com/path/param"),
		},
		{
			name: "#9 URL with fragment",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com/path#fragment",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://example.com/path#fragment"),
		},
		{
			name: "#10 valid request",
			requestBody: models.RequestModifyPost{
				Body: "https://wefwvfwefwef.ru",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://wefwvfwefwef.ru"),
		},
		{
			name: "#11 attempt to save a non-existent domain",
			requestBody: models.RequestModifyPost{
				Body: "https://nonexistent.domain",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://nonexistent.domain"),
		},
		{
			name: "#12 URL without scheme",
			requestBody: models.RequestModifyPost{
				Body: "example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name: "#13 URL with port number",
			requestBody: models.RequestModifyPost{
				Body: "https://example.com:8080/path",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://example.com:8080/path"),
		},
		{
			name: "#14 valid request",
			requestBody: models.RequestModifyPost{
				Body: "https://qwegwegerwqgewrgwergasdgdsgasdg.ru",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("https://qwegwegerwqgewrgwergasdgdsgasdg.ru"),
		},
		{
			name: "#15 HTTP URL",
			requestBody: models.RequestModifyPost{
				Body: "http://secure-site.com",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("http://secure-site.com"),
		},
		{
			name: "#16 URL with IP address",
			requestBody: models.RequestModifyPost{
				Body: "http://192.168.0.1",
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/" + GenerateShortURL("http://192.168.0.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			handler := middleware.MiddlewareAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h := NewHandler(mockRepo)
				ctx := r.Context()

				h.ModifPost(ctx, w, r)
			}))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var requestBody []byte
			var err error
			switch body := tt.requestBody.(type) {
			case models.RequestModifyPost:
				requestBody, err = json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
			case string:
				requestBody = []byte(body)
			}

			req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(requestBody))
			if err != nil {
				t.Fatal(err)
			}
			userID := "mockUserID"
			signature := middleware.SignData(userID)
			cookieValue := userID + "|" + signature
			req.AddCookie(&http.Cookie{Name: "session_id", Value: cookieValue})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req.WithContext(ctx))

			if status := w.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedBody != "" && !bytes.Contains(w.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					w.Body.String(), tt.expectedBody)
			}
		})
	}
}
