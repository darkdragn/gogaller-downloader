package catbox

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darkdragn/gogallery-downloader/common"
)

type CatboxGallery struct {
	common.GalleryBase
	Url string
}

func (g *CatboxGallery) Title() string {
	return "."
}

func (g *CatboxGallery) ImageList() (imgs []common.Image) {
	doc := g.LoadDoc(g.Url)
	selection := doc.Find(".imagecontainer a")
	selection.Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if exists {
			p := strings.Split(url, "/")
			filename := p[len(p)-1]
			imgs = append(imgs, common.Image{
				Url:      url,
				Filename: filename,
			})
		}
	})
	return
}
