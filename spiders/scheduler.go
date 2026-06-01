package spiders

import (
	"container/heap"
	"sort"
)

type SchedulerOptions struct {
	IncludeHeaders bool
	KeepFragments  bool
}

type Scheduler struct {
	queue   priorityQueue
	seen    map[string]struct{}
	counter int
	options SchedulerOptions
}

func NewScheduler(opts SchedulerOptions) *Scheduler {
	s := &Scheduler{
		seen:    make(map[string]struct{}),
		options: opts,
	}
	heap.Init(&s.queue)
	return s
}

func (s *Scheduler) Enqueue(request Request) (bool, error) {
	fp, err := request.Fingerprint(FingerprintOptions{
		IncludeHeaders: s.options.IncludeHeaders,
		KeepFragments:  s.options.KeepFragments,
	})
	if err != nil {
		return false, err
	}
	if !s.acceptFingerprint(fp, request.DontFilter) {
		return false, nil
	}

	s.counter++
	heap.Push(&s.queue, scheduledRequest{
		request: request.clone(),
		order:   s.counter,
	})
	return true, nil
}

func (s *Scheduler) acceptFingerprint(fp string, dontFilter bool) bool {
	if dontFilter {
		return true
	}
	if _, ok := s.seen[fp]; ok {
		return false
	}
	s.seen[fp] = struct{}{}
	return true
}

func (s *Scheduler) Dequeue() (Request, bool) {
	if s == nil || s.queue.Len() == 0 {
		return Request{}, false
	}
	item := heap.Pop(&s.queue).(scheduledRequest)
	return item.request.clone(), true
}

func (s *Scheduler) Len() int {
	if s == nil {
		return 0
	}
	return s.queue.Len()
}

func (s *Scheduler) Snapshot() SchedulerSnapshot {
	if s == nil {
		return SchedulerSnapshot{}
	}
	items := append([]scheduledRequest(nil), s.queue...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].request.Priority != items[j].request.Priority {
			return items[i].request.Priority > items[j].request.Priority
		}
		return items[i].order < items[j].order
	})
	requests := make([]Request, 0, len(items))
	for _, item := range items {
		requests = append(requests, item.request.clone())
	}
	return SchedulerSnapshot{Requests: requests, Seen: sortedSeenFingerprints(s.seen)}
}

func (s *Scheduler) Restore(snapshot SchedulerSnapshot) {
	if s == nil {
		return
	}
	s.queue = nil
	s.seen = restoredSeenFingerprints(snapshot, s.options)
	s.counter = 0
	heap.Init(&s.queue)
	for _, request := range snapshot.Requests {
		s.counter++
		heap.Push(&s.queue, scheduledRequest{request: request.clone(), order: s.counter})
	}
}

func restoredSeenFingerprints(snapshot SchedulerSnapshot, opts SchedulerOptions) map[string]struct{} {
	seen := make(map[string]struct{}, len(snapshot.Seen)+len(snapshot.Requests))
	for _, fp := range snapshot.Seen {
		seen[fp] = struct{}{}
	}
	for _, request := range snapshot.Requests {
		if request.DontFilter {
			continue
		}
		fp, err := request.Fingerprint(FingerprintOptions{
			IncludeHeaders: opts.IncludeHeaders,
			KeepFragments:  opts.KeepFragments,
		})
		if err != nil {
			continue
		}
		seen[fp] = struct{}{}
	}
	return seen
}

type scheduledRequest struct {
	request Request
	order   int
}

type priorityQueue []scheduledRequest

func (q priorityQueue) Len() int { return len(q) }

func (q priorityQueue) Less(i, j int) bool {
	if q[i].request.Priority != q[j].request.Priority {
		return q[i].request.Priority > q[j].request.Priority
	}
	return q[i].order < q[j].order
}

func (q priorityQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *priorityQueue) Push(x any) {
	*q = append(*q, x.(scheduledRequest))
}

func (q *priorityQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}
