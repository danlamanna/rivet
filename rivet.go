package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danlamanna/rivet/config"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/templates"
	"github.com/danlamanna/rivet/transfer"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "undefined"

	app     = kingpin.New("rivet", "sync files to girder")
	auth    = app.Flag("auth", "Authentication credentials, can be username:password, a token, or an api key`").Envar("RIVET_AUTH").Short('a').String()
	url     = app.Flag("url", "URL of the girder instance, e.g. data.kitware.com, somedomain.com/api/v1").Envar("RIVET_URL").Short('u').String()
	verbose = app.Flag("verbose", "Increase verbosity, can be passed up to two times.").Short('v').Counter()

	configure = app.Command("configure", "")

	sync   = app.Command("sync", "sync a local directory to a girder folder")
	source = sync.Arg("source", "source directory").Required().String()
	dest   = sync.Arg("dest", "dest girder folder").Required().String()

	versionCmd = app.Command("version", "")
)

func main() {
	app.HelpFlag.Short('h')
	app.UsageTemplate(templates.DefaultUsageTemplate)
	app.Version(version)

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

	profile, err := config.ReadDefaultProfile()
	if err != nil {
		log.Fatal(err)
	}
	if *auth == "" && profile != nil {
		auth = &profile.Auth
	}
	if *url == "" && profile != nil {
		url = &profile.URL
	}

	if res == "" {
		fmt.Printf(`usage: rivet [options] [subcommand] [arguments]
To see help text, you can run:

rivet help
rivet help <subcommand>`)
		os.Exit(1)
	}

	if res == "configure" {
		reader := bufio.NewReader(os.Stdin)
		var promptedURL string
		for {
			fmt.Print("girder url (e.g. data.kitware.com): ")
			promptedURL, _ = reader.ReadString('\n')
			promptedURL = strings.TrimSpace(promptedURL)

			if promptedURL != "" {
				break
			}
		}

		validURL, err := girder.GetValidURL(promptedURL)

		if err != nil {
			log.Fatal(err)
		}
		girderCtx := new(girder.Context)
		girderCtx.URL = validURL
		girderCtx.Logger = logrus.New()
		if err = girderCtx.CheckMinimumVersion(); err != nil {
			log.Fatal(err)
		}
		var promptedAuth string
		for {
			fmt.Print("auth credentials (e.g. username:password, token, api-key): ")
			promptedAuth, _ = reader.ReadString('\n')
			promptedAuth = strings.TrimSpace(promptedAuth)

			if promptedAuth != "" {
				break
			}
		}
		girderCtx.Auth = promptedAuth
		if err = girderCtx.ValidateAuth(); err != nil {
			log.Fatal(err)
		}
		config.WriteDefaultProfile(promptedAuth, validURL)
	} else if res == "sync" {
		*source = strings.TrimSuffix(*source, "/")

		if stat, err := os.Stat(*source); err != nil {
			if os.IsNotExist(err) {
				log.Fatalf("source directory %s does not exist.\n", *source)
			} else {
				log.Fatalf("failed to access source directory %s, err: %s.\n", *source, err)
			}
		} else if !stat.IsDir() {
			log.Fatalf("source %s is not a directory.\n", *source)
		}

		if *auth == "" {
			fmt.Println("See --auth flag")
			os.Exit(1)
		} else if *url == "" {
			fmt.Println("See --url flag")
			os.Exit(1)
		}

		validURL, err := girder.GetValidURL(*url)

		if err != nil {
			log.Fatal(err)
		}

		girderCtx := new(girder.Context)
		girderCtx.Auth = *auth
		girderCtx.URL = validURL
		girderCtx.Logger = logrus.New()
		girderCtx.ResourceMap = make(girder.ResourceMap)
		girderCtx.Destination = strings.TrimPrefix(*dest, "girder://")
		if *verbose >= 2 {
			girderCtx.Logger.Level = logrus.TraceLevel
		} else if *verbose == 1 {

			girderCtx.Logger.Level = logrus.DebugLevel
		} else {
			girderCtx.Logger.Level = logrus.InfoLevel
		}
		if err = girderCtx.CheckMinimumVersion(); err != nil {
			log.Fatal(err)
		}
		if err = girderCtx.ValidateAuth(); err != nil {
			log.Fatal(err)
		}

		transfer.Upload(girderCtx, *source, girder.GirderID(*dest))
	} else if res == "version" {
		fmt.Printf("rivet %s", version)
	}
}
