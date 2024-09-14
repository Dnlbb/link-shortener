package storage

import (
	"database/sql"
	"log"
	"sync"

	"github.com/Dnlbb/link-shortener/internal/models"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) GetDB() *sql.DB {
	return s.db
}

func (s *PostgresStorage) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(8) NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		owner VARCHAR(50) NOT NULL,
		DeletedFlag BOOL NOT NULL
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	indexQuery := `CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls (original_url);`
	_, err = s.db.Exec(indexQuery)
	return err
}
func (s *PostgresStorage) Save(shortURL, originalURL, owner string) error {
	log.Printf("Saving URL: shortURL=%s, originalURL=%s, owner=%s", shortURL, originalURL, owner)
	query := `
	INSERT INTO urls (short_url, original_url, owner, DeletedFlag)
	VALUES ($1, $2, $3, false)
	ON CONFLICT (short_url) DO NOTHING`
	_, err := s.db.Exec(query, shortURL, originalURL, owner)
	return err
}

func (s *PostgresStorage) Find(shortURL string) (string, bool) {
	var originalURL string
	var deleted bool
	query := `SELECT original_url, DeletedFlag FROM urls WHERE short_url = $1`
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL, &deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		return "", false
	}
	if deleted {
		return "deleted", true
	}
	return originalURL, true
}

func (s *PostgresStorage) GetUUID() int {
	var uuid int
	query := `SELECT COUNT(*) FROM urls`
	err := s.db.QueryRow(query).Scan(&uuid)
	if err != nil {
		return 0
	}
	return uuid
}
func (s *PostgresStorage) FindAllByOwner(owner string) ([]models.ResponseToOwner, error) {
	query := `SELECT short_url, original_url FROM urls WHERE owner = $1` // Исправьте запрос для выбора всех необходимых столбцов
	rows, err := s.db.Query(query, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resp []models.ResponseToOwner
	for rows.Next() {
		var shortURL string
		var originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}
		resp = append(resp, models.ResponseToOwner{
			ShortURL:    "http://localhost:8080/" + shortURL,
			OriginalURL: originalURL,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *PostgresStorage) DeleterURL(url string, user string, wg *sync.WaitGroup) error {
	defer wg.Done()
	query := `UPDATE urls SET DeletedFlag = true WHERE short_url = $1 AND owner = $2`
	_, err := s.db.Exec(query, url, user)
	if err != nil {
		return err
	}
	return nil
}
