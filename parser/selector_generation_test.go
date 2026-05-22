package parser

import (
	"strings"
	"testing"
)

func TestSelectorGeneration(t *testing.T) {
	const html = `
<!doctype html>
<html>
  <body>
    <main id="shop">
      <section class="products">
        <article class="product featured" data-id="p1">
          <h2>Trail Kit</h2>
          <a href="/products/trail-kit">Details</a>
        </article>
        <article class="product" data-id="p2">
          <h2>Camp Mug</h2>
          <a href="/products/camp-mug">Details</a>
        </article>
      </section>
      <section class="reviews">
        <article class="review">Great product!</article>
      </section>
    </main>
  </body>
</html>`

	doc, err := Parse(strings.NewReader(html), ParseOptions{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	link, ok := doc.CSS(`article[data-id="p2"] a`).First()
	if !ok {
		t.Fatal("expected camp mug link")
	}

	t.Run("short selectors anchor at the nearest id", func(t *testing.T) {
		if got, want := link.GenerateCSSSelector(), `#shop > section > article:nth-of-type(2) > a`; got != want {
			t.Fatalf("GenerateCSSSelector() = %q, want %q", got, want)
		}
		assertSelectorReselects(t, doc.CSS(link.GenerateCSSSelector()), link)

		if got, want := link.GenerateXPathSelector(), `//*[@id='shop']/section/article[2]/a`; got != want {
			t.Fatalf("GenerateXPathSelector() = %q, want %q", got, want)
		}
		assertSelectorReselects(t, doc.XPath(link.GenerateXPathSelector()), link)
	})

	t.Run("full selectors remain deterministic and include the document path", func(t *testing.T) {
		if got, want := link.GenerateFullCSSSelector(), `body > #shop > section > article:nth-of-type(2) > a`; got != want {
			t.Fatalf("GenerateFullCSSSelector() = %q, want %q", got, want)
		}
		assertSelectorReselects(t, doc.CSS(link.GenerateFullCSSSelector()), link)

		if got, want := link.GenerateFullXPathSelector(), `//body/*[@id='shop']/section/article[2]/a`; got != want {
			t.Fatalf("GenerateFullXPathSelector() = %q, want %q", got, want)
		}
		assertSelectorReselects(t, doc.XPath(link.GenerateFullXPathSelector()), link)

		if again := link.GenerateFullCSSSelector(); again != link.GenerateFullCSSSelector() {
			t.Fatalf("full CSS selector is not deterministic: %q then %q", again, link.GenerateFullCSSSelector())
		}
	})
}

func assertSelectorReselects(t *testing.T, selection Selection, expected *Element) {
	t.Helper()
	if got := selection.Len(); got != 1 {
		t.Fatalf("selector match count = %d, want 1", got)
	}
	actual, ok := selection.First()
	if !ok || actual.node != expected.node {
		t.Fatalf("selector selected %#v ok=%v, want original element", actual, ok)
	}
}
