package browser

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestBrowserActionsAndCapture(t *testing.T) {
	t.Run("passes automation and capture controls through the engine", func(t *testing.T) {
		screenshot := []byte("png screenshot bytes")
		downloadBody := []byte("download bytes")
		engine := &recordingBrowserEngine{result: BrowserResult{
			URL:        "https://example.com/download.bin",
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"application/octet-stream"}},
			Body:       downloadBody,
			Screenshot: screenshot,
			CapturedXHR: []BrowserResult{{
				URL:        "https://example.com/api/data",
				StatusCode: http.StatusAccepted,
				Reason:     "Accepted",
				Headers:    http.Header{"Content-Type": []string{"application/json"}},
				Body:       []byte(`{"ok":true}`),
			}},
		}}
		fetcher := BrowserFetcher{Engine: engine}
		actions := []BrowserAction{
			{Kind: BrowserActionClick, Selector: "#load"},
			{Kind: BrowserActionFill, Selector: "#query", Value: "scrapling"},
			{Kind: BrowserActionEvaluate, Value: "window.ready = true"},
			{Kind: BrowserActionWaitForSelector, Selector: ".result"},
		}

		response, err := fetcher.Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			NetworkIdle:  true,
			LoadDOM:      true,
			Wait:         75 * time.Millisecond,
			WaitSelector: BrowserWaitSelector{Selector: ".result", State: BrowserWaitVisible},
			Actions:      actions,
			CaptureXHR:   `^https://example\.com/api/`,
			Screenshot: BrowserScreenshotOptions{
				Enabled:  true,
				FullPage: true,
				Quality:  90,
			},
		})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}

		request := engine.request
		if !request.NetworkIdle || !request.LoadDOM || request.Wait != 75*time.Millisecond {
			t.Fatalf("wait controls were not forwarded: %#v", request)
		}
		if request.WaitSelector.Selector != ".result" || request.WaitSelector.State != BrowserWaitVisible {
			t.Fatalf("wait selector = %#v", request.WaitSelector)
		}
		if !reflect.DeepEqual(request.Actions, actions) {
			t.Fatalf("actions = %#v, want %#v", request.Actions, actions)
		}
		if request.CaptureXHR != `^https://example\.com/api/` {
			t.Fatalf("capture XHR pattern = %q", request.CaptureXHR)
		}
		if !request.Screenshot.Enabled || !request.Screenshot.FullPage || request.Screenshot.Quality != 90 {
			t.Fatalf("screenshot options = %#v", request.Screenshot)
		}

		if !bytes.Equal(response.Body(), downloadBody) {
			t.Fatalf("response body = %q, want download body %q", response.Body(), downloadBody)
		}
		shot, ok := response.Meta()["screenshot"].([]byte)
		if !ok || !bytes.Equal(shot, screenshot) {
			t.Fatalf("response screenshot metadata = %#v, want %q", response.Meta()["screenshot"], screenshot)
		}
		xhr := response.CapturedXHR()
		if len(xhr) != 1 {
			t.Fatalf("captured XHR count = %d, want 1", len(xhr))
		}
		if xhr[0].URL() != "https://example.com/api/data" || xhr[0].StatusCode() != http.StatusAccepted {
			t.Fatalf("captured XHR response = url %q status %d", xhr[0].URL(), xhr[0].StatusCode())
		}
		if got := xhr[0].Headers().Get("Content-Type"); got != "application/json" {
			t.Fatalf("captured XHR content type = %q", got)
		}
		if !bytes.Equal(xhr[0].Body(), []byte(`{"ok":true}`)) {
			t.Fatalf("captured XHR body = %q", xhr[0].Body())
		}
	})

	t.Run("validates capture patterns before engine fetch", func(t *testing.T) {
		engine := &recordingBrowserEngine{}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			CaptureXHR: "[",
		})
		if !errors.Is(err, ErrBrowserOptions) {
			t.Fatalf("Fetch error = %v, want ErrBrowserOptions", err)
		}
		if engine.request.URL != "" {
			t.Fatalf("engine was called for invalid options: %#v", engine.request)
		}
	})

	t.Run("validates action shapes before engine fetch", func(t *testing.T) {
		engine := &recordingBrowserEngine{}
		_, err := (BrowserFetcher{Engine: engine}).Fetch(context.Background(), "https://example.com/app", BrowserOptions{
			Actions: []BrowserAction{{Kind: BrowserActionClick}},
		})
		if !errors.Is(err, ErrBrowserOptions) {
			t.Fatalf("Fetch error = %v, want ErrBrowserOptions", err)
		}
		if engine.request.URL != "" {
			t.Fatalf("engine was called for invalid options: %#v", engine.request)
		}
	})
}
