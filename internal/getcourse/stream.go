package getcourse

import (
	"strings"
	"strconv"

	"golang.org/x/net/html"
)

type Stream struct {
	ID      int64
	Name    string
	Lessons string
	Teacher string
}

func (s Stream) Info() string {
	parts := make([]string, 0, 2)
	if s.Lessons != "" {
		parts = append(parts, s.Lessons)
	}
	if s.Teacher != "" {
		parts = append(parts, s.Teacher)
	}
	return strings.Join(parts, " · ")
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

	byID := make(map[int64]*Stream)
	var order []int64

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if id, ok := streamIDFromLink(n); ok {
				s, seen := byID[id]
				if !seen {
					s = &Stream{ID: id}
					byID[id] = s
					order = append(order, id)
				}
				fillStream(s, n)
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)

	streams := make([]Stream, 0, len(order))
	for _, id := range order {
		s := *byID[id]
		if s.Name == "" {
			s.Name = "Тренинг " + strconv.FormatInt(id, 10)
		}
		streams = append(streams, s)
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

func streamIDFromLink(a *html.Node) (int64, bool) {
	const marker = "/teach/control/stream/view/id/"

	for _, attr := range a.Attr {
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
		if id, err := strconv.ParseInt(rest[:end], 10, 64); err == nil {
			return id, true
		}
	}
	return 0, false
}

func fillStream(s *Stream, a *html.Node) {
	if s.Name == "" {
		title := findNode(a, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "span" && hasClass(n, "stream-title")
		})
		if title != nil {
			s.Name = nodeText(title)
		}
	}

	if s.Lessons == "" {
		if b := findNode(a, isTag("b")); b != nil {
			s.Lessons = trimDot(nodeText(b))
		}
	}

	if s.Teacher == "" {
		if d := findNode(a, isTag("div")); d != nil {
			var sb strings.Builder
			for ch := d.FirstChild; ch != nil; ch = ch.NextSibling {
				if ch.Type == html.TextNode {
					sb.WriteString(ch.Data)
					sb.WriteByte(' ')
				}
			}
			s.Teacher = trimDot(strings.Join(strings.Fields(sb.String()), " "))
		}
	}
}

func trimDot(s string) string {
	return strings.TrimSuffix(strings.TrimSpace(s), ".")
}

func isTag(name string) func(*html.Node) bool {
	return func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == name }
}

func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if r := findNode(ch, match); r != nil {
			return r
		}
	}
	return nil
}