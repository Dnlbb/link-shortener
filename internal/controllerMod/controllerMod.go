package controllermod

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Dnlbb/link-shortener/internal/handlers"
	"github.com/Dnlbb/link-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

type ModController struct {
	logger  logger.Logger
	storage handlers.Handler
	ctx     context.Context
}

func NewModController(ctx context.Context, logger logger.Logger, handler handlers.Handler) *ModController {
	return &ModController{
		logger:  logger,
		storage: handler,
		ctx:     ctx,
	}
}

func (c *ModController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/shorten", c.WithLogging(c.storage.ModifPost))
	r.Get("/shortenGet", c.WithLogging(c.storage.ModifFget))
	r.Post("/shorten/batch", c.WithLogging(c.storage.Batch))

	return r
}

func (c *ModController) WithLogging(h func(ctx context.Context, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &logger.ResponseData{}

		lw := logger.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h(r.Context(), &lw, r)

		fmt.Println("Отправленный JSON:", lw.ResponseData)
		duration := time.Since(start)

		c.logger.WithFields(map[string]interface{}{
			"uri":      r.RequestURI,
			"method":   r.Method,
			"status":   responseData.Status,
			"duration": duration,
			"size":     responseData.Size,
		}).Info("Request processed")
	}
}
