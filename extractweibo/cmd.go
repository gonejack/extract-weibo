package extractweibo

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/gonejack/extract-weibo/model"
)

type Weibo struct {
	Options
}

func (c *Weibo) Run() (err error) {
	if c.About {
		fmt.Println("Visit https://github.com/gonejack/extract-weibo")
		return
	}
	if c.Convert {
		return c.convertLink()
	}
	if len(c.HTML) == 0 {
		return errors.New("no HTML files given")
	}
	return c.run()
}
func (c *Weibo) run() error {
	for _, html := range c.HTML {
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
func (c *Weibo) parseHTML(html string) (json []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parse %s error: %w", html, err)
		}
	}()

	fd, err := os.Open(html)
	if err != nil {
		return
	}
	defer fd.Close()

	doc, err := goquery.NewDocumentFromReader(fd)
	if err != nil {
		return
	}

	script := doc.Find(`body>script:contains("$render_data")`).First().Text()
	script = fmt.Sprintf("%s\n console.log(JSON.stringify($render_data))", script)

	return exec.Command("node", "-e", script).Output()
}
func (c *Weibo) decodeWeibo(dat []byte) (data *model.Weibo, err error) {
	data = new(model.Weibo)
	return data, data.From(dat)
}
func (c *Weibo) operateDoc(doc *goquery.Document, data *model.Weibo) *goquery.Document {
	doc.Find("div.wrap").Remove()
	doc.Find("div.weibo-media-wraps").Parent().Remove()
	for _, pic := range data.Status.Pics {
		doc.Find("div.weibo-og").AppendHtml(fmt.Sprintf(`<img src="%s">`, pic.Large.URL))
	}
	return doc
}
func (c *Weibo) convertLink() (err error) {
	var urls []string

	// scan urls from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			urls = append(urls, scan.Text())
		}
	}
	urls = append(urls, c.HTML...)

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
