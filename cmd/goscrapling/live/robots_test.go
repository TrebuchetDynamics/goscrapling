package live_test

import (
	"strings"
	"testing"
	"time"
)

func TestEvaluateRobotsAllowsAndDisallowsPaths(t *testing.T) {
	decision := evaluateRobots([]byte(`
User-agent: *
Disallow: /private/
Allow: /private/open
`), "/private/closed", "goscrapling-live-e2e")
	if decision.allowed {
		t.Fatalf("expected /private/closed to be disallowed: %+v", decision)
	}
	if !strings.Contains(decision.reason, "disallowed") {
		t.Fatalf("decision reason = %q, want disallowed", decision.reason)
	}

	decision = evaluateRobots([]byte(`
User-agent: *
Disallow: /private/
Allow: /private/open
`), "/private/open/page", "goscrapling-live-e2e")
	if !decision.allowed {
		t.Fatalf("expected /private/open/page to be allowed: %+v", decision)
	}
}

func TestEvaluateRobotsUsesSpecificUserAgentAndCrawlDelay(t *testing.T) {
	decision := evaluateRobots([]byte(`
User-agent: *
Disallow: /
Crawl-delay: 9

User-agent: goscrapling-live-e2e
Disallow: /blocked
Crawl-delay: 2
`), "/allowed", "goscrapling-live-e2e")
	if !decision.allowed {
		t.Fatalf("expected specific user-agent group to allow /allowed: %+v", decision)
	}
	if decision.crawlDelay != 2*time.Second {
		t.Fatalf("crawlDelay = %s, want 2s", decision.crawlDelay)
	}

	decision = evaluateRobots([]byte(`
User-agent: *
Disallow: /

User-agent: goscrapling-live-e2e
Disallow: /blocked
`), "/blocked/path", "goscrapling-live-e2e")
	if decision.allowed {
		t.Fatalf("expected specific user-agent disallow to apply: %+v", decision)
	}
}

func TestEvaluateRobotsRejectsHTMLRobotsBody(t *testing.T) {
	decision := evaluateRobots([]byte(`<!doctype html><title>not robots</title>`), "/", "goscrapling-live-e2e")
	if decision.allowed {
		t.Fatalf("expected HTML robots response to be unusable: %+v", decision)
	}
	if !strings.Contains(decision.reason, "not a usable robots.txt") {
		t.Fatalf("decision reason = %q, want unusable robots", decision.reason)
	}
}

func TestEvaluateRobotsAllowsRobotsWithNoMatchingRules(t *testing.T) {
	decision := evaluateRobots([]byte(`# content signals only`), "/posts/1", "goscrapling-live-e2e")
	if !decision.allowed {
		t.Fatalf("expected robots file with no matching directives to allow by default: %+v", decision)
	}
	if !strings.Contains(decision.reason, "no matching robots rules") {
		t.Fatalf("decision reason = %q, want no matching rules", decision.reason)
	}
}
