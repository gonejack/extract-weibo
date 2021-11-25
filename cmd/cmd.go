package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/kong"

	"github.com/gonejack/extract-weibo/model"
)

type options struct {
	Convert bool `short:"c" help:"Convert weibo.com links to m.weibo.cn."`
	Verbose bool `short:"v" help:"Verbose printing."`

	Args []string `arg:"" optional:""`
}

type ExtractWeibo struct {
	options
}

func (c *ExtractWeibo) Run() error {
	kong.Parse(&c.options,
		kong.Name("extract-weibo"),
		kong.Description("Command line tool for extracting weibo content from m.weibo.cn html files"),
		kong.UsageOnError(),
	)

	if c.Convert {
		return c.convertLink()
	} else {
		if len(c.Args) == 0 {
			c.Args, _ = filepath.Glob("*.html")
		}
		if len(c.Args) == 0 {
			return errors.New("no HTML files given")
		}
		return c.run()
	}
}
func (c *ExtractWeibo) run() error {
	for _, html := range c.Args {
		if c.Verbose {
			log.Printf("processing %s", html)
		}
		json, err := c.parseHTML(html)
		if err != nil {
			return err
		}
		weibo, err := c.decodeWeibo(json)
		if err != nil {
			return err
		}

		author := strings.TrimSpace(weibo.Status.User.ScreenName)
		title := []rune(strings.TrimSpace(weibo.Status.StatusTitle))
		if len(title) > 30 {
			title = append(title[:30], '.', '.', '.')
		}
		output := fmt.Sprintf("[%s][%s][%s].wb.html", author, weibo.CreateTimeString(), string(title))
		output = strings.ReplaceAll(output, "/", ".")
		output = strings.ReplaceAll(output, ":", ".")

		err = os.WriteFile(output, []byte(weibo.HTML()), 0666)
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *ExtractWeibo) parseHTML(html string) (json string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse %s error: %w", html, err)
		}
	}()

	fd, err := os.Open(html)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(fd)
	if err != nil {
		return
	}

	script := doc.Find("body").Find("script").First().Text()
	script = fmt.Sprintf("%s\n console.log(JSON.stringify($render_data))", script)

	cmd := exec.Command("node", "-e", script)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		json = sc.Text()
	}
	err = cmd.Wait()

	return
}
func (c *ExtractWeibo) decodeWeibo(j string) (data *model.Weibo, err error) {
	data = new(model.Weibo)
	return data, data.From([]byte(j))
}
func (c *ExtractWeibo) operateDoc(doc *goquery.Document, data *model.Weibo) *goquery.Document {
	doc.Find("div.wrap").Remove()
	doc.Find("div.weibo-media-wraps").Parent().Remove()
	for _, pic := range data.Status.Pics {
		doc.Find("div.weibo-og").AppendHtml(fmt.Sprintf(`<img src="%s">`, pic.Large.URL))
	}
	return doc
}
func (c *ExtractWeibo) convertLink() (err error) {
	var urls []string

	// scan urls from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			urls = append(urls, scan.Text())
		}
	}

	urls = append(urls, c.Args...)

	for _, u := range urls {
		p, err := url.Parse(u)
		if err != nil {
			continue
		}
		if !strings.Contains(p.Host, "weibo") {
			continue
		}

		switch {
		case p.Host == "share.api.weibo.cn": // weibo international
			wid := p.Query().Get("weibo_id")
			wid = regexp.MustCompile(`^\d+`).FindString(wid)
			p.Path = path.Join("status", wid)
			p.RawQuery = ""
		default:
			// weibo china
		}
		p.Host = "m.weibo.cn"

		_, _ = fmt.Fprintln(os.Stdout, p.String())
	}

	return
}
