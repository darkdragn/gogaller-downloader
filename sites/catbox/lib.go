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
	parseAndAdd := func(attr string) func(int, *goquery.Selection) {
		return func(i int, s *goquery.Selection) {
			url, exists := s.Attr(attr)
			if exists {
				p := strings.Split(url, "/")
				filename := p[len(p)-1]
				imgs = append(imgs, common.Image{
					Url:      url,
					Filename: filename,
				})
			}
		}
	}
	doc := g.LoadDoc(g.Url)
	selection := doc.Find(".imagecontainer a")
	selection.Each(parseAndAdd("href"))
	vidSelection := doc.Find("video")
	vidSelection.Each(parseAndAdd("src"))
	return
}
