package extractweibo

import (
	"path/filepath"

	"github.com/alecthomas/kong"
)

type Options struct {
	Convert bool `short:"c" help:"Convert weibo.com links to m.weibo.cn."`
	Verbose bool `short:"v" help:"Verbose printing."`
	About   bool `help:"About."`

	HTML []string `name:".html" arg:"" optional:"" help:"list of .html files"`
}

func MustParseOptions() (opt Options) {
	kong.Parse(&opt,
		kong.Name("extract-weibo"),
		kong.Description("Command line tool for extracting weibo content from m.weibo.cn html files"),
		kong.UsageOnError(),
	)
	if len(opt.HTML) == 0 || opt.HTML[0] == "*.html" {
		opt.HTML, _ = filepath.Glob("*.html")
	}
	return
}
