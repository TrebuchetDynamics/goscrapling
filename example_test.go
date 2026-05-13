package goscrapling_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
)

func ExampleDocument_Relocate() {
	ctx := context.Background()
	store := goscrapling.NewMemoryStore()

	before, _ := goscrapling.Parse(strings.NewReader(`<article class="product" id="p1">Product 1</article>`), goscrapling.ParseOptions{
		URL:   "https://example.com/products",
		Store: store,
	})
	element, _ := before.CSS("#p1").First()
	_ = before.Save(ctx, element, "featured-product")

	after, _ := goscrapling.Parse(strings.NewReader(`<article class="product" data-id="p1"><span>Product 1</span></article>`), goscrapling.ParseOptions{
		URL:   "https://example.com/products",
		Store: store,
	})
	match, ok, _ := after.Relocate(ctx, "featured-product")
	fmt.Println(ok, match.Element.Text())

	// Output:
	// true Product 1
}
