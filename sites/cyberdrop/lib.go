package cyberdrop

import (

	// "log"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darkdragn/gogallery-downloader/common"
	log "github.com/sirupsen/logrus"
)

type CyberdropGallery struct {
	Client common.Client
	Url    string
	doc    *goquery.Document
}

func (g *CyberdropGallery) Title() string {
	doc := g.Document()
	title, exists := doc.Find("h1").Attr("title")
	if !exists {
		g.Client.Logger.Fatal("Unable to find a title")
	}
	return title
}

func (g *CyberdropGallery) GetClient() common.Client {
	return g.Client
}

func (g *CyberdropGallery) Logger() *log.Logger {
	return g.Client.Logger
}

func (g *CyberdropGallery) Document() *goquery.Document {
	if g.doc == nil {
		g.doc = g.Client.LoadDoc(g.Url)
	}
	return g.doc
}

func (g *CyberdropGallery) ImageList() (imgs []common.Image) {
	doc := g.Document()
	selection := doc.Find("a.image")
	selection.Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if exists {
			p := strings.Split(url, "/")
			fn := p[len(p)-1]
			imgs = append(imgs, common.Image{
				Url:      url,
				Filename: fn,
			})
		}
	})
	return
}
