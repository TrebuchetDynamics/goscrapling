package goscrapling

import (
	"context"
	"io"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type ParseOptions struct {
	URL   string
	Store Store
}

type Store interface{}

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
		domain: "default",
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
	return errAdaptiveNotImplemented
}

func (d *Document) Relocate(ctx context.Context, identifier string) (Match, bool, error) {
	return Match{}, false, errAdaptiveNotImplemented
}
