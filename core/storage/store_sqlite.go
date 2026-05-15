package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

const sqliteStoreSchemaVersion = 1

var ErrClosedStore = errors.New("adaptive store closed")

type SQLiteStore struct {
	mu     sync.RWMutex
	db     *sql.DB
	closed bool
}

var _ Store = (*SQLiteStore)(nil)

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &SQLiteStore{db: db}
	if err := store.setup(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Save(ctx context.Context, key Key, fp Fingerprint) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return ErrClosedStore
	}

	body, err := json.Marshal(cloneFingerprint(fp))
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO storage (url, identifier, element_data)
		 VALUES (?, ?, ?)
		 ON CONFLICT(url, identifier) DO UPDATE SET element_data = excluded.element_data`,
		key.Domain,
		key.Identifier,
		string(body),
	)
	return err
}

func (s *SQLiteStore) Load(ctx context.Context, key Key) (Fingerprint, bool, error) {
	if err := ctx.Err(); err != nil {
		return Fingerprint{}, false, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return Fingerprint{}, false, ErrClosedStore
	}

	var body string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT element_data FROM storage WHERE url = ? AND identifier = ?`,
		key.Domain,
		key.Identifier,
	).Scan(&body)
	if errors.Is(err, sql.ErrNoRows) {
		return Fingerprint{}, false, nil
	}
	if err != nil {
		return Fingerprint{}, false, err
	}

	var fp Fingerprint
	if err := json.Unmarshal([]byte(body), &fp); err != nil {
		return Fingerprint{}, false, err
	}
	return cloneFingerprint(fp), true, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.db.Close()
}

func (s *SQLiteStore) setup(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA busy_timeout=5000`); err != nil {
		return err
	}

	version, err := sqliteUserVersion(ctx, s.db)
	if err != nil {
		return err
	}
	hasStorage, err := sqliteTableExists(ctx, s.db, "storage")
	if err != nil {
		return err
	}
	if version == 0 && hasStorage {
		return fmt.Errorf("%w: got %d want %d", ErrUnsupportedStoreSchema, version, sqliteStoreSchemaVersion)
	}
	if version != 0 && version != sqliteStoreSchemaVersion {
		return fmt.Errorf("%w: got %d want %d", ErrUnsupportedStoreSchema, version, sqliteStoreSchemaVersion)
	}

	if _, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS storage (
			id INTEGER PRIMARY KEY,
			url TEXT NOT NULL,
			identifier TEXT NOT NULL,
			element_data TEXT NOT NULL,
			UNIQUE (url, identifier)
		)
	`); err != nil {
		return err
	}

	if version == 0 {
		_, err = s.db.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, sqliteStoreSchemaVersion))
		return err
	}
	return nil
}

func sqliteUserVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	if err := db.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {
		return 0, err
	}
	return version, nil
}

func sqliteTableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var tableName string
	err := db.QueryRowContext(
		ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
		name,
	).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
