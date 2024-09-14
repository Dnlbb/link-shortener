package storage

type Repository interface {
	Save(shortURL, originalURL, owner string) error
	Find(shortURL string) (string, bool)
	GetUUID() int
	CreateTable() error
}
