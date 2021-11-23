package main

import (
	"github.com/alecthomas/kong"
	"github.com/darkdragn/gocyberdrop/common"
	"github.com/darkdragn/gocyberdrop/sites/cyberdrop"
	"github.com/darkdragn/gocyberdrop/sites/rule34xxx"
	log "github.com/sirupsen/logrus"
)

type DownloadCmd struct {
	Url []string `arg:"" name:"url" help:"The gallery URL to download from."`
}

type Rule34xxxCmd struct {
	Tag string `arg:"" name:"tag" help:"The gallery tag to download from."`
}

func (r *DownloadCmd) Run(logger *log.Logger) error {
	c := common.New(logger, 15)
	var gals []cyberdrop.CyberdropGallery
	for _, url := range r.Url {
		g := cyberdrop.CyberdropGallery{
			Client: c,
			Url:    url,
		}
		gals = append(gals, g)
	}
	for _, gal := range gals {
		err := common.PullGallery(&gal)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rule34xxxCmd) Run(logger *log.Logger) error {
	c := common.New(logger, 15)
	g := rule34xxx.R34xGallery{
		Client: c,
		Tag:    r.Tag,
	}
	err := common.PullGallery(&g)
	if err != nil {
		return err
	}
	return nil
}

var cli struct {
	Debug     bool         `help:"Run the logger in debug mode."`
	Download  DownloadCmd  `cmd:"" help:"Download a cyberdrop gallery"`
	Rule34xxx Rule34xxxCmd `cmd:"" help:"Download a cyberdrop gallery"`
}

func generateLogger() *log.Logger {
	var logLevel log.Level
	if cli.Debug {
		logLevel = log.DebugLevel
	} else {
		logLevel = log.InfoLevel
	}
	logger := log.New()
	logger.SetLevel(logLevel)
	return logger
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
	err := ctx.Run(generateLogger())
	ctx.FatalIfErrorf(err)
}
