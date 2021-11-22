package rule34xxx

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/darkdragn/gocyberdrop"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type Gallery struct {
	Client gocyberdrop.Client
	URL    string
}

func (g *Gallery) PullGallery() error {
	doc := g.Client.LoadDoc(g.URL)
	u, _ := url.Parse(g.URL)
	title := u.Query()["tags"][0]

	g.Client.Logger.Info(title)
	os.Mkdir(title, 0755)

	var imgs []Image
	var wg sync.WaitGroup
	completion := make(chan bool, 1)
	count := int64(0)

	pid := g.FindLast(*doc)
	for i := 0; i <= pid; i += 42 {
		q := u.Query()
		q.Set("pid", strconv.Itoa(i))
		u.RawQuery = q.Encode()
		doc = g.Client.LoadDoc(u.String())
		selection := doc.Find("span.thumb a")

		selection.Each(func(i int, s *goquery.Selection) {
			url, exists := s.Attr("href")
			if exists {
				img := g.ParsePost(url)
				imgs = append(imgs, img)
				// wg.Add(1)
				// go g.Client.PullImage(img.Url, img.Filename, title, completion, &wg)
			}
		})
	}

	bar := progressbar.Default(int64(len(imgs)))
	for _, img := range imgs {
		wg.Add(1)
		go g.Client.PullImage(img.Url, img.Filename, title, completion, &wg)
	}
	go func() {
		for range completion {
			count += 1
			bar.Add(1)
		}
	}()
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	if count < int64(len(imgs)) {
		fmt.Printf("\n")
		g.Client.Logger.
			WithFields(log.Fields{
				"Total":   len(imgs),
				"Current": count,
			}).
			Error("Unable to download all, please retry")
		os.Exit(1)
	} else {
		bar.Set(int(count))
	}
	return nil
}

type Image struct {
	Url      string
	Filename string
}

func NewImage(input string) Image {
	u, _ := url.Parse(input)
	path := u.Path
	fn := path[strings.LastIndex(path, "/")+1:]
	return Image{input, fn}
}

func (g *Gallery) FindLast(doc goquery.Document) int {
	sel := doc.Find(".pagination a").Last()
	page, exists := sel.Attr("href")
	if exists {
		pageParsed, _ := url.ParseQuery(page)
		pid, err := strconv.Atoi(pageParsed["pid"][0])
		g.Client.Catch(err)
		return pid
	}
	return 0
}
func (g *Gallery) ParsePost(post string) Image {
	base := "https://rule34.xxx/"
	doc := g.Client.LoadDoc(base + post)
	selection := doc.Find("meta[property='og:image']").First()
	output, _ := selection.Attr("content")
	return NewImage(output)
}
