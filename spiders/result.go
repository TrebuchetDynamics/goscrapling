package spiders

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Output struct {
	Item    map[string]any
	Request *Request
}

func Item(item map[string]any) Output {
	return Output{Item: cloneMeta(item)}
}

func Next(request Request) Output {
	copied := request.clone()
	return Output{Request: &copied}
}

type Result struct {
	Items  []map[string]any
	Errors []error
	Stats  Stats
	Paused bool
}

func (r Result) ItemList() ItemList {
	items := make(ItemList, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, cloneMeta(item))
	}
	return items
}

type ItemList []map[string]any

func (l ItemList) ToJSON(path string, indent bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var body []byte
	var err error
	if indent {
		body, err = json.MarshalIndent([]map[string]any(l), "", "  ")
	} else {
		body, err = json.Marshal([]map[string]any(l))
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func (l ItemList) ToJSONL(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, item := range l {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}
	return nil
}

type Stats struct {
	Requests                    int
	Items                       int
	Skipped                     int
	Failed                      int
	BlockedRequests             int
	BlockedRetries              int
	OffsiteRequests             int
	RobotsDisallowed            int
	ItemsDropped                int
	ConcurrentRequests          int
	ConcurrentRequestsPerDomain int
	DownloadDelay               time.Duration
	Sessions                    map[string]int
	StatusCodes                 map[int]int
	ResponseBytes               int
	DomainResponseBytes         map[string]int
	StartTime                   time.Time
	EndTime                     time.Time
	Elapsed                     time.Duration
}
