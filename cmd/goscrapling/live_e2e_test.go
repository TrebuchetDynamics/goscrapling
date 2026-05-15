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

const (
	liveE2EFeatureCSSSelector   = "css-selector"
	liveE2EFeatureRawDocument   = "raw-document"
	liveE2EFeatureCustomHeader  = "custom-header"
	minLiveE2ECompletedScrapes  = 5
	minLiveE2ECSSSelectorScrape = 3
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
		feature     string
		delayBefore time.Duration
	}{
		{
			name:     "books-to-scrape",
			url:      "http://books.toscrape.com/",
			selector: "article.product_pod h3 a",
			want:     []string{"Tipping the Velvet"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "quotes-to-scrape",
			url:      "https://quotes.toscrape.com/",
			selector: ".quote .text",
			want:     []string{"The world as we have created it", "It is our choices"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "scrape-this-site",
			url:      "https://www.scrapethissite.com/pages/simple/",
			selector: ".country-name",
			want:     []string{"Andorra", "Afghanistan"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "oxylabs-sandbox",
			url:      "https://sandbox.oxylabs.io/products",
			selector: ".product-card h4",
			want:     []string{"Super Mario", "Metroid"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:        "web-scraping-dev",
			url:         "https://web-scraping.dev/products",
			selector:    ".product h3",
			want:        []string{"Box of Chocolate Candy", "Energy Potion"},
			feature:     liveE2EFeatureCSSSelector,
			delayBefore: 2 * time.Second,
		},
		{
			name:    "httpbin-headers",
			url:     "https://httpbin.org/headers",
			want:    []string{"X-Goscrapling-Live", "sandbox"},
			args:    []string{"-H", "X-Goscrapling-Live: sandbox"},
			feature: liveE2EFeatureCustomHeader,
		},
		{
			name:     "crawler-test",
			url:      "https://crawler-test.com/",
			selector: "title",
			want:     []string{"Crawler Test Site"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:    "jsonplaceholder",
			url:     "https://jsonplaceholder.typicode.com/posts/1",
			want:    []string{`"userId": 1`, `"id": 1`},
			feature: liveE2EFeatureRawDocument,
		},
		{
			name:     "mockaroo",
			url:      "https://www.mockaroo.com/",
			selector: "title",
			want:     []string{"Mockaroo", "Random Data Generator"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "the-internet",
			url:      "https://the-internet.herokuapp.com/",
			selector: "h1",
			want:     []string{"Welcome to the-internet"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "incolumitas-bot-test",
			url:      "https://bot.incolumitas.com/",
			selector: "title",
			want:     []string{"Bot", "Detection"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "sannysoft-bot-test",
			url:      "https://bot.sannysoft.com/",
			selector: "title",
			want:     []string{"Antibot"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "fingerprint-bot-detection",
			url:      "https://fingerprint.com/products/bot-detection/",
			selector: "h1",
			want:     []string{"Detect bots"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "creepjs",
			url:      "https://abrahamjuliot.github.io/creepjs/",
			selector: "title",
			want:     []string{"CreepJS"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "pixelscan",
			url:      "https://pixelscan.net/",
			selector: "title",
			want:     []string{"Pixelscan", "detection analysis"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "real-python-fake-jobs",
			url:      "https://realpython.github.io/fake-jobs/",
			selector: ".card-content h2.title",
			want:     []string{"Senior Python Developer", "Software Engineer"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "wikipedia",
			url:      "https://www.wikipedia.org/",
			selector: "strong.localized-slogan",
			want:     []string{"The Free Encyclopedia"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "old-reddit",
			url:      "https://old.reddit.com/",
			selector: "title",
			want:     []string{"reddit", "front page"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "security-crawl-maze",
			url:      "https://security-crawl-maze.app/",
			selector: "title",
			want:     []string{"CrawlMaze", "Web Crawlers"},
			feature:  liveE2EFeatureCSSSelector,
		},
		{
			name:     "github-topics",
			url:      "https://github.com/topics",
			selector: "h1",
			want:     []string{"Topics"},
			feature:  liveE2EFeatureCSSSelector,
		},
	}

	completedScrapes := 0
	completedFeatures := map[string]int{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.feature == "" {
				t.Fatalf("live E2E case %q is missing a feature lane", tt.name)
			}
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
			completedScrapes++
			completedFeatures[tt.feature]++
		})
	}

	if completedScrapes < minLiveE2ECompletedScrapes {
		t.Fatalf("completed %d real live scrapes, want at least %d; feature counts: %v", completedScrapes, minLiveE2ECompletedScrapes, completedFeatures)
	}
	if completedFeatures[liveE2EFeatureCSSSelector] < minLiveE2ECSSSelectorScrape {
		t.Fatalf("completed %d live CSS selector scrapes, want at least %d; feature counts: %v", completedFeatures[liveE2EFeatureCSSSelector], minLiveE2ECSSSelectorScrape, completedFeatures)
	}
	if completedFeatures[liveE2EFeatureRawDocument] == 0 {
		t.Fatalf("completed no live raw-document scrape; feature counts: %v", completedFeatures)
	}
	if completedFeatures[liveE2EFeatureCustomHeader] == 0 {
		t.Fatalf("completed no live custom-header scrape; feature counts: %v", completedFeatures)
	}
	t.Logf("completed %d real live scrapes by feature: %v", completedScrapes, completedFeatures)
}

func maxDuration(a time.Duration, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
