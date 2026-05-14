package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLivePracticeSitesEndToEnd(t *testing.T) {
	if os.Getenv("GOSCRAPLING_LIVE_E2E") != "1" {
		t.Skip("set GOSCRAPLING_LIVE_E2E=1 to run live practice-site scrapes")
	}

	binary := buildGoscraplingBinary(t)
	outputDir := t.TempDir()
	client := &http.Client{Timeout: 15 * time.Second}
	userAgentHeader := "User-Agent: " + liveE2EUserAgentValue

	tests := []struct {
		name        string
		url         string
		selector    string
		want        []string
		args        []string
		delayBefore time.Duration
	}{
		{
			name:     "books-to-scrape",
			url:      "http://books.toscrape.com/",
			selector: "article.product_pod h3 a",
			want:     []string{"Tipping the Velvet"},
		},
		{
			name:     "quotes-to-scrape",
			url:      "https://quotes.toscrape.com/",
			selector: ".quote .text",
			want:     []string{"The world as we have created it", "It is our choices"},
		},
		{
			name:     "scrape-this-site",
			url:      "https://www.scrapethissite.com/pages/simple/",
			selector: ".country-name",
			want:     []string{"Andorra", "Afghanistan"},
		},
		{
			name:     "oxylabs-sandbox",
			url:      "https://sandbox.oxylabs.io/products",
			selector: ".product-card h4",
			want:     []string{"Super Mario", "Metroid"},
		},
		{
			name:        "web-scraping-dev",
			url:         "https://web-scraping.dev/products",
			selector:    ".product h3",
			want:        []string{"Box of Chocolate Candy", "Energy Potion"},
			delayBefore: 2 * time.Second,
		},
		{
			name: "httpbin-headers",
			url:  "https://httpbin.org/headers",
			want: []string{"X-Goscrapling-Live", "sandbox"},
			args: []string{"-H", "X-Goscrapling-Live: sandbox"},
		},
		{
			name:     "crawler-test",
			url:      "https://crawler-test.com/",
			selector: "title",
			want:     []string{"Crawler Test Site"},
		},
		{
			name: "jsonplaceholder",
			url:  "https://jsonplaceholder.typicode.com/posts/1",
			want: []string{`"userId": 1`, `"id": 1`},
		},
		{
			name:     "mockaroo",
			url:      "https://www.mockaroo.com/",
			selector: "title",
			want:     []string{"Mockaroo", "Random Data Generator"},
		},
		{
			name:     "the-internet",
			url:      "https://the-internet.herokuapp.com/",
			selector: "h1",
			want:     []string{"Welcome to the-internet"},
		},
		{
			name:     "incolumitas-bot-test",
			url:      "https://bot.incolumitas.com/",
			selector: "title",
			want:     []string{"Bot", "Detection"},
		},
		{
			name:     "sannysoft-bot-test",
			url:      "https://bot.sannysoft.com/",
			selector: "title",
			want:     []string{"Antibot"},
		},
		{
			name:     "fingerprint-bot-detection",
			url:      "https://fingerprint.com/products/bot-detection/",
			selector: "h1",
			want:     []string{"Detect bots"},
		},
		{
			name:     "creepjs",
			url:      "https://abrahamjuliot.github.io/creepjs/",
			selector: "title",
			want:     []string{"CreepJS"},
		},
		{
			name:     "pixelscan",
			url:      "https://pixelscan.net/",
			selector: "title",
			want:     []string{"Pixelscan", "detection analysis"},
		},
		{
			name:     "real-python-fake-jobs",
			url:      "https://realpython.github.io/fake-jobs/",
			selector: ".card-content h2.title",
			want:     []string{"Senior Python Developer", "Software Engineer"},
		},
		{
			name:     "wikipedia",
			url:      "https://www.wikipedia.org/",
			selector: "strong.localized-slogan",
			want:     []string{"The Free Encyclopedia"},
		},
		{
			name:     "old-reddit",
			url:      "https://old.reddit.com/",
			selector: "title",
			want:     []string{"reddit", "front page"},
		},
		{
			name:     "security-crawl-maze",
			url:      "https://security-crawl-maze.app/",
			selector: "title",
			want:     []string{"CrawlMaze", "Web Crawlers"},
		},
		{
			name:     "github-topics",
			url:      "https://github.com/topics",
			selector: "h1",
			want:     []string{"Topics"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			robots := fetchRobotsDecision(ctx, client, tt.url, liveE2EUserAgentValue)
			if !robots.allowed {
				t.Skipf("robots: %s", robots.reason)
			}

			delay := maxDuration(tt.delayBefore, robots.crawlDelay)
			if delay > 0 {
				t.Logf("honoring robots/test delay: %s", delay)
				time.Sleep(delay)
			}

			outputPath := filepath.Join(outputDir, tt.name+".txt")
			args := []string{"extract", "get", tt.url, outputPath, "--timeout", "15"}
			if tt.selector != "" {
				args = append(args, "--css-selector", tt.selector)
			}
			args = append(args, "-H", userAgentHeader)
			args = append(args, tt.args...)

			result := runGoscraplingBinary(t, binary, args...)
			if result.err != nil {
				t.Fatalf("live scrape failed: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
			}

			body, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			text := string(body)
			if strings.TrimSpace(text) == "" {
				t.Fatalf("empty extraction for %s selector %q", tt.url, tt.selector)
			}
			for _, want := range tt.want {
				if !strings.Contains(text, want) {
					t.Fatalf("output missing %q:\n%s", want, text)
				}
			}
			t.Logf("%s extracted %d bytes", tt.url, len(body))
		})
	}
}

func maxDuration(a time.Duration, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
