package spiders

import "container/heap"

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
	if !request.DontFilter {
		if _, ok := s.seen[fp]; ok {
			return false, nil
		}
	}
	s.seen[fp] = struct{}{}

	s.counter++
	heap.Push(&s.queue, scheduledRequest{
		request: request.clone(),
		order:   s.counter,
	})
	return true, nil
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
