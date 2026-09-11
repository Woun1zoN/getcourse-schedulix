package getcourse

import (
	"strings"

	"golang.org/x/net/html"
)

func (c *Client) GetLessonIDs() ([]string, error) {
	body, err := c.get(c.baseURL + "/teach/control/stream/view/id/935798936")
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	ids := make(map[string]struct{})

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				const prefix = "/teach/control/lesson/view/id/"

				if strings.HasPrefix(attr.Val, prefix) {
					id := strings.TrimPrefix(attr.Val, prefix)

					if i := strings.IndexByte(id, '&'); i != -1 {
						id = id[:i]
					}

					if id != "" {
						ids[id] = struct{}{}
					}
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}

	return result, nil
}