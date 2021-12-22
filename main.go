package main

import (
	"log"

	"github.com/gonejack/extract-weibo/cmd"
)

func main() {
	var c cmd.ExtractWeibo

	if e := c.Run(); e != nil {
		log.Fatal(e)
	}
}
