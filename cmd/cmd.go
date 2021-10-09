package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
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
		data, err := os.ReadFile(html)
		if err != nil {
			return err
		}
		jsons, err := c.parseJSON(bytes.NewReader(data))
		if err != nil {
			return err
		}
		rdata, err := c.decodeData(jsons)
		if err != nil {
			return err
		}

		author := strings.TrimSpace(rdata.Status.User.ScreenName)
		title := []rune(strings.TrimSpace(rdata.Status.StatusTitle))
		if len(title) > 30 {
			title = append(title[:30], '.', '.', '.')
		}
		output := fmt.Sprintf("[%s][%s][%s].wb.html", author, rdata.CreateTimeString(), string(title))
		output = strings.ReplaceAll(output, "/", ".")
		output = strings.ReplaceAll(output, ":", ".")

		err = ioutil.WriteFile(output, []byte(rdata.HTML()), 0666)
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *ExtractWeibo) parseJSON(reader io.Reader) (renderData string, err error) {
	doc, err := goquery.NewDocumentFromReader(reader)
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
		renderData = sc.Text()
	}
	err = cmd.Wait()

	return
}
func (c *ExtractWeibo) decodeData(j string) (rd *model.RenderData, err error) {
	rd = new(model.RenderData)
	return rd, rd.From([]byte(j))
}
func (c *ExtractWeibo) operateDoc(doc *goquery.Document, data *model.RenderData) *goquery.Document {
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
		if err == nil && strings.Contains(p.Host, "weibo") {
			p.Host = "m.weibo.cn"
			u = p.String()
		}
		_, _ = fmt.Fprintln(os.Stdout, u)
	}

	return
}
