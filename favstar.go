package favstar

import (
	"github.com/mattn/go-iconv"
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
		for _, c := range n.Child {
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
			if len(attr) > 0 {
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
		for _, c := range n.Child {
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

func convert_utf8(s string) string {
	ic, err := iconv.Open("char", "UTF-8")
	if err != nil {
		return s
	}
	defer ic.Close()
	ret, _ := ic.Conv(s)
	return ret
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
	tweetWithStats := walk(doc, "div", cond{"class": "tweetWithStats"})
	for _, tweetWithStat := range tweetWithStats {
		theTweet := walk(tweetWithStat, "div", cond{"class": "theTweet"})
		if len(theTweet) == 0 {
			continue
		}
		var e Entry
		e.Text = text(theTweet[0])
		avatarLists := walk(tweetWithStat, "div", cond{"class": "avatarList"})
		for _, avatarList := range avatarLists {
			id := attr(avatarList, "id")
			avatars := walk(avatarList, "a", cond{"class": "avatar"})
			if isFav(id) {
				for _, avatar := range avatars {
					e.Fav = append(e.Fav, attr(avatar, "title"))
				}
			}
			if isRt(id) {
				for _, avatar := range avatars {
					e.RT = append(e.RT, attr(avatar, "title"))
				}
			}
		}
		f.Entry = append(f.Entry, e)
	}
	return
}
