package goscrapling

import (
	"strings"

	"golang.org/x/net/html"
)

func (e *Element) HasClass(className string) bool {
	if e == nil || e.node == nil || className == "" {
		return false
	}
	classAttr, ok := e.Attr("class")
	if !ok {
		return false
	}
	for _, class := range strings.Fields(classAttr) {
		if class == className {
			return true
		}
	}
	return false
}

func (e *Element) Parent() (*Element, bool) {
	if e == nil || e.node == nil {
		return nil, false
	}
	for parent := e.node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			return &Element{doc: e.doc, node: parent}, true
		}
	}
	return nil, false
}

func (e *Element) Children() Selection {
	if e == nil || e.node == nil {
		return Selection{}
	}
	children := make([]*Element, 0)
	for child := e.node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			children = append(children, &Element{doc: e.doc, node: child})
		}
	}
	return Selection{elements: children}
}

func (e *Element) Siblings() Selection {
	parent, ok := e.Parent()
	if !ok {
		return Selection{}
	}
	return parent.Children().Filter(func(candidate *Element) bool {
		return candidate != nil && candidate.node != e.node
	})
}

func (e *Element) Next() (*Element, bool) {
	if e == nil || e.node == nil {
		return nil, false
	}
	for sibling := e.node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return &Element{doc: e.doc, node: sibling}, true
		}
	}
	return nil, false
}

func (e *Element) Previous() (*Element, bool) {
	if e == nil || e.node == nil {
		return nil, false
	}
	for sibling := e.node.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
		if sibling.Type == html.ElementNode {
			return &Element{doc: e.doc, node: sibling}, true
		}
	}
	return nil, false
}

func (e *Element) Ancestors() Selection {
	if e == nil || e.node == nil {
		return Selection{}
	}
	ancestors := make([]*Element, 0)
	for parent := e.node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			ancestors = append(ancestors, &Element{doc: e.doc, node: parent})
		}
	}
	return Selection{elements: ancestors}
}

func (e *Element) FindAncestor(predicate func(*Element) bool) (*Element, bool) {
	if predicate == nil {
		return nil, false
	}
	for _, ancestor := range e.Ancestors().elements {
		if predicate(ancestor) {
			return ancestor, true
		}
	}
	return nil, false
}

func allDescendantElements(doc *Document, root *html.Node) []*Element {
	if root == nil {
		return nil
	}
	elements := make([]*Element, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode {
			elements = append(elements, &Element{doc: doc, node: node})
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return elements
}

func elementDepth(node *html.Node) int {
	depth := 0
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			depth++
		}
	}
	return depth
}

func parentElementNode(node *html.Node) *html.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			return parent
		}
	}
	return nil
}
