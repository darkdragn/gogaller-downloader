package common

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Logger *log.Logger
	Client http.Client
}

func New(logger *log.Logger, maxConnsPerHost int) Client {
	transport := &http.Transport{
		TLSNextProto:    map[string]func(string, *tls.Conn) http.RoundTripper{},
		MaxIdleConns:    10,
		MaxConnsPerHost: maxConnsPerHost,
		IdleConnTimeout: 10 * time.Second,
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

func (c *Client) LoadDoc(url string) *goquery.Document {
	res, err := c.Client.Get(url)
	c.Catch(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.Logger.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	c.Catch(err)
	return doc
}

func (c *Client) PullImage(s string, filename string, folder string, completion chan bool, wg *sync.WaitGroup) {
	fn := path.Join(folder, filename)

	pullAndWrite := func() error {
		if strings.Contains(s, "web.archive") {
			s = strings.Replace(s, "/https", "if_/https", 1)
		}
		req, _ := http.NewRequest("GET", s, nil)
		req.Header.Set("Cache-Control", "max-age=0")
		req.Header.Set("Accept-Encoding", "identity,compress,deflate,gzip")
		req.Close = true
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
		lastModified := resp.Header.Get("Last-Modified")
		if lastModified != "" {
			format := "Mon, 02 Jan 2006 15:04:05 MST"
			t, _ := time.Parse(format, lastModified)
			err = os.Chtimes(fn, t, t)
			c.Catch(err)
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
		}
	} else {
		completion <- true
	}
	wg.Done()
}
