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
		Geo  struct {
			Width  string `json:"width"`
			Height string `json:"height"`
			Croped bool   `json:"croped"`
		}
	} `json:"large"`
}
type RenderData struct {
	Status Status `json:"status"`
}

func (r *RenderData) HTML() (s string) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<!DOCTYPE html><html lang="zh-cn"></html>`))
	head := doc.Find("head")
	{
		head.AppendHtml(`<meta charset="UTF-8">`)
		head.AppendHtml(fmt.Sprintf(`<meta name="inostar:publish" content="%s">`, r.CreateTime().Format(time.RFC1123Z)))
		head.AppendHtml(fmt.Sprintf(`<title>%s</title>`, html.EscapeString(r.Status.StatusTitle)))
	}

	body := doc.Find("body")
	{
		body.AppendHtml(r.info())
		body.AppendHtml(r.Status.Text)
		body.AppendHtml("<br />")
		body.Find("a").Each(func(i int, a *goquery.Selection) {
			ref, _ := a.Attr("href")
			if ref != "" {
				a.SetAttr("href", r.patchRef(ref))
			}
		})
		for _, p := range r.Status.Pics {
			body.AppendHtml(fmt.Sprintf(`<a href="%s"><img src="%s" alt="%s"></a><br />`, p.Large.URL, p.Large.URL, p.Large.URL))
		}
		body.AppendHtml(r.foot())
	}

	s, _ = doc.Html()

	return
}
func (r *RenderData) CreateTime() time.Time {
	t, err := time.Parse(time.RubyDate, r.Status.CreatedAt)
	if err != nil {
		t = time.Now()
	}
	return t
}
func (r *RenderData) CreateTimeString() string {
	return r.CreateTime().Format("2006-01-02 15:04:05")
}
func (r *RenderData) Link() string {
	return fmt.Sprintf("https://m.weibo.cn/status/%s", r.Status.Bid)
}
func (r *RenderData) From(data []byte) (err error) {
	return json.Unmarshal(data, r)
}
func (r *RenderData) patchRef(ref string) string {
	h, err := url.Parse(r.Link())
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
func (r *RenderData) info() string {
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
		"{time}", r.CreateTimeString(),
		"{link}", r.Link(),
		"{source}", html.EscapeString(r.Status.User.ScreenName),
		"{title}", html.EscapeString(r.Status.StatusTitle),
	).Replace(tpl)
}
func (r *RenderData) foot() string {
	const tpl = `
<br/><br/>
<a style="display: inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="{link}">{link}</a>
<p style="color:#999;">Save with <a style="color:#666; text-decoration:none; font-weight: bold;"
                                    href="https://github.com/gonejack/extract-weibo">extract-weibo</a>
</p>`

	return strings.NewReplacer(
		"{link}", fmt.Sprintf("https://m.weibo.cn/status/%s", r.Status.Bid),
	).Replace(tpl)
}
