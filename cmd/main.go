package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	middleware "github.com/Dnlbb/link-shortener/internal/Middlewares"
	"github.com/Dnlbb/link-shortener/internal/config"
	"github.com/Dnlbb/link-shortener/internal/controller"
	controllermod "github.com/Dnlbb/link-shortener/internal/controllerMod"
	"github.com/Dnlbb/link-shortener/internal/handlers"
	"github.com/Dnlbb/link-shortener/internal/logger"
	"github.com/Dnlbb/link-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

type App struct {
	DB *sql.DB
}

func main() {
	ctx := context.Background()

	config.ParseFlags()

	var repo storage.Repository
	var db *sql.DB
	var err error

	if config.Conf.DB != "" {
		db, err = sql.Open("pgx", config.Conf.DB)
		if err != nil {
			log.Fatal("Error with database connection:", err)
		}
		defer db.Close()

		repo = storage.NewPostgresStorage(db)
		err = repo.CreateTable()
		if err != nil {
			log.Fatal("Error creating table:", err)
		}
	} else {
		repo = storage.NewInMemoryStorage()
	}

	handler := handlers.NewHandler(repo)

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	WrappedLogger := logger.NewLogrusLogger(log)
	controller := controller.NewBaseController(ctx, WrappedLogger, *handler)
	modController := controllermod.NewModController(ctx, WrappedLogger, *handler)

	r := chi.NewRouter()
	r.Use(middleware.MiddlewareAuth)
	r.Use(middleware.GzipMiddleware)
	r.Mount("/", controller.Route())
	r.Mount("/api/", modController.Route())
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			err := db.Ping()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handler.GetUserURLs(context.Background(), w, r)
	})
	r.Delete("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		handler.DelUserUrls(context.Background(), w, r)
	})

	log.Info(fmt.Sprintf("Server start on port: %s", config.Conf.Start))
	err = http.ListenAndServe(config.Conf.Start, r)
	if err != nil {
		log.Fatal("Error when starting the server:", err)
	}
}
