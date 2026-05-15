package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const fileStoreSchemaVersion = 1

var ErrUnsupportedStoreSchema = errors.New("unsupported adaptive store schema")

type FileStore struct {
	mu     sync.RWMutex
	path   string
	values map[Key]Fingerprint
}

type fileStoreDocument struct {
	SchemaVersion int               `json:"schema_version"`
	Records       []fileStoreRecord `json:"records"`
}

type fileStoreRecord struct {
	Domain      string      `json:"domain"`
	Identifier  string      `json:"identifier"`
	Fingerprint Fingerprint `json:"fingerprint"`
}

func NewFileStore(path string) (*FileStore, error) {
	store := &FileStore{
		path:   path,
		values: make(map[Key]Fingerprint),
	}

	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FileStore) Save(ctx context.Context, key Key, fp Fingerprint) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = cloneFingerprint(fp)
	return s.writeLocked()
}

func (s *FileStore) Load(ctx context.Context, key Key) (Fingerprint, bool, error) {
	if err := ctx.Err(); err != nil {
		return Fingerprint{}, false, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	fp, ok := s.values[key]
	if !ok {
		return Fingerprint{}, false, nil
	}
	return cloneFingerprint(fp), true, nil
}

func (s *FileStore) load() error {
	body, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	var document fileStoreDocument
	if err := json.Unmarshal(body, &document); err != nil {
		return err
	}
	if document.SchemaVersion != fileStoreSchemaVersion {
		return fmt.Errorf("%w: got %d want %d", ErrUnsupportedStoreSchema, document.SchemaVersion, fileStoreSchemaVersion)
	}

	for _, record := range document.Records {
		key := Key{Domain: record.Domain, Identifier: record.Identifier}
		s.values[key] = cloneFingerprint(record.Fingerprint)
	}
	return nil
}

func (s *FileStore) writeLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	document := fileStoreDocument{
		SchemaVersion: fileStoreSchemaVersion,
		Records:       make([]fileStoreRecord, 0, len(s.values)),
	}
	for key, fp := range s.values {
		document.Records = append(document.Records, fileStoreRecord{
			Domain:      key.Domain,
			Identifier:  key.Identifier,
			Fingerprint: cloneFingerprint(fp),
		})
	}
	sort.Slice(document.Records, func(i, j int) bool {
		left := document.Records[i]
		right := document.Records[j]
		if left.Domain != right.Domain {
			return left.Domain < right.Domain
		}
		return left.Identifier < right.Identifier
	})

	body, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')

	tempFile, err := os.CreateTemp(filepath.Dir(s.path), ".adaptive-store-*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(body); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}
