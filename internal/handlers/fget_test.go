package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func longString(length int) string {
	str := make([]byte, length)
	for i := 0; i < length; i++ {
		str[i] = 'a'
	}
	return string(str)
}

func TestFget(t *testing.T) {
	type Want struct {
		statusCode int
		location   string
	}

	type Request struct {
		path string
	}

	testCases := []struct {
		name        string
		want        Want
		request     Request
		originalURL string
		owner       string
	}{
		{
			name: "#1 Valid short URL",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://practicum.yandex.ru/"),
			},
			originalURL: "https://practicum.yandex.ru/",
			owner:       "user1",
		},
		{
			name: "#2 Valid short URL with different path",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://google.com/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://google.com/"),
			},
			originalURL: "https://google.com/",
			owner:       "user1",
		},
		{
			name: "#3 Valid short URL with query parameters",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com?query=1",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com?query=1"),
			},
			originalURL: "https://example.com?query=1",
			owner:       "user1",
		},
		{
			name: "#4 Valid short URL with fragment",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com#section",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com#section"),
			},
			originalURL: "https://example.com#section",
			owner:       "user1",
		},
		{
			name: "#5 Valid short URL with complex URL",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/path/to/page?query=1#section",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/path/to/page?query=1#section"),
			},
			originalURL: "https://example.com/path/to/page?query=1#section",
			owner:       "user1",
		},
		{
			name: "#6 Invalid short URL (not found)",
			want: Want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
			request: Request{
				path: "/invalidURL",
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#7 Missing short URL in request",
			want: Want{
				statusCode: http.StatusNotFound,
				location:   "",
			},
			request: Request{
				path: "/",
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#8 Invalid URL format",
			want: Want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
			request: Request{
				path: "/123@!#",
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#9 Short URL with no original URL associated",
			want: Want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://nonexistent.com/"),
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#10 Valid short URL",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://safasfsdgsdsdf.com",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://safasfsdgsdsdf.com"),
			},
			originalURL: "https://safasfsdgsdsdf.com",
			owner:       "user1",
		},
		{
			name: "#11 Long URL path",
			want: Want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/"+longString(1000)),
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#12 Valid short URL",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://edfsbsdfbd/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://edfsbsdfbd/"),
			},
			originalURL: "https://edfsbsdfbd/",
			owner:       "user1",
		},
		{
			name: "#13 Valid short URL with subdomain",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://subdomain.example.com/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://subdomain.example.com/"),
			},
			originalURL: "https://subdomain.example.com/",
			owner:       "user1",
		},
		{
			name: "#14 Valid short URL with IP address",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://192.168.1.1/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://192.168.1.1/"),
			},
			originalURL: "https://192.168.1.1/",
			owner:       "user1",
		},
		{
			name: "#15 Valid short URL with port number",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com:8080/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com:8080/"),
			},
			originalURL: "https://example.com:8080/",
			owner:       "user1",
		},
		{
			name: "#16 Valid short URL with FTP",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "ftp://example.com/",
			},
			request: Request{
				path: "/" + GenerateShortURL("ftp://example.com/"),
			},
			originalURL: "ftp://example.com/",
			owner:       "user1",
		},
		{
			name: "#17 Valid short URL with file scheme",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "file:///C:/path/to/file",
			},
			request: Request{
				path: "/" + GenerateShortURL("file:///C:/path/to/file"),
			},
			originalURL: "file:///C:/path/to/file",
			owner:       "user1",
		},
		{
			name: "#18 Invalid short URL with no alphanumeric characters",
			want: Want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
			request: Request{
				path: "/------",
			},
			originalURL: "",
			owner:       "user1",
		},
		{
			name: "#19 Valid short URL with underscore",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/_page",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/_page"),
			},
			originalURL: "https://example.com/_page",
			owner:       "user1",
		},
		{
			name: "#20 Valid short URL with tilde",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/~page",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/~page"),
			},
			originalURL: "https://example.com/~page",
			owner:       "user1",
		},
		{
			name: "#21 Valid short URL with UTF-8 characters",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://пример.рф/",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://пример.рф/"),
			},
			originalURL: "https://пример.рф/",
			owner:       "user1",
		},
		{
			name: "#22 Valid short URL with emoji",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/😊",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/😊"),
			},
			originalURL: "https://example.com/😊",
			owner:       "user1",
		},
		{
			name: "#23 Valid short URL with escaped characters",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/%20page",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/%20page"),
			},
			originalURL: "https://example.com/%20page",
			owner:       "user1",
		},
		{
			name: "#24 Valid short URL with percent sign",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/100%",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/100%"),
			},
			originalURL: "https://example.com/100%",
			owner:       "user1",
		},
		{
			name: "#25 Valid short URL with parenthesis",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/(page)",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/(page)"),
			},
			originalURL: "https://example.com/(page)",
			owner:       "user1",
		},
		{
			name: "#26 Valid short URL with ampersand",
			want: Want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com/page&section",
			},
			request: Request{
				path: "/" + GenerateShortURL("https://example.com/page&section"),
			},
			originalURL: "https://example.com/page&section",
			owner:       "user1",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			handler := NewHandler(mockRepo)
			shortURL := GenerateShortURL(test.originalURL)

			if test.originalURL != "" {
				mockRepo.Save(shortURL, test.originalURL, test.owner)
			}

			r := chi.NewRouter()
			r.Get("/{shortURL}", handler.FgetAdapter())

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			request := httptest.NewRequest(http.MethodGet, test.request.path, nil).WithContext(ctx)
			w := httptest.NewRecorder()

			request.Header.Set("X-Owner-ID", test.owner)

			r.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, test.want.statusCode, result.StatusCode)

			location := result.Header.Get("Location")
			assert.Equal(t, test.want.location, location)
			err := result.Body.Close()
			require.NoError(t, err)
		})
	}

}

func (h *Handler) FgetAdapter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Fget(r.Context(), w, r)
	}
}
