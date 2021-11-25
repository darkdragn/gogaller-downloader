package common

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type Gallery interface {
	GetClient() Client
	Logger() *log.Logger
	Title() string
	ImageList() []Image
}

type GalleryBase struct {
	Client
}

func (g *GalleryBase) GetClient() Client {
	return g.Client
}

func (g *GalleryBase) Logger() *log.Logger {
	return g.Client.Logger
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
