package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/gonejack/extract-weibo/model"
)

type ExtractWeibo struct {
	Verbose bool
}

func (w *ExtractWeibo) Run(htmls []string) (err error) {
	if len(htmls) == 0 {
		return errors.New("no HTML files given")
	}

	for _, html := range htmls {
		err = w.process(html)
		if err != nil {
			return fmt.Errorf("parse %s failed: %s", html, err)
		}
	}

	return
}
func (w *ExtractWeibo) process(html string) (err error) {
	if w.Verbose {
		log.Printf("processing %s", html)
	}

	fd, err := os.Open(html)
	if err != nil {
		return
	}
	defer fd.Close()

	jsons, err := w.parseJSON(fd)
	if err != nil {
		return err
	}
	rdata, err := w.decodeData(jsons)
	if err != nil {
		return
	}
	htm := rdata.HTML()

	out := fmt.Sprintf("[%s][%s][%s].wb.html", strings.TrimSpace(rdata.Status.User.ScreenName), strings.TrimSpace(rdata.Status.StatusTitle), rdata.CreateTimeString())
	out = strings.ReplaceAll(out, "/", "_")
	return ioutil.WriteFile(out, []byte(htm), 0666)
}
func (w *ExtractWeibo) parseJSON(reader io.Reader) (renderData string, err error) {
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
func (w *ExtractWeibo) decodeData(j string) (rd *model.RenderData, err error) {
	rd = new(model.RenderData)
	return rd, rd.From([]byte(j))
}
func (w *ExtractWeibo) operateDoc(doc *goquery.Document, data *model.RenderData) *goquery.Document {
	doc.Find("div.wrap").Remove()
	doc.Find("div.weibo-media-wraps").Parent().Remove()

	for _, pic := range data.Status.Pics {
		img := fmt.Sprintf(`<img src="%s">`, pic.Large.URL)
		doc.Find("div.weibo-og").AppendHtml(img)
	}

	return doc
}
