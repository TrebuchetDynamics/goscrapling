package browser

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type ChromedpBrowserOptions struct {
	ExecutablePath string
}

type ChromedpBrowserEngine struct {
	options ChromedpBrowserOptions
}

func NewChromedpBrowserEngine(opts ChromedpBrowserOptions) *ChromedpBrowserEngine {
	return &ChromedpBrowserEngine{options: opts}
}

func (e *ChromedpBrowserEngine) Fetch(ctx context.Context, request BrowserRequest) (BrowserResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	allocOptions := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	allocOptions = append(allocOptions,
		chromedp.Flag("headless", request.Headless),
		chromedp.NoSandbox,
	)
	if e != nil && e.options.ExecutablePath != "" {
		allocOptions = append(allocOptions, chromedp.ExecPath(e.options.ExecutablePath))
	}

	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(ctx, allocOptions...)
	defer allocatorCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)
	defer browserCancel()

	var renderedHTML string
	actions := []chromedp.Action{network.Enable()}
	if len(request.Headers) > 0 {
		actions = append(actions, network.SetExtraHTTPHeaders(chromedpNetworkHeaders(request.Headers)))
	}
	actions = append(actions, chromedp.Navigate(request.URL))
	if request.LoadDOM {
		actions = append(actions, chromedp.WaitReady("body", chromedp.ByQuery))
	}
	actions = append(actions, chromedpActions(request.Actions)...)
	if request.WaitSelector.Selector != "" {
		actions = append(actions, chromedpWaitSelector(request.WaitSelector))
	}
	if request.NetworkIdle {
		actions = append(actions, chromedp.Sleep(500*time.Millisecond))
	}
	if request.Wait > 0 {
		actions = append(actions, chromedp.Sleep(request.Wait))
	}
	actions = append(actions, chromedp.OuterHTML("html", &renderedHTML, chromedp.ByQuery))

	response, err := chromedp.RunResponse(browserCtx, actions...)
	if err != nil {
		return BrowserResult{}, err
	}

	resultURL := request.URL
	statusCode := http.StatusOK
	reason := http.StatusText(http.StatusOK)
	headers := http.Header{}
	if response != nil {
		if response.URL != "" {
			resultURL = response.URL
		}
		if response.Status > 0 {
			statusCode = int(response.Status)
		}
		if response.StatusText != "" {
			reason = response.StatusText
		} else if statusText := http.StatusText(statusCode); statusText != "" {
			reason = statusText
		}
		headers = httpHeadersFromChromedp(response.Headers)
		if headers.Get("Content-Type") == "" && response.MimeType != "" {
			headers.Set("Content-Type", response.MimeType)
		}
	}

	return BrowserResult{
		URL:        resultURL,
		StatusCode: statusCode,
		Reason:     reason,
		Headers:    headers,
		Body:       []byte(renderedHTML),
	}, nil
}

func chromedpActions(actions []BrowserAction) []chromedp.Action {
	output := make([]chromedp.Action, 0, len(actions))
	for _, action := range actions {
		switch action.Kind {
		case BrowserActionClick:
			output = append(output, chromedp.Click(action.Selector, chromedp.ByQuery))
		case BrowserActionFill:
			output = append(output, chromedp.SetValue(action.Selector, action.Value, chromedp.ByQuery))
		case BrowserActionWaitForSelector:
			output = append(output, chromedp.WaitReady(action.Selector, chromedp.ByQuery))
		case BrowserActionEvaluate:
			output = append(output, chromedp.Evaluate(action.Value, nil))
		}
	}
	return output
}

func chromedpWaitSelector(wait BrowserWaitSelector) chromedp.Action {
	switch wait.State {
	case BrowserWaitDetached:
		return chromedp.WaitNotPresent(wait.Selector, chromedp.ByQuery)
	case BrowserWaitHidden:
		return chromedp.WaitNotVisible(wait.Selector, chromedp.ByQuery)
	case BrowserWaitVisible:
		return chromedp.WaitVisible(wait.Selector, chromedp.ByQuery)
	default:
		return chromedp.WaitReady(wait.Selector, chromedp.ByQuery)
	}
}

func chromedpNetworkHeaders(headers http.Header) network.Headers {
	output := network.Headers{}
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		output[key] = strings.Join(values, ", ")
	}
	return output
}

func httpHeadersFromChromedp(headers network.Headers) http.Header {
	output := http.Header{}
	for key, value := range headers {
		switch typed := value.(type) {
		case string:
			output.Add(key, typed)
		case []string:
			for _, item := range typed {
				output.Add(key, item)
			}
		case []any:
			for _, item := range typed {
				output.Add(key, fmt.Sprint(item))
			}
		default:
			output.Add(key, fmt.Sprint(value))
		}
	}
	return output
}
