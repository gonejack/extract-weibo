package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"

	"github.com/gonejack/extract-weibo/model"
)

type ExtractWeibo struct {
	Verbose bool
}

func (w *ExtractWeibo) Run(htmlList []string) (err error) {
	if len(htmlList) == 0 {
		return errors.New("no HTML files given")
	}

	for _, html := range htmlList {
		err = w.process(html)
		if err != nil {
			return fmt.Errorf("parse %s failed: %s", html, err)
		}
	}

	return
}
func (w *ExtractWeibo) process(url string) (err error) {
	if w.Verbose {
		log.Printf("processing %s", url)
	}

	rdata, err := w.getData2(url)
	if err != nil {
		return err
	}
	data, err := w.decodeData(rdata)
	if err != nil {
		return
	}

	htm := data.HTML()

	out := fmt.Sprintf("%s.html", strings.ReplaceAll(data.Status.StatusTitle, "/", "_"))
	return ioutil.WriteFile(out, []byte(htm), 0766)
}
func (w *ExtractWeibo) getData(url string) (renderData string, err error) {
	timeout, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	)
	ctx, cancel := chromedp.NewExecAllocator(timeout, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
		chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second/5),
		chromedp.EvaluateAsDevTools("JSON.stringify($render_data)", &renderData),
	)

	return
}
func (w *ExtractWeibo) getData2(url string) (renderData string, err error) {
	timeout, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(timeout, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
func (w *ExtractWeibo) decodeData(j string) (data *model.RenderData, err error) {
	data = new(model.RenderData)
	return data, json.Unmarshal([]byte(j), data)
}
func (w *ExtractWeibo) openLocalFile(htmlFile string, ref string) (fd *os.File, err error) {
	fd, err = os.Open(ref)
	if err == nil {
		return
	}

	// compatible with evernote's exported htmls
	{
		basename := strings.TrimSuffix(htmlFile, filepath.Ext(htmlFile))
		filename := filepath.Base(ref)
		fd, err = os.Open(filepath.Join(basename+"_files", filename))
		if err == nil {
			return
		}
		fd, err = os.Open(filepath.Join(basename+".resources", filename))
		if err == nil {
			return
		}
		if strings.HasSuffix(ref, ".") {
			return w.openLocalFile(htmlFile, strings.TrimSuffix(ref, "."))
		}
	}

	return
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
