package spiders

import "time"

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
}

type Stats struct {
	Requests                    int
	Items                       int
	Skipped                     int
	Failed                      int
	OffsiteRequests             int
	ConcurrentRequests          int
	ConcurrentRequestsPerDomain int
	DownloadDelay               time.Duration
	Sessions                    map[string]int
}
