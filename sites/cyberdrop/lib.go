package cyberdrop

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darkdragn/gogallery-downloader/common"
)

type CyberdropGallery struct {
	common.GalleryBase
	Url string
	doc *goquery.Document
}

func (g *CyberdropGallery) Title() string {
	doc := g.Document()
	title, exists := doc.Find("h1").Attr("title")
	if !exists {
		g.Logger().Fatal("Unable to find a title")
	}
	return title
}

func (g *CyberdropGallery) Document() *goquery.Document {
	if g.doc == nil {
		g.doc = g.LoadDoc(g.Url)
	}
	return g.doc
}

func (g *CyberdropGallery) ImageList() (imgs []common.Image) {
	doc := g.Document()
	selection := doc.Find("a.image")
	selection.Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if exists {
			filename, exists := s.Attr("title")
			if !exists {
				p := strings.Split(url, "/")
				filename = p[len(p)-1]
			}
			imgs = append(imgs, common.Image{
				Url:      url,
				Filename: filename,
			})
		}
	})
	return
}
