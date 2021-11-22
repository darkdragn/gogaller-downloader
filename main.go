package gocyberdrop

import (
	"errors"
	"fmt"
	"time"

	// "log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Logger log.Logger
	Client http.Client
}

func New(logger log.Logger) Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     5 * time.Second,
	}
	return Client{
		Logger: logger,
		Client: http.Client{
			Transport: transport,
		},
	}
}

func (c *Client) Catch(err error) {
	if err != nil {
		c.Logger.Panic(err)
	}
}

func (c *Client) PullGallery(s string) error {
	doc := c.LoadDoc(s)

	var wg sync.WaitGroup
	completion := make(chan bool, 1)
	count := int64(0)

	title, exists := doc.Find("h1").Attr("title")
	if exists {
		c.Logger.Info(title)
		os.Mkdir(title, 0755)
	}
	selection := doc.Find("a.image")
	bar := progressbar.Default(int64(selection.Length()))
	selection.Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if exists {
			wg.Add(1)
			p := strings.Split(url, "/")
			fn := p[len(p)-1]
			go c.PullImage(url, fn, title, completion, &wg)
		}
	})
	go func() {
		for range completion {
			count += 1
			bar.Add(1)
		}
	}()
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	if count < int64(selection.Length()) {
		fmt.Printf("\n")
		c.Logger.
			WithFields(log.Fields{
				"Total":   selection.Length(),
				"Current": count,
			}).
			Error("Unable to download all, please retry")
		os.Exit(1)
	} else {
		bar.Set(int(count))
	}
	return nil
}

func (c *Client) PullImage(s string, filename string, folder string, completion chan bool, wg *sync.WaitGroup) {
	fn := path.Join(folder, filename)

	pullAndWrite := func() error {
		if strings.Contains(s, "web.archive") {
			s = strings.Replace(s, "/https", "if_/https", 1)
		}
		req, _ := http.NewRequest("GET", s, nil)
		req.Header.Set("Connection", "keep-alive")
		resp, err := c.Client.Do(req)
		if err != nil {
			return errors.New("couldn't open url; " + err.Error())
		}
		defer resp.Body.Close()

		out, err := os.Create(fn)
		c.Catch(err)

		if _, err := out.ReadFrom(resp.Body); err != nil {
			out.Close()
			os.Remove(fn)
			return errors.New("failed to pull; " + err.Error())
		} else {
			out.Close()
		}
		return nil
	}
	if _, err := os.Stat(fn); errors.Is(err, os.ErrNotExist) {
		for i := range make([]bool, 5) {
			err := pullAndWrite()
			if err == nil {
				completion <- true
				break
			}
			c.Logger.
				WithFields(log.Fields{
					"URL":   s,
					"Retry": i,
					"Error": err,
				}).
				Debug("Failed download")
			//time.Sleep(500 * time.Millisecond)
		}
	} else {
		completion <- true
	}
	wg.Done()
}

func (c *Client) LoadDoc(url string) *goquery.Document {
	res, err := c.Client.Get(url)
	c.Catch(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.Logger.Fatal("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	c.Catch(err)
	return doc
}
