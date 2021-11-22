package main

import (
	"github.com/alecthomas/kong"
	"github.com/darkdragn/gocyberdrop"
	log "github.com/sirupsen/logrus"
)

type DownloadCmd struct {
	Url []string `arg:"" name:"url" help:"The gallery URL to download from."`
}

func (r *DownloadCmd) Run(debug bool) error {
	var logLevel log.Level
	if debug {
		logLevel = log.DebugLevel
	} else {
		logLevel = log.InfoLevel
	}
	logger := log.New()
	logger.SetLevel(logLevel)
	c := gocyberdrop.New(*logger)
	for _, url := range r.Url {
		err := c.PullGallery(url)
		if err != nil {
			return err
		}
	}
	return nil
}

var cli struct {
	Debug bool `help:"Run the logger in debug mode."`
	Download DownloadCmd `cmd:"" help:"Download a cyberdrop gallery"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Description(`
A command for downloading from cyberdrop.me.
  Note: 
    To find some good ones, try google searching: site:cyberdrop.me $WhoYouWant
    For example: site:cyberdrop.me carrykey`,
		),
		kong.UsageOnError(),
	)
	err := ctx.Run(cli.Debug)
	ctx.FatalIfErrorf(err)
}
