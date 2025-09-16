package parser

import (
	"bytes"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// ExtractLinks извлекает все абсолютные ссылки из HTML
func ExtractLinks(base *url.URL, data []byte) ([]*url.URL, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var links []*url.URL
	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var attrName string
			switch n.Data {
			case "a":
				attrName = "href"
			case "img", "script":
				attrName = "src"
			case "link":
				for _, a := range n.Attr {
					if a.Key == "rel" && a.Val == "stylesheet" {
						attrName = "href"
					}
				}
			}

			if attrName != "" {
				for _, a := range n.Attr {
					if a.Key == attrName {
						abs, err := base.Parse(a.Val)
						if err == nil && (abs.Scheme == "http" || abs.Scheme == "https") {
							links = append(links, abs)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(doc)

	return links, nil
}

// RewriteLinks переписывает все ссылки в HTML через mapper
func RewriteLinks(base *url.URL, data []byte, mapper func(*url.URL) string) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var dfs func(*html.Node)
	dfs = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var attrName string
			switch n.Data {
			case "a":
				attrName = "href"
			case "img", "script":
				attrName = "src"
			case "link":
				for _, a := range n.Attr {
					if a.Key == "rel" && a.Val == "stylesheet" {
						attrName = "href"
					}
				}
			}

			if attrName != "" {
				for i, a := range n.Attr {
					if a.Key == attrName {
						abs, err := base.Parse(a.Val)
						if err == nil && (abs.Scheme == "http" || abs.Scheme == "https") {
							local := mapper(abs)
							// Обрезаем префикс до host (без корня вывода), чтобы ссылка стала относительной к зеркалу
							hostPrefix := base.Hostname() + "/"
							if idx := strings.Index(local, hostPrefix); idx >= 0 {
								local = local[idx+len(hostPrefix):]
							}
							local = path.Clean(local)
							n.Attr[i].Val = local
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}
	dfs(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
