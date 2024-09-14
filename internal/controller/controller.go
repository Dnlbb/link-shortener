package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/Dnlbb/link-shortener/internal/handlers"
	"github.com/Dnlbb/link-shortener/internal/logger"
	"github.com/go-chi/chi/v5"
)

type BaseController struct {
	logger  logger.Logger
	storage handlers.Handler
	ctx     context.Context
}

func NewBaseController(ctx context.Context, logger logger.Logger, handler handlers.Handler) *BaseController {
	return &BaseController{
		logger:  logger,
		storage: handler,
		ctx:     ctx,
	}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", c.WithLogging(c.storage.Fpost))
	r.Get("/{shortURL}", c.WithLogging(c.storage.Fget))
	return r
}

func (c *BaseController) WithLogging(h func(ctx context.Context, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &logger.ResponseData{}

		lw := logger.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h(r.Context(), &lw, r)

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
