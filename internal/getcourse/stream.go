package getcourse

import (
	"strings"
	"strconv"

	"golang.org/x/net/html"
)

type Stream struct {
	ID   int64
	Name string
}

func (c *Client) GetLessonIDs(streamID int64) ([]string, error) {
	body, err := c.get(c.baseURL + "/teach/control/stream/view/id/" + strconv.FormatInt(streamID, 10))
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

func (c *Client) GetStreams() ([]Stream, error) {
	body, err := c.get(c.baseURL + "/teach/control")
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	const marker = "/teach/control/stream/view/id/"

	names := make(map[int64]string)
	var order []int64

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}

				i := strings.Index(attr.Val, marker)
				if i == -1 {
					continue
				}

				rest := attr.Val[i+len(marker):]

				end := 0
				for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
					end++
				}

				id, err := strconv.ParseInt(rest[:end], 10, 64)
				if err != nil {
					continue
				}

				if _, seen := names[id]; !seen {
					order = append(order, id)
					names[id] = ""
				}

				if text := nodeText(n); names[id] == "" && text != "" {
					names[id] = text
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	streams := make([]Stream, 0, len(order))
	for _, id := range order {
		name := names[id]

		if name == "" {
			name = "Поток " + strconv.FormatInt(id, 10)
		}

		streams = append(streams, Stream{ID: id, Name: name})
	}

	return streams, nil
}

func nodeText(n *html.Node) string {
	var sb strings.Builder
	var rec func(*html.Node)

	rec = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
			sb.WriteByte(' ')
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			rec(child)
		}
	}

	rec(n)

	return strings.Join(strings.Fields(sb.String()), " ")
}