package main

import (
	"log"

	"github.com/gonejack/extract-weibo/cmd"
)

func main() {
	err := new(cmd.ExtractWeibo).Run()
	if err != nil {
		log.Fatal(err)
	}
}
