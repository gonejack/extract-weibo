package model

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Status struct {
	StatusTitle string    `json:"status_title"`
	CreatedAt   string    `json:"created_at"`
	Text        string    `json:"text"`
	User        User      `json:"user"`
	Bid         string    `json:"bid"`
	Pics        []Picture `json:"pics"`
}
type User struct {
	ScreenName string `json:"screen_name"`
}
type Picture struct {
	Large struct {
		Size string `json:"size"`
		URL  string `json:"url"`
	} `json:"large"`
}
type Weibo struct {
	Status Status `json:"status"`
}

func (wb *Weibo) From(jsons []byte) (err error) {
	return json.Unmarshal(jsons, wb)
}
func (wb *Weibo) HTML() (content string) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<!DOCTYPE html><html lang="zh-cn"></html>`))
	head := doc.Find("head")
	head.AppendHtml(`<meta charset="UTF-8">`)
	head.AppendHtml(fmt.Sprintf(`<meta name="inostar:publish" content="%s">`, wb.CreateTime().Format(time.RFC1123Z)))
	head.AppendHtml(fmt.Sprintf(`<title>%s</title>`, html.EscapeString(wb.Status.StatusTitle)))

	body := doc.Find("body")
	body.AppendHtml(wb.header())
	body.AppendHtml(wb.Status.Text)
	body.AppendHtml("<br />")
	body.Find("a").Each(func(i int, a *goquery.Selection) {
		ref, _ := a.Attr("href")
		if ref != "" {
			a.SetAttr("href", wb.patchRef(ref))
		}
	})
	for _, p := range wb.Status.Pics {
		body.AppendHtml(fmt.Sprintf(`<a href="%s"><img src="%s" alt="%s"></a><br />`, p.Large.URL, p.Large.URL, p.Large.URL))
	}
	body.AppendHtml(wb.footer())

	content, _ = doc.Html()

	return
}
func (wb *Weibo) CreateTime() time.Time {
	t, err := time.Parse(time.RubyDate, wb.Status.CreatedAt)
	if err != nil {
		t = time.Now()
	}
	return t
}
func (wb *Weibo) CreateTimeString() string {
	return wb.CreateTime().Format("2006-01-02 15:04:05")
}
func (wb *Weibo) Link() string {
	return fmt.Sprintf("https://m.weibo.cn/status/%s", wb.Status.Bid)
}

func (wb *Weibo) header() string {
	const tpl = `
<p>
    <a title="Published: {time}" href="{link}"
       style="display:block; color: #000; padding-bottom: 10px; text-decoration: none; font-size:1em; font-weight: normal;">
        <span style="display: block; color: #666; font-size:1.0em; font-weight: normal;">{source}</span>
        <span style="font-size: 1.5em;">{title}</span>
    </a>
</p>
`
	return strings.NewReplacer(
		"{time}", wb.CreateTimeString(),
		"{link}", wb.Link(),
		"{source}", html.EscapeString(wb.Status.User.ScreenName),
		"{title}", html.EscapeString(wb.Status.StatusTitle),
	).Replace(tpl)
}
func (wb *Weibo) footer() string {
	const tpl = `
<br/><br/>
<a style="display: inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="{link}">{link}</a>
<p style="color:#999;">Save with <a style="color:#666; text-decoration:none; font-weight: bold;"
                                    href="https://github.com/gonejack/extract-weibo">extract-weibo</a>
</p>`

	return strings.NewReplacer(
		"{link}", fmt.Sprintf("https://m.weibo.cn/status/%s", wb.Status.Bid),
	).Replace(tpl)
}
func (wb *Weibo) patchRef(ref string) string {
	h, err := url.Parse(wb.Link())
	if err != nil {
		return ref
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if u.Scheme == "" {
		u.Scheme = h.Scheme
	}
	if u.Host == "" {
		u.Host = h.Host
	}
	return u.String()
}
