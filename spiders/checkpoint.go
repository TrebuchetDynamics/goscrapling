package spiders

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

const checkpointFileName = "checkpoint.json"

type SchedulerSnapshot struct {
	Requests []Request `json:"requests"`
	Seen     []string  `json:"seen"`
}

type CheckpointManager struct {
	dir  string
	path string
}

func NewCheckpointManager(dir string) *CheckpointManager {
	return &CheckpointManager{dir: dir, path: filepath.Join(dir, checkpointFileName)}
}

func (m *CheckpointManager) Save(snapshot SchedulerSnapshot) error {
	if m == nil || m.dir == "" {
		return nil
	}
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, append(body, '\n'), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, m.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (m *CheckpointManager) Load() (SchedulerSnapshot, bool, error) {
	if m == nil || m.path == "" {
		return SchedulerSnapshot{}, false, nil
	}
	body, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SchedulerSnapshot{}, false, nil
		}
		return SchedulerSnapshot{}, false, err
	}
	var snapshot SchedulerSnapshot
	if err := json.Unmarshal(body, &snapshot); err != nil {
		return SchedulerSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func (m *CheckpointManager) Cleanup() error {
	if m == nil || m.path == "" {
		return nil
	}
	if err := os.Remove(m.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func sortedSeenFingerprints(seen map[string]struct{}) []string {
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
