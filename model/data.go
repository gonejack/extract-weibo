package model

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Status struct {
	StatusTitle string `json:"status_title"`
	Text        string `json:"text"`
	User        User   `json:"user"`
	Bid         string `json:"bid"`
	Pics        []Pic  `json:"pics"`
}

type User struct {
	ScreenName string `json:"screen_name"`
}

type Pic struct {
	Large struct {
		Size string
		URL  string
		Geo  struct {
			Width  string
			Height string
			Croped bool
		}
	}
}

type RenderData struct {
	Status Status `json:"status"`
}

func (r *RenderData) HTML() (s string) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(`<!DOCTYPE html><html lang="zh-cn"></html>`))
	doc.Find("head").AppendHtml(`<meta charset="UTF-8">`)

	title := fmt.Sprintf(`<title>%s</title>`, html.EscapeString(r.Status.StatusTitle))
	doc.Find("head").AppendHtml(title)
	body := doc.Find("body")
	body.AppendHtml(r.info())
	body.AppendHtml(r.Status.Text)
	for _, pic := range r.Status.Pics {
		img := fmt.Sprintf(`<img src="%s">`, pic.Large.URL)
		body.AppendHtml(img)
	}
	body.Find("a").Each(func(i int, a *goquery.Selection) {
		ref, _ := a.Attr("href")
		if ref != "" {
			a.SetAttr("href", r.patchRef(ref))
		}
	})
	body.AppendHtml(r.foot())

	s, _ = doc.Html()

	return
}

func (r *RenderData) Link() string {
	return fmt.Sprintf("https://m.weibo.cn/status/%s", r.Status.Bid)
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
    <a title="Published: 2021-03-30 17:24:49" href="{link}"
       style="display:block; color: #000; padding-bottom: 10px; text-decoration: none; font-size:1em; font-weight: normal;">
        <span style="display: block; color: #666; font-size:1.0em; font-weight: normal;">{source}</span>
        <span style="font-size: 1.5em;">{title}</span>
    </a>
</p>
`
	return strings.NewReplacer(
		"{link}", r.Link(),
		"{source}", html.EscapeString(r.Status.User.ScreenName),
		"{title}", html.EscapeString(r.Status.StatusTitle),
	).Replace(tpl)
}

func (r *RenderData) foot() string {
	const tpl = `
<br/><br/>
<a style="display: block; display: inline-block; border-top: 1px solid #ccc; padding-top: 5px; color: #666; text-decoration: none;"
   href="{link}">{link}</a>
<p style="color:#999;">Save with <a style="color:#666; text-decoration:none; font-weight: bold;"
                                    href="https://github.com/gonejack/inostar">inostar</a>
</p>`

	return strings.NewReplacer(
		"{link}", fmt.Sprintf("https://m.weibo.cn/status/%s", r.Status.Bid),
	).Replace(tpl)
}
