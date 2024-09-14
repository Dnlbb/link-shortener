package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	middleware "github.com/Dnlbb/link-shortener/internal/Middlewares"
	"github.com/Dnlbb/link-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFpost(t *testing.T) {
	type Want struct {
		contentType string
		statusCode  int
	}

	type Request struct {
		path string
		body string
	}

	testCases := []struct {
		name    string
		want    Want
		request Request
	}{
		{
			name: "#1 Valid URL",
			want: Want{
				contentType: "text/plain",
				statusCode:  201},
			request: Request{
				path: "/",
				body: "https://practicum.yandex.ru/"},
		},
		{
			name: "#2 Valid URL",
			want: Want{
				contentType: "text/plain",
				statusCode:  201},
			request: Request{
				path: "/",
				body: "https://sdfsdfsdxcxcv.yandex.ru/"},
		},
		{
			name: "#3 Valid URL",
			want: Want{
				contentType: "text/plain",
				statusCode:  201},
			request: Request{
				path: "/",
				body: "https://aaa.yandex.ru/"},
		},
		{
			name: "#4 Invalid URL in Body",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "invalid-url",
			},
		},
		{
			name: "#5 Empty Body",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "",
			},
		},
		{
			name: "#6 Invalid URL in body",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest},
			request: Request{
				path: "/",
				body: "https:///"},
		},
		{
			name: "#7 Valid URL",
			want: Want{
				contentType: "text/plain",
				statusCode:  201},
			request: Request{
				path: "/",
				body: "https://.ru/"},
		},
		{
			name: "#8 Valid URL with HTTP scheme",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "http://example.com/",
			},
		},
		{
			name: "#9 Valid URL with subdomain",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://sub.example.com/",
			},
		},
		{
			name: "#10 Valid URL with port number",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com:8080/",
			},
		},
		{
			name: "#11 Valid URL with query parameters",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com/?param=value",
			},
		},
		{
			name: "#12 Valid URL with fragment",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com/#section",
			},
		},
		{
			name: "#13 Valid URL with complex path and parameters",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com/path/to/resource?param=value#section",
			},
		},
		{
			name: "#14 Invalid URL with missing scheme",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "://example.com",
			},
		},
		{
			name: "#15 Invalid URL with missing host",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "https:///path",
			},
		},
		{
			name: "#16 Valid URL with non-ASCII characters",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://пример.рф/",
			},
		},
		{
			name: "#17 Valid URL with new scheme",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "ftp://example.com/",
			},
		},
		{
			name: "#18 Valid URL with uncommon TLD",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.xyz/",
			},
		},
		{
			name: "#19 Valid URL with IP address",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "http://192.168.1.1/",
			},
		},
		{
			name: "#20 Valid test",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com/",
			},
		},
		{
			name: "#21 Valid URL with multiple subdomains",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://sub.sub.example.com/",
			},
		},
		{
			name: "#22 Invalid URL with empty scheme and host",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: ":/path",
			},
		},
		{
			name: "#23 Invalid URL with spaces",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "https://exa mple.com/",
			},
		},
		{
			name: "#24 Valid URL with HTTPS and port",
			want: Want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
			},
			request: Request{
				path: "/",
				body: "https://example.com:8443/",
			},
		},
		{
			name: "#25 Valid URL with path only",
			want: Want{
				contentType: "",
				statusCode:  http.StatusBadRequest,
			},
			request: Request{
				path: "/",
				body: "/path/to/resource",
			},
		},
	}
	config.Conf.Key = "test-secret-key"

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			handler := middleware.MiddlewareAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h := NewHandler(mockRepo)
				ctx := r.Context()

				h.Fpost(ctx, w, r)
			}))

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.request.body))

			userID := "mockUserID"
			signature := middleware.SignData(userID)
			cookieValue := userID + "|" + signature
			req.AddCookie(&http.Cookie{Name: "session_id", Value: cookieValue})

			w := httptest.NewRecorder()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			handler.ServeHTTP(w, req.WithContext(ctx))

			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)

		})
	}
}
