package goscrapling

import (
	"bytes"
	"strings"
	"testing"
)

func TestXPathAndTranslator(t *testing.T) {
	const html = `
<!doctype html>
<html>
  <body>
    <section id="catalog">
      <article class="product featured" data-id="p1">
        <h2>Trail Kit</h2>
        <a href="/products/trail-kit">Details</a>
      </article>
      <article class="product" data-id="p2">
        <h2>Camp Mug</h2>
        <a href="/products/camp-mug">Details</a>
      </article>
    </section>
    <aside class="featured"><h2>Newsletter</h2></aside>
  </body>
</html>`

	doc, err := Parse(strings.NewReader(html), ParseOptions{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	t.Run("document xpath selects elements text and attributes", func(t *testing.T) {
		if got := doc.XPath(`//article`).Len(); got != 2 {
			t.Fatalf("article count = %d, want 2", got)
		}
		if got := doc.XPath(`//article[@data-id="p1"]//h2/text()`).Text(); got != "Trail Kit" {
			t.Fatalf("title text = %q, want Trail Kit", got)
		}
		if got := doc.XPath(`//article//a/@href`).Text(); got != "/products/trail-kit\n/products/camp-mug" {
			t.Fatalf("href text = %q, want product hrefs", got)
		}
		if got := doc.XPath(`(//article)[1]`).Text(); !strings.Contains(got, "Trail Kit") || strings.Contains(got, "Camp Mug") {
			t.Fatalf("first article text = %q, want only Trail Kit article", got)
		}
	})

	t.Run("response xpath delegates to parsed document", func(t *testing.T) {
		response, err := NewResponse(bytes.NewReader([]byte(html)), ResponseOptions{URL: "https://example.com/products"})
		if err != nil {
			t.Fatalf("NewResponse: %v", err)
		}
		if got := response.XPath(`//article[@data-id="p2"]//h2/text()`).Text(); got != "Camp Mug" {
			t.Fatalf("response xpath text = %q, want Camp Mug", got)
		}
	})

	t.Run("element and selection xpath are relative", func(t *testing.T) {
		first, ok := doc.XPath(`//article[@data-id="p1"]`).First()
		if !ok {
			t.Fatal("expected first product")
		}
		if got := first.XPath(`.//a/@href`).Get(); got != "/products/trail-kit" {
			t.Fatalf("relative element href = %q, want trail kit href", got)
		}

		articles := doc.XPath(`//section[@id="catalog"]`).XPath(`.//article`)
		if got := articles.Len(); got != 2 {
			t.Fatalf("relative selection count = %d, want 2", got)
		}
	})

	t.Run("css translator output selects equivalent xpath nodes", func(t *testing.T) {
		tests := []struct {
			name string
			css  string
			want string
		}{
			{name: "descendant text pseudo", css: `.product h2::text`, want: "Trail Kit\nCamp Mug"},
			{name: "child combinator", css: `section#catalog > article.product h2::text`, want: "Trail Kit\nCamp Mug"},
			{name: "attribute contains attr pseudo", css: `article[data-id="p1"] a[href*="/products/"]::attr(href)`, want: "/products/trail-kit"},
			{name: "grouped selectors", css: `article.product, aside.featured`, want: "Trail Kit Details\nCamp Mug Details\nNewsletter"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				xpath, err := CSSToXPath(tt.css)
				if err != nil {
					t.Fatalf("CSSToXPath(%q): %v", tt.css, err)
				}
				if got := doc.XPath(xpath).Text(); got != tt.want {
					t.Fatalf("translated XPath %q text = %q, want %q", xpath, got, tt.want)
				}
			})
		}
	})

	t.Run("translator rejects unsupported selectors explicitly", func(t *testing.T) {
		if _, err := CSSToXPath(`article:nth-child(1)`); err == nil {
			t.Fatal("expected unsupported selector error")
		}
	})
}
