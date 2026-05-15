package goscrapling

import (
	"context"
	"sync"
)

type Key struct {
	Domain     string
	Identifier string
}

type Store interface {
	Save(ctx context.Context, key Key, fp Fingerprint) error
	Load(ctx context.Context, key Key) (Fingerprint, bool, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	values map[Key]Fingerprint
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{values: make(map[Key]Fingerprint)}
}

func (s *MemoryStore) Save(ctx context.Context, key Key, fp Fingerprint) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = cloneFingerprint(fp)
	return nil
}

func (s *MemoryStore) Load(ctx context.Context, key Key) (Fingerprint, bool, error) {
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

func cloneFingerprint(fp Fingerprint) Fingerprint {
	return Fingerprint{
		Tag:              fp.Tag,
		Text:             fp.Text,
		Attributes:       cloneStringMap(fp.Attributes),
		ParentTag:        fp.ParentTag,
		ParentText:       fp.ParentText,
		ParentAttributes: cloneStringMap(fp.ParentAttributes),
		SiblingTags:      append([]string(nil), fp.SiblingTags...),
		ChildrenTags:     append([]string(nil), fp.ChildrenTags...),
		PathTags:         append([]string(nil), fp.PathTags...),
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
