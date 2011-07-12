package main

import (
	"fmt"
	"html"
	"http"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"github.com/mattn/go-iconv/iconv"
)

type cond map[string]string

func walk(n *html.Node, tag string, attr cond) (l []*html.Node) {
	switch n.Type {
	case html.ErrorNode:
		return
	case html.DocumentNode:
		if len(n.Child) > 0 {
			l = walk(n.Child[0], tag, attr)
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

func main() {
	if len(os.Args) != 2 {
		println("usage: favstar [user_id]")
		os.Exit(-1)
	}
	user := os.Args[1]
	res, err := http.Get("http://favstar.fm/users/" + user + "/recent")
	if err != nil {
		log.Fatal("failed to display favstar:", err)
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		log.Fatal("failed to parse html:", err)
	}
	tweetWithStats := walk(doc, "div", cond{"class": "tweetWithStats"})
	for _, tweetWithStat := range tweetWithStats {
		theTweet := walk(tweetWithStat, "div", cond{"class": "theTweet"})
		if len(theTweet) == 0 {
			continue
		}
		fmt.Println(convert_utf8(text(theTweet[0])))
		avatarLists := walk(tweetWithStat, "div", cond{"class": "avatarList"})
		for _, avatarList := range avatarLists {
			id := attr(avatarList, "id")
			avatars := walk(avatarList, "a", cond{"class": "avatar"})
			if isFav(id) {
				fmt.Print("FAV: ")
				for _, avatar := range avatars {
					fmt.Print(attr(avatar, "title"), " ")
				}
				fmt.Println()
			}
			if isRt(id) {
				fmt.Print("RT: ")
				for _, avatar := range avatars {
					fmt.Print(attr(avatar, "title"), " ")
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}
}
