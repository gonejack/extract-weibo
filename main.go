package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/gonejack/extract-weibo/cmd"
)

var (
	verbose bool
	convert bool

	prog = &cobra.Command{
		Use:   "extract-weibo *.html",
		Short: "Command line tool for extracting weibo content from m.weibo.cn html files",
		Run: func(c *cobra.Command, args []string) {
			err := run(c, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)
	prog.Flags().SortFlags = false

	flags := prog.PersistentFlags()
	{
		flags.SortFlags = false
		flags.BoolVarP(&convert, "convert", "c", false, "convert weibo.com links to m.weibo.cn")
		flags.BoolVarP(&verbose, "verbose", "v", false, "verbose")
	}
}
func run(c *cobra.Command, args []string) error {
	if convert {
		convertLink(args)
		return nil
	}

	exec := cmd.ExtractWeibo{
		Verbose: verbose,
	}

	if len(args) == 0 {
		args, _ = filepath.Glob("*.html")
	}

	return exec.Run(args)
}
func convertLink(args []string) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			args = append(args, scan.Text())
		}
	}
	for _, ref := range args {
		u, err := url.Parse(ref)
		if err == nil {
			u.Host = "m.weibo.cn"
			ref = u.String()
		}
		_, _ = fmt.Fprintln(os.Stdout, ref)
	}
}
func main() {
	_ = prog.Execute()
}
