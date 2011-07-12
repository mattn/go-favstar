package main

import (
	"fmt"
	"log"
	"os"
	"github.com/mattn/go-favstar"
	"github.com/mattn/go-iconv"
)

func convert_utf8(s string) string {
	ic, err := iconv.Open("char", "UTF-8")
	if err != nil {
		return s
	}
	defer ic.Close()
	ret, _ := ic.Conv(s)
	return ret
}

func main() {
	if len(os.Args) != 2 {
		println("usage: favstar [user_id]")
		os.Exit(-1)
	}
	user := os.Args[1]
	f, err := favstar.Get(user)
	if err != nil {
		log.Fatal("failed to display favstar:", err)
	}
	for _, e := range f.Entry {
		fmt.Println(convert_utf8(e.Text))
		if len(e.Fav) > 0 {
			fmt.Print("FAV: ")
			for _, ef := range e.Fav {
				fmt.Print(ef + " ")
			}
			fmt.Println()
		}
		if len(e.RT) > 0 {
			fmt.Print("RT: ")
			for _, er := range e.RT {
				fmt.Print(er + " ")
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
