package getcourse

import (
	"strings"

	"golang.org/x/net/html"
)

type Document struct {
	Name string
	URL  string
}

func (c *Client) GetLessonDocuments(id string) ([]Document, error) {
	url := c.baseURL + "/pl/teach/control/lesson/view?id=" + id + "&editMode=0"

	body, err := c.get(url)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var documents []Document

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href string
			var name string

			for _, attr := range n.Attr {
				switch attr.Key {
				case "href":
					href = attr.Val
				}
			}

			if strings.Contains(href, "/fileservice/user/file/download/") {
				name = strings.TrimSpace(textContent(n))

				if strings.HasPrefix(href, "/") {
					href = c.baseURL + href
				}

				documents = append(documents, Document{
					Name: name,
					URL:  href,
				})
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	return documents, nil
}

func textContent(n *html.Node) string {
	var result strings.Builder

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			result.WriteString(n.Data)
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(n)

	return result.String()
}