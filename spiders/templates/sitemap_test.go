package templates

import (
	"reflect"
	"testing"
)

func TestParseSitemapBodyIgnoresNestedExtensionLocs(t *testing.T) {
	sitemap := &SitemapSpider{}
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>https://example.com/page</loc>
    <image:image>
      <image:loc>https://cdn.example.com/image.jpg</image:loc>
    </image:image>
  </url>
</urlset>`)

	result, err := sitemap.ParseSitemapBody(body, "application/xml")
	if err != nil {
		t.Fatalf("ParseSitemapBody returned error: %v", err)
	}
	want := []string{"https://example.com/page"}
	if !reflect.DeepEqual(result.URLs, want) {
		t.Fatalf("URLs = %#v, want only top-level url locs %#v", result.URLs, want)
	}
	if len(result.Sitemaps) != 0 {
		t.Fatalf("Sitemaps = %#v, want none", result.Sitemaps)
	}
}
