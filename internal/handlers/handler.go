package handlers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"path/filepath"

	middlewares "github.com/Dnlbb/link-shortener/internal/Middlewares"
	"github.com/Dnlbb/link-shortener/internal/config"
	"github.com/Dnlbb/link-shortener/internal/models"
	"github.com/Dnlbb/link-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	repo storage.Repository
}

func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

func saveToFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() > 0 {
		if _, err = file.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) Fpost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():

		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:
		userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error reading the request body"))
			return
		}

		originalURL := string(body)

		parsedURL, err := url.ParseRequestURI(originalURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Url parsing error or empty schema or empty host"))
			return
		}

		shortURL := GenerateShortURL(originalURL)

		_, exists := h.repo.Find(shortURL)
		if exists {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("http://localhost:8080/" + shortURL))
			return
		}

		err = h.repo.Save(shortURL, originalURL, userID)
		if err != nil {
			log.Printf("Error saving to repository: %v", err)
			http.Error(w, "Error saving the link to the repository", http.StatusInternalServerError)
			return
		}
		path := config.Conf.Result
		if path == "" {
			path = "http://localhost:8080"
		}
		if config.Conf.File != "" {
			ToFile := struct {
				UUID        int    `json:"uuid"`
				ShortURL    string `json:"short_url"`
				OriginalURL string `json:"original_url"`
			}{
				UUID:        h.repo.GetUUID(),
				ShortURL:    "http://localhost:8080/" + shortURL,
				OriginalURL: originalURL,
			}

			if jsonToFile, err := json.Marshal(ToFile); err == nil {
				saveToFile(config.Conf.File, jsonToFile)
			}
		}
		response := fmt.Sprintf("%s/%s", path, shortURL)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(response))
	}
}

func GenerateShortURL(url string) string {
	hash := sha1.New()
	hash.Write([]byte(url))
	return hex.EncodeToString(hash.Sum(nil))[:8]
}

func (h *Handler) Fget(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:

		shortURL := chi.URLParam(r, "shortURL")
		originalURL, exists := h.repo.Find(shortURL)
		if originalURL == "deleted" {
			w.WriteHeader(http.StatusGone)
			return
		}
		if shortURL == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !exists {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("The link was not found in the repository."))
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) ModifPost(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:
		var req models.RequestModifyPost
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Write([]byte("Error reading the request body"))
			return
		}

		if len(buf.Bytes()) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: empty request body"))
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("An anmarshaling error"))
			return
		}

		const maxURLLength = 2048
		if len(req.Body) > maxURLLength {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: the request body is too long"))
			return
		}

		parsedURL, err := url.ParseRequestURI(req.Body)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Url parsing error or empty schema or empty host"))
			return
		}
		userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}
		shortURL := GenerateShortURL(req.Body)

		_, exists := h.repo.Find(shortURL)
		if exists {
			respStruct := models.ResponseModifyPost{
				Body: "http://localhost:8080/" + shortURL,
			}
			resp, err := json.Marshal(respStruct)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(resp)
			return
		}
		err = h.repo.Save(shortURL, req.Body, userID)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("error saving the link to the repository"))
			return
		}
		if config.Conf.File != "" {
			ToFile := struct {
				UUID        int    `json:"uuid"`
				ShortURL    string `json:"short_url"`
				OriginalURL string `json:"original_url"`
			}{
				UUID:        h.repo.GetUUID(),
				ShortURL:    "http://localhost:8080/" + shortURL,
				OriginalURL: req.Body,
			}

			if jsonToFile, err := json.Marshal(ToFile); err == nil {
				saveToFile(config.Conf.File, jsonToFile)
			}
		}
		respStruct := models.ResponseModifyPost{
			Body: "http://localhost:8080/" + shortURL,
		}
		resp, err := json.Marshal(respStruct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
	}
}

func (h *Handler) ModifFget(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:
		var req models.RequestModifyGet
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			w.Write([]byte("Error reading the request body"))
			return
		}

		if len(buf.Bytes()) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error: empty request body"))
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("An anmarshaling error"))
			return
		}

		parsedURL, err := url.ParseRequestURI(req.Body)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Url parsing error or empty schema or empty host"))
			return
		}

		baseURL := "http://localhost:8080/"
		if strings.HasPrefix(req.Body, "http://127.0.0.1:8080/") {
			baseURL = "http://127.0.0.1:8080/"
		}

		key := strings.Replace(req.Body, baseURL, "", 1)
		originalURL, exists := h.repo.Find(key)
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Url parsing error or empty schema or empty host"))
			return
		}

		if !exists {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("The link was not found in the repository."))
			return
		}

		respStruct := models.ResponseModifyGet{
			Body: originalURL,
		}
		resp, err := json.Marshal(respStruct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func (h *Handler) Batch(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:
		userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}
		var reqBatch models.ReqBatch
		err := json.NewDecoder(r.Body).Decode(&reqBatch)
		var FlagToFile bool

		if config.Conf.File != "" {
			FlagToFile = true
		} else {
			FlagToFile = false
		}

		if err != nil {
			http.Error(w, "Error reading or unmarshaling the request body", http.StatusBadRequest)
			return
		}

		if len(reqBatch) == 0 {
			http.Error(w, "Error: empty request body", http.StatusBadRequest)
			return
		}

		db := h.repo.(*storage.PostgresStorage).GetDB()
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Error starting the transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var resp models.RespBatch
		for _, req := range reqBatch {
			shortURL := GenerateShortURL(req.OriginalURL)
			err = h.repo.Save(shortURL, req.OriginalURL, userID)
			if err != nil {
				http.Error(w, "Error saving the link to the repository.", http.StatusInternalServerError)
				return
			}

			if FlagToFile {
				ToFile := struct {
					UUID        int    `json:"uuid"`
					ShortURL    string `json:"short_url"`
					OriginalURL string `json:"original_url"`
				}{
					UUID:        h.repo.GetUUID(),
					ShortURL:    "http://localhost:8080/" + shortURL,
					OriginalURL: req.OriginalURL,
				}

				if jsonToFile, err := json.Marshal(ToFile); err == nil {
					saveToFile(config.Conf.File, jsonToFile)
				}
			}

			BatchResp := models.MiniBatchResp{
				ID:       req.ID,
				ShortURL: "http://localhost:8080/" + shortURL,
			}
			resp = append(resp, BatchResp)
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Error committing the transaction", http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Error marshaling the response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(response)
	}
}

func (h *Handler) GetUserURLs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:

		userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		urls, err := h.repo.(*storage.PostgresStorage).FindAllByOwner(userID)
		if err != nil {
			http.Error(w, "Error retrieving URLs from storage", http.StatusInternalServerError)
			return
		}
		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resp, err := json.Marshal(urls)
		if err != nil {
			http.Error(w, "Error with marshal", http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func (h *Handler) DelUserUrls(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Request cancelled by the client", http.StatusRequestTimeout)
		}
		return
	default:
		var wg sync.WaitGroup
		var req models.UserDelUrls
		userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Error reading or unmarshaling the request body", http.StatusBadRequest)
			return
		}

		if len(req) == 0 {
			http.Error(w, "Error: empty request body", http.StatusBadRequest)
			return
		}
		for _, url := range req {
			wg.Add(1)
			go h.repo.(*storage.PostgresStorage).DeleterURL(url, userID, &wg)
		}
		wg.Wait()
		w.WriteHeader(http.StatusAccepted)
	}

}
