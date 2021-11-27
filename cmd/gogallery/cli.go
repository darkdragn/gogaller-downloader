package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/darkdragn/gogallery-downloader/common"
	"github.com/darkdragn/gogallery-downloader/sites/cyberdrop"
	"github.com/darkdragn/gogallery-downloader/sites/rule34xxx"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/chalk"
)

var appVersion string
var buildTime string

type VersionCmd struct{}

type CyberdropCmd struct {
	Url []string `arg:"" name:"url" help:"The gallery URL to download from."`
}

type Rule34xxxCmd struct {
	Tag string `arg:"" name:"tag" help:"The gallery tag to download from."`
}

func (r *CyberdropCmd) Run(logger *log.Logger) error {
	c := common.New(logger, 15)
	var gals []cyberdrop.CyberdropGallery
	for _, url := range r.Url {
		g := cyberdrop.CyberdropGallery{
			GalleryBase: common.GalleryBase{Client: c},
			Url:         url,
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
	c := common.New(logger, 50)
	g := rule34xxx.R34xGallery{
		GalleryBase: common.GalleryBase{Client: c},
		Tag:         r.Tag,
	}
	err := common.PullGallery(&g)
	if err != nil {
		return err
	}
	return nil
}

func (r *VersionCmd) Run(logger *log.Logger) error {
	// logger.Printf(`GoGallery Build Info:
	lime := chalk.Green.NewStyle().
		WithBackground(chalk.Black).
		WithTextStyle(chalk.Bold).
		Style
	fmt.Println(chalk.Cyan, "GoGallery Build Info:")
	fmt.Println(chalk.Magenta, "    Build time: ", lime(buildTime))
	fmt.Println(chalk.Magenta, "    Version:    ", lime(appVersion))
	return nil
}

var cli struct {
	Debug     bool         `help:"Run the logger in debug mode."`
	Cyberdrop CyberdropCmd `cmd:"" help:"Download a cyberdrop gallery"`
	Rule34xxx Rule34xxxCmd `cmd:"" help:"Download a rule34.xxx gallery" name:"r34xxx"`
	Version   VersionCmd   `cmd:"" help:"Print version"`
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
A command for downloading from online galleries.
	Note:
		To find some good ones, try google searching: site:cyberdrop.me $WhoYouWant
		For example: site:cyberdrop.me carrykey`,
		),
		kong.UsageOnError(),
	)
	err := ctx.Run(generateLogger())
	ctx.FatalIfErrorf(err)
}
