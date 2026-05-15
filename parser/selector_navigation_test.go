package parser

import (
	"regexp"
	"strings"
	"testing"
)

func TestSelectorNavigationAndSearch(t *testing.T) {
	const html = `
<!doctype html>
<html>
  <body>
    <main id="shop">
      <section class="products">
        <article class="product featured" data-id="p1" data-category="kit" data-price="10">
          <h2>Trail Kit</h2>
          <p class="stock">In stock: 5</p>
          <a href="/products/trail-kit">Details</a>
        </article>
        <article class="product" data-id="p2" data-category="kit" data-price="15">
          <h2>Camp Mug</h2>
          <p class="stock">In stock: 2</p>
          <a href="/products/camp-mug">Details</a>
        </article>
        <article class="product disabled" data-id="p3" data-category="kit" data-price="0">
          <h2>Repair Kit</h2>
          <p class="stock">Out of stock</p>
          <a href="/products/repair-kit">Details</a>
        </article>
      </section>
      <section class="reviews">
        <article class="review" data-rating="5">Great product!</article>
        <article class="review" data-rating="3">Works fine</article>
      </section>
    </main>
  </body>
</html>`

	doc, err := Parse(strings.NewReader(html), ParseOptions{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	t.Run("element traversal follows parent child sibling and ancestor relationships", func(t *testing.T) {
		camp, ok := doc.CSS(`article[data-id="p2"]`).First()
		if !ok {
			t.Fatal("expected camp mug product")
		}

		parent, ok := camp.Parent()
		if !ok || parent.TagName() != "section" {
			t.Fatalf("parent = %#v ok=%v, want section", parent, ok)
		}
		if got := parent.Children().Len(); got != 3 {
			t.Fatalf("parent children = %d, want 3 product articles", got)
		}
		if got := camp.Siblings().Len(); got != 2 {
			t.Fatalf("siblings = %d, want 2", got)
		}

		next, ok := camp.Next()
		if !ok || !next.HasClass("disabled") {
			t.Fatalf("next = %#v ok=%v, want disabled repair kit", next, ok)
		}
		previous, ok := camp.Previous()
		if !ok || previous.Text() != "Trail Kit In stock: 5 Details" {
			t.Fatalf("previous text = %q ok=%v, want Trail Kit product", previous.Text(), ok)
		}

		ancestor, ok := camp.FindAncestor(func(element *Element) bool {
			value, _ := element.Attr("id")
			return value == "shop"
		})
		if !ok || ancestor.TagName() != "main" {
			t.Fatalf("ancestor = %#v ok=%v, want main#shop", ancestor, ok)
		}
		if got := camp.Ancestors().Filter(func(element *Element) bool { return element.TagName() == "section" }).Len(); got != 1 {
			t.Fatalf("section ancestors = %d, want 1", got)
		}
	})

	t.Run("selection filter and search are chainable", func(t *testing.T) {
		products := doc.CSS("article.product")
		available := products.Filter(func(element *Element) bool {
			value, _ := element.Attr("data-price")
			return value != "0"
		})
		featured := available.Filter(func(element *Element) bool {
			return element.HasClass("featured")
		})
		if got := available.Len(); got != 2 {
			t.Fatalf("available products = %d, want 2", got)
		}
		if got := featured.Text(); got != "Trail Kit In stock: 5 Details" {
			t.Fatalf("featured text = %q, want Trail Kit product", got)
		}

		found, ok := products.Search(func(element *Element) bool {
			value, _ := element.Attr("data-id")
			return value == "p2"
		})
		if !ok || found.Text() != "Camp Mug In stock: 2 Details" {
			t.Fatalf("search result text = %q ok=%v, want Camp Mug product", found.Text(), ok)
		}
		if _, ok := products.Filter(func(element *Element) bool { return element.HasClass("missing") }).Search(func(*Element) bool { return true }); ok {
			t.Fatal("expected search on empty selection to miss")
		}
	})

	t.Run("find by text and regex support first and all matches", func(t *testing.T) {
		firstStock, ok := doc.FindByRegex(regexp.MustCompile(`In stock: \d+`), TextSearchOptions{CaseSensitive: true})
		if !ok || firstStock.Text() != "In stock: 5" {
			t.Fatalf("first regex text = %q ok=%v, want first stock", firstStock.Text(), ok)
		}
		allStock, err := doc.FindAllByRegex(`In stock: \d+`, TextSearchOptions{})
		if err != nil {
			t.Fatalf("FindAllByRegex: %v", err)
		}
		if got := allStock.Len(); got != 2 {
			t.Fatalf("all stock matches = %d, want 2", got)
		}
		outOfStock := doc.FindAllByText("Out of stock", TextSearchOptions{})
		if got := outOfStock.Len(); got != 1 {
			t.Fatalf("out of stock matches = %d, want 1", got)
		}
		partial := doc.FindAllByText("stock", TextSearchOptions{Partial: true})
		if got := partial.Len(); got != 3 {
			t.Fatalf("partial stock matches = %d, want 3", got)
		}
		caseSensitive := doc.FindAllByText("in stock:", TextSearchOptions{Partial: true, CaseSensitive: true})
		if got := caseSensitive.Len(); got != 0 {
			t.Fatalf("case-sensitive lowercase matches = %d, want 0", got)
		}

		products, ok := doc.CSS(".products").First()
		if !ok {
			t.Fatal("expected products section")
		}
		withinProducts := products.FindAllByText("Details", TextSearchOptions{Partial: true})
		if got := withinProducts.Len(); got != 3 {
			t.Fatalf("product-scoped details matches = %d, want 3", got)
		}
	})

	t.Run("find similar returns same-depth same-shape elements", func(t *testing.T) {
		first, ok := doc.CSS(`article[data-id="p1"]`).First()
		if !ok {
			t.Fatal("expected first product")
		}
		similar := first.FindSimilar(SimilarOptions{})
		if got := similar.Len(); got != 2 {
			t.Fatalf("similar products = %d, want 2", got)
		}
		if got := similar.Text(); !strings.Contains(got, "Camp Mug") || !strings.Contains(got, "Repair Kit") {
			t.Fatalf("similar text = %q, want other products", got)
		}

		onlyKit := first.FindSimilar(SimilarOptions{
			Threshold:        0.7,
			IgnoreAttributes: []string{"data-id", "data-price"},
		})
		if got := onlyKit.Len(); got != 2 {
			t.Fatalf("similar kit products = %d, want 2", got)
		}
		withText := first.FindSimilar(SimilarOptions{Threshold: 0.9, MatchText: true})
		if withText.Len() >= similar.Len() {
			t.Fatalf("text-sensitive similar count = %d, want fewer than %d", withText.Len(), similar.Len())
		}
	})
}
