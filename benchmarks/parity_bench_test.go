package benchmarks

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func BenchmarkParserNestedText(b *testing.B) {
	html := nestedHTML(500)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		doc, err := goscrapling.Parse(strings.NewReader(html), goscrapling.ParseOptions{URL: "https://example.test/products"})
		if err != nil {
			b.Fatalf("parse fixture: %v", err)
		}
		if got := doc.CSS("span.leaf::text").Get().String(); got != "needle" {
			b.Fatalf("leaf text = %q, want needle", got)
		}
	}
}

func BenchmarkStaticFetcherLocalResponse(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<html><body><h1>fixture</h1></body></html>`)
	}))
	b.Cleanup(server.Close)

	fetcher := goscrapling.Fetcher{Client: server.Client()}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		response, err := fetcher.Get(server.URL, goscrapling.RequestOptions{})
		if err != nil {
			b.Fatalf("fetch fixture: %v", err)
		}
		if got := response.CSS("h1::text").Get().String(); got != "fixture" {
			b.Fatalf("title = %q, want fixture", got)
		}
	}
}

func BenchmarkSpiderSchedulerFingerprint(b *testing.B) {
	requests := make([]spiders.Request, 256)
	for i := range requests {
		requests[i] = spiders.Request{
			URL:      "https://example.test/products/" + strconv.Itoa(i%64) + "?page=" + strconv.Itoa(i%8),
			Priority: i % 7,
			SID:      "default",
		}
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		scheduler := spiders.NewScheduler(spiders.SchedulerOptions{})
		for _, request := range requests {
			if _, err := scheduler.Enqueue(request); err != nil {
				b.Fatalf("enqueue fixture request: %v", err)
			}
		}
		for scheduler.Len() > 0 {
			if _, ok := scheduler.Dequeue(); !ok {
				b.Fatal("scheduler reported length but dequeue failed")
			}
		}
	}
}

func BenchmarkCLIExtractFixture(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<html><body><article class="card">fixture</article></body></html>`)
	}))
	b.Cleanup(server.Close)

	outputPath := filepath.Join(b.TempDir(), "extract.txt")
	args := []string{"extract", "get", server.URL, outputPath, "--css-selector", ".card::text"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var stdout, stderr bytes.Buffer
		if err := cli.Run(&stdout, &stderr, args); err != nil {
			b.Fatalf("cli extract fixture: %v\nstderr: %s", err, stderr.String())
		}
	}
}

func nestedHTML(depth int) string {
	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString(`<div class="node">`)
	}
	b.WriteString(`<span class="leaf">needle</span>`)
	for i := 0; i < depth; i++ {
		b.WriteString(`</div>`)
	}
	return b.String()
}
