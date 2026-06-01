package spiders

import (
	"context"
	"errors"
	"time"
)

type crawlRuntime struct {
	crawler            Crawler
	ctx                context.Context
	cancel             context.CancelFunc
	allowedDomains     []string
	concurrentRequests int
	domainLimiters     *crawlerDomainLimiters
	sleep              func(context.Context, time.Duration) error
	scheduler          *Scheduler
	checkpoint         *CheckpointManager
	result             Result
	itemStream         chan<- map[string]any
	active             int
	activeRequests     map[int]Request
	nextTaskID         int
	stopErr            error
	taskResults        chan crawlerTaskResult
}

func newCrawlRuntime(ctx context.Context, crawler Crawler) *crawlRuntime {
	runCtx, cancel := context.WithCancel(ctx)
	concurrentRequests := effectiveConcurrentRequests(crawler.ConcurrentRequests)
	sleep := crawler.sleep
	if sleep == nil {
		sleep = defaultCrawlerSleep
	}
	scheduler := crawler.Scheduler
	if scheduler == nil {
		scheduler = NewScheduler(SchedulerOptions{})
	}
	if crawler.RobotsTxtObey && crawler.RobotsTxtManager == nil {
		crawler.RobotsTxtManager = NewRobotsTxtManager(func(ctx context.Context, robotsURL, sid string) (Response, error) {
			return crawler.Sessions.Fetch(ctx, Request{URL: robotsURL, SID: sid})
		})
	}
	var checkpoint *CheckpointManager
	if crawler.CheckpointDir != "" {
		checkpoint = NewCheckpointManager(crawler.CheckpointDir)
	}

	return &crawlRuntime{
		crawler:            crawler,
		ctx:                runCtx,
		cancel:             cancel,
		allowedDomains:     normalizeAllowedDomains(crawler.AllowedDomains),
		concurrentRequests: concurrentRequests,
		domainLimiters:     newCrawlerDomainLimiters(crawler.ConcurrentRequestsPerDomain),
		sleep:              sleep,
		scheduler:          scheduler,
		checkpoint:         checkpoint,
		result: Result{Stats: Stats{
			ConcurrentRequests:          concurrentRequests,
			ConcurrentRequestsPerDomain: crawler.ConcurrentRequestsPerDomain,
			DownloadDelay:               crawler.DownloadDelay,
			Sessions:                    make(map[string]int),
			StatusCodes:                 make(map[int]int),
			DomainResponseBytes:         make(map[string]int),
		}},
		activeRequests: make(map[int]Request),
		taskResults:    make(chan crawlerTaskResult),
	}
}

func (r *crawlRuntime) run(start []Request) (Result, error) {
	defer r.cancel()
	r.result.Stats.StartTime = time.Now()
	if err := r.crawler.Sessions.Start(r.ctx); err != nil {
		return r.result, err
	}
	defer r.crawler.Sessions.Close(r.ctx)

	if err := r.prefetchRobots(start); err != nil {
		return r.finish(err)
	}
	resuming, err := r.restoreOrEnqueueStartRequests(start)
	if err != nil {
		return r.finish(err)
	}
	if r.crawler.OnStart != nil {
		if err := r.crawler.OnStart(r.ctx, resuming); err != nil {
			return r.finish(err)
		}
	}

	for {
		r.startReadyTasks()
		if r.isComplete() {
			if r.stopErr != nil {
				if r.checkpoint != nil && errors.Is(r.stopErr, r.ctx.Err()) {
					r.result.Paused = true
					if err := r.checkpoint.Save(r.pauseSnapshot()); err != nil {
						return r.finish(err)
					}
					return r.finish(nil)
				}
				return r.finish(r.stopErr)
			}
			if r.checkpoint != nil {
				if err := r.checkpoint.Cleanup(); err != nil {
					return r.finish(err)
				}
			}
			return r.finish(nil)
		}

		select {
		case task := <-r.taskResults:
			r.handleTaskResult(task)
		case <-r.doneChannel():
			r.stop(r.ctx.Err())
		}
	}
}

func (r *crawlRuntime) finish(err error) (Result, error) {
	r.result.Stats.EndTime = time.Now()
	r.result.Stats.Elapsed = r.result.Stats.EndTime.Sub(r.result.Stats.StartTime)
	if r.crawler.OnClose != nil {
		if closeErr := r.crawler.OnClose(r.ctx, r.result); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return r.result, err
}

func (r *crawlRuntime) prefetchRobots(start []Request) error {
	if !r.crawler.RobotsTxtObey || r.crawler.RobotsTxtManager == nil || len(start) == 0 {
		return nil
	}
	for _, request := range start {
		if err := r.crawler.RobotsTxtManager.Prefetch(r.ctx, []string{request.URL}, request.SID); err != nil {
			return err
		}
	}
	return nil
}

func (r *crawlRuntime) restoreOrEnqueueStartRequests(start []Request) (bool, error) {
	if r.checkpoint != nil {
		snapshot, ok, err := r.checkpoint.Load()
		if err != nil {
			return false, err
		}
		if ok {
			r.scheduler.Restore(snapshot)
			return true, nil
		}
	}
	for _, request := range start {
		queued, err := r.scheduler.Enqueue(request)
		if err != nil {
			return false, err
		}
		if !queued {
			r.result.Stats.Skipped++
		}
	}
	return false, nil
}

func (r *crawlRuntime) pauseSnapshot() SchedulerSnapshot {
	snapshot := r.scheduler.Snapshot()
	if len(r.activeRequests) == 0 {
		return snapshot
	}
	active := make([]Request, 0, len(r.activeRequests))
	for taskID := 1; taskID <= r.nextTaskID; taskID++ {
		request, ok := r.activeRequests[taskID]
		if ok {
			active = append(active, request.clone())
		}
	}
	snapshot.Requests = append(active, snapshot.Requests...)
	return snapshot
}

func (r *crawlRuntime) startReadyTasks() {
	if err := r.ctx.Err(); err != nil {
		r.stop(err)
		return
	}
	for r.stopErr == nil && r.active < r.concurrentRequests && r.scheduler.Len() > 0 {
		request, ok := r.scheduler.Dequeue()
		if !ok {
			break
		}
		r.active++
		r.nextTaskID++
		taskID := r.nextTaskID
		r.activeRequests[taskID] = request.clone()
		go func() {
			result := r.crawler.processRequest(r.ctx, request, r.domainLimiters, r.sleep)
			result.taskID = taskID
			r.taskResults <- result
		}()
	}
}

func (r *crawlRuntime) isComplete() bool {
	if r.active != 0 {
		return false
	}
	return r.stopErr != nil || r.scheduler.Len() == 0
}

func (r *crawlRuntime) doneChannel() <-chan struct{} {
	if r.stopErr != nil {
		return nil
	}
	return r.ctx.Done()
}

func (r *crawlRuntime) handleTaskResult(task crawlerTaskResult) {
	r.active--
	if task.err != nil {
		if r.ctx.Err() == nil || !errors.Is(task.err, r.ctx.Err()) {
			delete(r.activeRequests, task.taskID)
		}
		r.handleTaskError(task.request, task.err)
		return
	}
	delete(r.activeRequests, task.taskID)

	if task.robotsDisallowed {
		r.result.Stats.RobotsDisallowed++
		return
	}

	r.recordResponseStats(task.response)
	r.result.Stats.Requests++
	r.result.Stats.Sessions[task.response.Request.SID]++
	if task.blocked {
		r.result.Stats.BlockedRequests++
		if task.retry != nil {
			queued, err := r.scheduler.Enqueue(*task.retry)
			if err != nil {
				r.stop(err)
				return
			}
			if queued {
				r.result.Stats.BlockedRetries++
			} else {
				r.result.Stats.Skipped++
			}
		}
		return
	}
	for _, output := range task.outputs {
		r.handleOutput(output)
	}
}

func (r *crawlRuntime) recordResponseStats(response Response) {
	if response.Response == nil {
		return
	}
	status := response.StatusCode()
	if status != 0 {
		r.result.Stats.StatusCodes[status]++
	}
	bodyBytes := len(response.Body())
	r.result.Stats.ResponseBytes += bodyBytes
	domain := crawlerDomain(response.Request.URL)
	if domain != "" {
		r.result.Stats.DomainResponseBytes[domain] += bodyBytes
	}
}

func (r *crawlRuntime) handleTaskError(request Request, err error) {
	if r.ctx.Err() != nil && errors.Is(err, r.ctx.Err()) {
		r.stop(r.ctx.Err())
		return
	}
	r.result.Errors = append(r.result.Errors, err)
	r.result.Stats.Failed++
	if r.crawler.OnError != nil {
		if hookErr := r.crawler.OnError(r.ctx, request, err); hookErr != nil {
			r.result.Errors = append(r.result.Errors, hookErr)
		}
	}
}

func (r *crawlRuntime) handleOutput(output Output) {
	if output.Item != nil {
		item := cloneMeta(output.Item)
		if r.crawler.OnScrapedItem != nil {
			processed, err := r.crawler.OnScrapedItem(r.ctx, item)
			if err != nil {
				r.stop(err)
				return
			}
			if processed == nil {
				r.result.Stats.ItemsDropped++
				return
			}
			item = cloneMeta(processed)
		}
		r.result.Items = append(r.result.Items, item)
		r.result.Stats.Items++
		if r.itemStream != nil {
			select {
			case r.itemStream <- cloneMeta(item):
			case <-r.ctx.Done():
				r.stop(r.ctx.Err())
			}
		}
	}
	if output.Request == nil {
		return
	}
	if !isDomainAllowed(output.Request.URL, r.allowedDomains) {
		r.result.Stats.OffsiteRequests++
		return
	}
	queued, err := r.scheduler.Enqueue(*output.Request)
	if err != nil {
		r.stop(err)
		return
	}
	if !queued {
		r.result.Stats.Skipped++
	}
}

func (r *crawlRuntime) stop(err error) {
	if r.stopErr != nil {
		return
	}
	r.stopErr = err
	r.cancel()
}
