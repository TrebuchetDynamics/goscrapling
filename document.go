package goscrapling

import (
	"context"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type ParseOptions struct {
	URL   string
	Store Store
}

type Document struct {
	root   *html.Node
	query  *goquery.Document
	url    string
	domain string
	store  Store
}

func Parse(r io.Reader, opts ParseOptions) (*Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return &Document{
		root:   root,
		query:  goquery.NewDocumentFromNode(root),
		url:    opts.URL,
		domain: adaptiveDomain(opts.URL),
		store:  opts.Store,
	}, nil
}

func (d *Document) CSS(selector string) Selection {
	if d == nil || d.query == nil {
		return Selection{}
	}

	var elements []*Element
	d.query.Find(selector).Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			elements = append(elements, &Element{doc: d, node: node})
		}
	})

	return Selection{elements: elements}
}

func (d *Document) Save(ctx context.Context, element *Element, identifier string) error {
	if d == nil || d.store == nil {
		return ErrMissingStore
	}
	if element == nil || element.node == nil {
		return ErrNilElement
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return ErrEmptyIdentifier
	}

	return d.store.Save(ctx, Key{Domain: d.domain, Identifier: identifier}, fingerprintNode(element.node))
}

func (d *Document) Relocate(ctx context.Context, identifier string) (Match, bool, error) {
	if d == nil || d.store == nil {
		return Match{}, false, ErrMissingStore
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return Match{}, false, ErrEmptyIdentifier
	}

	target, ok, err := d.store.Load(ctx, Key{Domain: d.domain, Identifier: identifier})
	if err != nil || !ok {
		return Match{}, false, err
	}

	var best Match
	found := false
	d.query.Find("*").Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			candidate := &Element{doc: d, node: node}
			score := scoreFingerprint(fingerprintNode(node), target)
			if score < defaultMinScore {
				continue
			}
			if !found || score > best.Score {
				best = Match{Element: candidate, Score: score}
				found = true
			}
		}
	})

	return best, found, nil
}
