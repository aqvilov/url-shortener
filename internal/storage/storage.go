package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

// указатель на логгер, чтобы в общих логах отображать бдшные
type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

func New(log *slog.Logger) (*Storage, error) {
	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		return nil, errors.New("env DB_CONN isnt set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	//создание таблицы
	if err := createTable(db, log); err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	log.Info("database connected")

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func createTable(db *sql.DB, log *slog.Logger) error {
	query := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			alias VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			clicks INTEGER DEFAULT 0
		);
		
		CREATE INDEX IF NOT EXISTS idx_alias ON urls(alias);
	`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Info("table created successfully")
	return nil
}

// saveurl
// тут просто запрос и все
func (s *Storage) SaveUrl(alias string, originalUrl string) error {
	query := `INSERT INTO urls (alias, original_url) VALUES ($1, $2)`

	_, err := s.db.Exec(query, alias, originalUrl)
	if err != nil {
		return fmt.Errorf("failed to save url in db: %w", err)
	}

	s.log.Info("url saved successfully", "alias", alias)
	return nil
}

func (s *Storage) GetUrl(alias string) (string, error) {
	query := `
		UPDATE urls 
		SET clicks = clicks + 1 
		WHERE alias = $1 
		RETURNING original_url
	`
	var originalUrl string

	// так как вернется одна строка используем row и сканим сразу же
	err := s.db.QueryRow(query, alias).Scan(&originalUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("alias not found: %w", err)
		}
		return "", fmt.Errorf("failed to get and update url: %w", err)
	}
	// переход по ссылкам происиходит постоянно, логируем на уровень ниже (в дебаг)
	s.log.Debug("url retrieved successfully", "alias", alias)

	return originalUrl, nil
}

func (s *Storage) DeleteUrl(alias string) error {
	updateQuery := `DELETE FROM urls WHERE alias = $1`
	_, err := s.db.Exec(updateQuery, alias)
	if err != nil {
		return fmt.Errorf("failed to delete url %w", err)
	}

	s.log.Info("url deleted successfully", "alias", alias)

	return nil
}

type Info struct {
	originalUrl string
	alias       string
	clicks      int
}

func (s *Storage) GetInfo() ([]Info, error) {
	query := `SELECT alias, original_url, clicks FROM urls ORDER BY clicks DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("fail: %w", err)
	}
	defer rows.Close()

	var infos []Info

	for rows.Next() {
		var info Info
		err = rows.Scan(&info.alias, &info.originalUrl, &info.clicks)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		infos = append(infos, info)
	}

	s.log.Info("infos retrieved successfully", "infos", len(infos))

	return infos, nil
}
