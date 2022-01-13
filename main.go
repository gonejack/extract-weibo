package main

import (
	"log"
	"os"

	"github.com/gonejack/extract-weibo/extractweibo"
)

func init() {
	log.SetOutput(os.Stdout)
}
func main() {
	cmd := extractweibo.Weibo{
		Options: extractweibo.MustParseOptions(),
	}
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
