package templates

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

const maxSitemapGunzipSize = 64 * 1024 * 1024

type SitemapSpider struct {
	Rules          []CrawlRule
	SitemapFollow  *spiders.LinkExtractor
	AlternateLinks bool
}

type SitemapResult struct {
	URLs     []string
	Sitemaps []string
}

func (s *SitemapSpider) StartRequests(urls []string) []spiders.Request {
	requests := make([]spiders.Request, 0, len(urls))
	for _, rawURL := range urls {
		requests = append(requests, spiders.Request{URL: rawURL, Callback: s.ParseSitemap})
	}
	return requests
}

func (s *SitemapSpider) ParseSitemap(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
	if strings.HasSuffix(parsedPath(response.URL()), "/robots.txt") {
		return s.parseRobotsSitemaps(response)
	}
	result, err := s.ParseSitemapBody(response.Body(), response.Headers().Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	outputs := make([]spiders.Output, 0, len(result.Sitemaps)+len(result.URLs))
	for _, childURL := range result.Sitemaps {
		if s.SitemapFollow != nil && !s.SitemapFollow.Matches(childURL) {
			continue
		}
		request, err := response.Follow(childURL, spiders.FollowOptions{Callback: s.ParseSitemap})
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, spiders.Next(request))
	}
	for _, pageURL := range result.URLs {
		request, err := DispatchByRules(response, pageURL, s.Rules, nil)
		if err != nil {
			return nil, err
		}
		if request != nil {
			outputs = append(outputs, spiders.Next(*request))
		}
	}
	return outputs, nil
}

func (s *SitemapSpider) ParseSitemapBody(body []byte, contentType string) (SitemapResult, error) {
	body, err := decompressSitemap(body, contentType)
	if err != nil {
		return SitemapResult{}, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(body))
	var result SitemapResult
	var stack []string
	var inURL bool
	var inSitemap bool
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return SitemapResult{}, err
		}
		switch token := tok.(type) {
		case xml.StartElement:
			name := token.Name.Local
			stack = append(stack, name)
			if name == "url" {
				inURL = true
			}
			if name == "sitemap" {
				inSitemap = true
			}
			if s.AlternateLinks && inURL && name == "link" {
				for _, attr := range token.Attr {
					if attr.Name.Local == "href" && strings.TrimSpace(attr.Value) != "" {
						result.URLs = append(result.URLs, strings.TrimSpace(attr.Value))
					}
				}
			}
		case xml.EndElement:
			name := token.Name.Local
			if name == "url" {
				inURL = false
			}
			if name == "sitemap" {
				inSitemap = false
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 || stack[len(stack)-1] != "loc" {
				continue
			}
			value := strings.TrimSpace(string(token))
			if value == "" {
				continue
			}
			if inSitemap {
				result.Sitemaps = append(result.Sitemaps, value)
			} else if inURL {
				result.URLs = append(result.URLs, value)
			}
		}
	}
	return result, nil
}

func (s *SitemapSpider) parseRobotsSitemaps(response spiders.Response) ([]spiders.Output, error) {
	outputs := make([]spiders.Output, 0)
	for _, line := range strings.Split(response.Text(), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "sitemap") {
			continue
		}
		sitemapURL := strings.TrimSpace(value)
		if sitemapURL == "" {
			continue
		}
		request, err := response.Follow(sitemapURL, spiders.FollowOptions{Callback: s.ParseSitemap})
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, spiders.Next(request))
	}
	return outputs, nil
}

func decompressSitemap(body []byte, contentType string) ([]byte, error) {
	if !strings.Contains(strings.ToLower(contentType), "gzip") && !bytes.HasPrefix(body, []byte{0x1f, 0x8b}) {
		return body, nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	limited := io.LimitReader(reader, maxSitemapGunzipSize+1)
	out, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(out) > maxSitemapGunzipSize {
		return nil, fmt.Errorf("gzip output exceeds %d bytes", maxSitemapGunzipSize)
	}
	return out, nil
}

func parsedPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Path
}
