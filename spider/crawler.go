package spider

import (
	"context"
	"fmt"
)

type Crawler struct {
	Sessions        *SessionManager
	Scheduler       *Scheduler
	DefaultCallback Callback
}

func (c Crawler) Run(ctx context.Context, start []Request) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.Sessions == nil {
		return Result{}, fmt.Errorf("sessions are required")
	}

	scheduler := c.Scheduler
	if scheduler == nil {
		scheduler = NewScheduler(SchedulerOptions{})
	}

	result := Result{Stats: Stats{Sessions: make(map[string]int)}}
	if err := c.Sessions.Start(ctx); err != nil {
		return result, err
	}
	defer c.Sessions.Close(ctx)

	for _, request := range start {
		queued, err := scheduler.Enqueue(request)
		if err != nil {
			return result, err
		}
		if !queued {
			result.Stats.Skipped++
		}
	}

	for scheduler.Len() > 0 {
		request, ok := scheduler.Dequeue()
		if !ok {
			break
		}

		response, err := c.Sessions.Fetch(ctx, request)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.Stats.Failed++
			continue
		}
		result.Stats.Requests++
		result.Stats.Sessions[response.Request.SID]++

		callback := request.Callback
		if callback == nil {
			callback = c.DefaultCallback
		}
		if callback == nil {
			continue
		}

		outputs, err := callback(ctx, response)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.Stats.Failed++
			continue
		}
		for _, output := range outputs {
			if output.Item != nil {
				result.Items = append(result.Items, cloneMeta(output.Item))
				result.Stats.Items++
			}
			if output.Request != nil {
				queued, err := scheduler.Enqueue(*output.Request)
				if err != nil {
					return result, err
				}
				if !queued {
					result.Stats.Skipped++
				}
			}
		}
	}

	return result, nil
}
