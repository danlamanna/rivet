package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danlamanna/rivet/config"
	"github.com/danlamanna/rivet/download"
	"github.com/danlamanna/rivet/girder"
	"github.com/danlamanna/rivet/templates"
	"github.com/danlamanna/rivet/transfer"
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

	configure = app.Command("configure", "")

	sync   = app.Command("sync", "sync a local directory to or from a girder folder")
	source = sync.Arg("source", "source directory or girder folder").Required().String()
	dest   = sync.Arg("dest", "dest directory or girder folder").Required().String()

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

	if newerVersion, _ := version.IsNewVersionAvailable(); newerVersion != "" {
		fmt.Printf("Your version of rivet (v%s) is out of date.\n", version.Version)
		fmt.Printf("Download the newest version (v%s) from https://github.com/danlamanna/rivet/releases.\n\n", newerVersion)
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
		girderCtx.Logger.SetFormatter(&log.TextFormatter{
			DisableLevelTruncation: true,
			FullTimestamp:          true,
		})
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
		*dest = strings.TrimSuffix(*dest, "/")

		sourceIsGirder := strings.HasPrefix(*source, "girder://")
		destIsGirder := strings.HasPrefix(*dest, "girder://")
		if sourceIsGirder && destIsGirder {
			log.Fatal("cannot sync between two girder instances")
		} else if !sourceIsGirder && !destIsGirder {
			log.Fatal("cannot sync between two local directories")
		}
		if destIsGirder {
			if stat, err := os.Stat(*source); err != nil {
				if os.IsNotExist(err) {
					log.Fatalf("source directory %s does not exist.\n", *source)
				} else {
					log.Fatalf("failed to access source directory %s, err: %s.\n", *source, err)
				}
			} else if !stat.IsDir() {
				log.Fatalf("source %s is not a directory.\n", *source)
			}
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

		girderCtx.Logger.SetFormatter(&log.TextFormatter{
			DisableLevelTruncation: true,
			FullTimestamp:          true,
		})
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

		if destIsGirder {
			transfer.Upload(girderCtx, *source, girder.GirderID(*dest))
		} else if sourceIsGirder {
			download.Download(girderCtx, girder.GirderID(strings.TrimPrefix(*source, "girder://")), *dest)
		}
	} else if res == "version" {
		fmt.Printf("rivet       v%s\n", version.Version)
		fmt.Printf("build:      %s\n", version.GitCommit)
		fmt.Printf("built:      %s\n", version.BuildDate)
		fmt.Printf("go version: %s\n", version.GoVersion)
		fmt.Printf("os/arch:    %s\n", version.OsArch)
	}
}
