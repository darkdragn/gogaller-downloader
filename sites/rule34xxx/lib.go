package rule34xxx

import (
	"net/url"
	"strconv"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/darkdragn/gogallery-downloader/common"
)

type R34xGallery struct {
	common.GalleryBase
	Tag string
}

func (g *R34xGallery) Title() string {
	return g.Tag
}

func (g *R34xGallery) ImageList() (imgs []common.Image) {
	u, _ := url.Parse("https://rule34.xxx/index.php?page=post&s=list")
	q := u.Query()
	q.Set("tags", g.Tag)
	u.RawQuery = q.Encode()

	doc := g.LoadDoc(u.String())
	pid := g.FindLast(*doc)

	var wg sync.WaitGroup
	imgChan := make(chan common.Image)
	limit := make(chan bool, 3)
	pullPage := func(url string) {
		limit <- true
		doc = g.LoadDoc(url)
		selection := doc.Find("span.thumb a")

		selection.Each(func(i int, s *goquery.Selection) {
			url, exists := s.Attr("href")
			if exists {
				img := g.ParsePost(url)
				imgChan <- img
			}
		})
		wg.Done()
		<-limit
	}
	for i := 0; i <= pid; i += 42 {
		q := u.Query()
		q.Set("pid", strconv.Itoa(i))
		u.RawQuery = q.Encode()
		wg.Add(1)
		go pullPage(u.String())
	}
	go func() {
		wg.Wait()
		close(imgChan)
	}()
	for img := range imgChan {
		imgs = append(imgs, img)
	}
	return
}

func (g *R34xGallery) FindLast(doc goquery.Document) int {
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
func (g *R34xGallery) ParsePost(post string) common.Image {
	base := "https://rule34.xxx/"
	doc := g.LoadDoc(base + post)
	selection := doc.Find("meta[property='og:image']").First()
	output, _ := selection.Attr("content")
	return common.NewImage(output)
}
