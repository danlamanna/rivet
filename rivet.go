package main

import (
	"fmt"
	"os"

	"github.com/danlamanna/rivet/commands"
	"github.com/danlamanna/rivet/config"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/templates"
	"github.com/danlamanna/rivet/version"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("rivet", "sync files to girder")
	auth    = app.Flag("auth", "Authentication credentials, can be username:password, a token, or an api key`").Envar("RIVET_AUTH").Short('a').String()
	url     = app.Flag("url", "URL of the girder instance, e.g. data.kitware.com, somedomain.com/api/v1").Envar("RIVET_URL").Short('u').String()
	verbose = app.Flag("verbose", "Increase verbosity, can be passed up to two times.").Short('v').Counter()

	// hidden global flags
	noConfigFile = app.Flag("no-config", "Skip loading a configuration file").Bool()

	// configure command
	configure = app.Command("configure", "")

	// sync command
	sync   = app.Command("sync", "sync a local directory to or from a girder folder")
	source = sync.Arg("source", "source directory or girder folder").Required().String()
	dest   = sync.Arg("dest", "dest directory or girder folder").Required().String()

	// version command
	versionCmd = app.Command("version", "")
)

func main() {
	app.HelpFlag.Short('h')
	app.UsageTemplate(templates.DefaultUsageTemplate)
	app.Version(version.Version)

	// kingpin doesn't allow help for subcommands, so we hook in before parsing
	// to possible show help pages
	if len(os.Args) >= 3 && os.Args[1] == "help" && os.Args[2] == "configure" {
		fmt.Print(fmt.Errorf(templates.ConfigureUsageTemplate))
		os.Exit(1)
	} else if len(os.Args) >= 3 && os.Args[1] == "help" && os.Args[2] == "sync" {
		fmt.Print(fmt.Errorf(templates.SyncUsageTemplate))
		os.Exit(1)
	}

	res, _ := app.Parse(os.Args[1:])

	// setup context/logger
	ctx := new(girder.Context)
	ctx.Logger = logrus.New()
	ctx.Logger.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		FullTimestamp:          true,
	})
	if *verbose >= 2 {
		ctx.Logger.Level = logrus.TraceLevel
	} else if *verbose == 1 {
		ctx.Logger.Level = logrus.DebugLevel
	} else {
		ctx.Logger.Level = logrus.InfoLevel
	}

	// fill in auth/url, defaulting to the "default" profile
	if !*noConfigFile {
		profile, err := config.ReadDefaultProfile()
		if err != nil {
			log.Fatal(err)
		}

		if profile != nil {
			ctx.Auth = profile.Auth
			ctx.URL = profile.URL
		}
	}

	// override default profile with envvars/flags
	if *auth != "" {
		ctx.Auth = *auth
	} else if *url != "" {
		ctx.URL = *url
	} else if res == "" {
		fmt.Printf(`usage: rivet [options] [subcommand] [arguments]
To see help text, you can run:

rivet help
rivet help <subcommand>`)
		os.Exit(1)
	}

	if newerVersion, _ := version.IsNewVersionAvailable(); newerVersion != "" {
		fmt.Printf("Your version of rivet (v%s) is out of date.\n", version.Version)
		fmt.Printf("Download the newest version (v%s) from https://github.com/danlamanna/rivet/releases.\n\n", newerVersion)
	}

	switch res {
	case "configure":
		commands.Configure(ctx)
	case "sync":
		if ctx.Auth == "" {
			fmt.Println("See --auth flag")
			os.Exit(1)
		} else if ctx.URL == "" {
			fmt.Println("See --url flag")
			os.Exit(1)
		}

		validURL, err := girder.GetValidURL(ctx.URL)
		ctx.URL = validURL

		if err != nil {
			log.Fatal(err)
		}

		commands.Sync(ctx, source, dest)
	case "version":
		commands.Version()
	}
}
