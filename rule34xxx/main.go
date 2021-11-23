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

type Gallery interface {
	GetClient() gocyberdrop.Client
	Title() string
	ImageList() []Image
	Logger() *log.Logger
}

type R34xGallery struct {
	Client gocyberdrop.Client
	Tag    string
}

func (g *R34xGallery) Title() string {
	return g.Tag
}

func (g *R34xGallery) GetClient() gocyberdrop.Client {
	return g.Client
}

func (g *R34xGallery) Logger() *log.Logger {
	return g.Client.Logger
}

func (g *R34xGallery) ImageList() (imgs []Image) {
	u, _ := url.Parse("https://rule34.xxx/index.php?page=post&s=list")
	q := u.Query()
	q.Set("tags", g.Tag)
	u.RawQuery = q.Encode()

	doc := g.Client.LoadDoc(u.String())
	pid := g.FindLast(*doc)

	var wg sync.WaitGroup
	imgChan := make(chan Image)
	limit := make(chan bool, 3)
	pullPage := func(url string) {
		limit <- true
		doc = g.Client.LoadDoc(url)
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

func PullGallery(g Gallery) error {

	title := g.Title()
	logger := g.Logger()
	logger.Info(title)
	os.Mkdir(title, 0755)

	var wg sync.WaitGroup
	client := g.GetClient()
	completion := make(chan bool, 1)
	count := int64(0)
	imgs := g.ImageList()

	bar := progressbar.Default(int64(len(imgs)))
	limit := make(chan bool, 50)
	go func() {
		for range completion {
			count += 1
			bar.Add(1)
			<-limit
		}
	}()
	for _, img := range imgs {
		wg.Add(1)
		limit <- true
		go client.PullImage(img.Url, img.Filename, title, completion, &wg)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	if count < int64(len(imgs)) {
		fmt.Printf("\n")
		logger.
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
func (g *R34xGallery) ParsePost(post string) Image {
	base := "https://rule34.xxx/"
	doc := g.Client.LoadDoc(base + post)
	selection := doc.Find("meta[property='og:image']").First()
	output, _ := selection.Attr("content")
	return NewImage(output)
}
