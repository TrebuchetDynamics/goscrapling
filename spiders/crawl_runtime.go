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
	result             Result
	active             int
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

	return &crawlRuntime{
		crawler:            crawler,
		ctx:                runCtx,
		cancel:             cancel,
		allowedDomains:     normalizeAllowedDomains(crawler.AllowedDomains),
		concurrentRequests: concurrentRequests,
		domainLimiters:     newCrawlerDomainLimiters(crawler.ConcurrentRequestsPerDomain),
		sleep:              sleep,
		scheduler:          scheduler,
		result: Result{Stats: Stats{
			ConcurrentRequests:          concurrentRequests,
			ConcurrentRequestsPerDomain: crawler.ConcurrentRequestsPerDomain,
			DownloadDelay:               crawler.DownloadDelay,
			Sessions:                    make(map[string]int),
		}},
		taskResults: make(chan crawlerTaskResult),
	}
}

func (r *crawlRuntime) run(start []Request) (Result, error) {
	defer r.cancel()
	if err := r.crawler.Sessions.Start(r.ctx); err != nil {
		return r.result, err
	}
	defer r.crawler.Sessions.Close(r.ctx)

	if err := r.enqueueStartRequests(start); err != nil {
		return r.result, err
	}

	for {
		r.startReadyTasks()
		if r.isComplete() {
			if r.stopErr != nil {
				return r.result, r.stopErr
			}
			return r.result, nil
		}

		select {
		case task := <-r.taskResults:
			r.handleTaskResult(task)
		case <-r.doneChannel():
			r.stop(r.ctx.Err())
		}
	}
}

func (r *crawlRuntime) enqueueStartRequests(start []Request) error {
	for _, request := range start {
		queued, err := r.scheduler.Enqueue(request)
		if err != nil {
			return err
		}
		if !queued {
			r.result.Stats.Skipped++
		}
	}
	return nil
}

func (r *crawlRuntime) startReadyTasks() {
	for r.stopErr == nil && r.active < r.concurrentRequests && r.scheduler.Len() > 0 {
		request, ok := r.scheduler.Dequeue()
		if !ok {
			break
		}
		r.active++
		go func() {
			r.taskResults <- r.crawler.processRequest(r.ctx, request, r.domainLimiters, r.sleep)
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
		r.handleTaskError(task.err)
		return
	}

	r.result.Stats.Requests++
	r.result.Stats.Sessions[task.response.Request.SID]++
	for _, output := range task.outputs {
		r.handleOutput(output)
	}
}

func (r *crawlRuntime) handleTaskError(err error) {
	if r.ctx.Err() != nil && errors.Is(err, r.ctx.Err()) {
		r.stop(r.ctx.Err())
		return
	}
	r.result.Errors = append(r.result.Errors, err)
	r.result.Stats.Failed++
}

func (r *crawlRuntime) handleOutput(output Output) {
	if output.Item != nil {
		r.result.Items = append(r.result.Items, cloneMeta(output.Item))
		r.result.Stats.Items++
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
