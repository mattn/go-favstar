package favstar

import (
	"exp/html"
	"io/ioutil"
	"net/http"
	"strings"
)

type cond map[string]string

func walk(n *html.Node, tag string, attr cond) (l []*html.Node) {
	switch n.Type {
	case html.ErrorNode:
		return
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			l = walk(c, tag, attr)
		}
		return
	case html.CommentNode:
		return
	case html.TextNode:
		if tag == "TEXT" {
			l = append(l, n)
		} else {
			return
		}
	default:
		return
	case html.ElementNode:
		if strings.ToLower(tag) == strings.ToLower(n.Data) {
			for _, e := range l {
				if e == n {
					return
				}
			}
			if attr != nil && len(attr) > 0 {
				for _, a := range n.Attr {
					val, found := attr[a.Key]
					if found {
						for _, as := range strings.Split(a.Val, " ") {
							if as == val {
								l = append(l, n)
								break
							}
						}
						break
					}
				}
			} else {
				l = append(l, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			for _, f := range walk(c, tag, attr) {
				for _, e := range l {
					if e == f {
						return
					}
				}
				l = append(l, f)
			}
		}
	}

	return
}

func text(node *html.Node) (text string) {
	for _, t := range walk(node, "TEXT", cond{}) {
		text += strings.TrimSpace(t.Data)
	}
	return
}

func attr(node *html.Node, name string) (attr string) {
	for _, a := range node.Attr {
		if a.Key == name {
			attr = a.Val
		}
	}
	return
}

func isFav(id string) bool {
	return strings.Index(id, "faved_by_others_") == 0
}

func isRt(id string) bool {
	return strings.Index(id, "rt_by_others_") == 0
}

type Entry struct {
	Text string
	Fav  []string
	RT   []string
}

type Favstar struct {
	Entry []Entry
}

func Get(user_id string) (f Favstar, err error) {
	res, err := http.Get("http://favstar.fm/users/" + user_id + "/recent")
	if err != nil {
		return
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return
	}
	tweetWithStats := walk(doc, "div", cond{"class": "fs-tweet"})
	for _, tweetWithStat := range tweetWithStats {
		t := walk(tweetWithStat, "p", cond{"class": "fs-tweet-text"})
		if t == nil {
			continue
		}
		var e Entry
		e.Text = t[0].FirstChild.Data

		favs := walk(tweetWithStat, "div", cond{"data-type": "favs"})
		if favs != nil {
			for _, aa := range walk(favs[0], "a", nil) {
				e.Fav = append(e.Fav, attr(aa, "title"))
			}
		}
		rts := walk(tweetWithStat, "div", cond{"data-type": "rts"})
		if rts != nil {
			for _, aa := range walk(rts[0], "a", nil) {
				e.RT = append(e.RT, attr(aa, "title"))
			}
		}
		f.Entry = append(f.Entry, e)
	}
	return
}
