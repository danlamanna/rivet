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

	profile, err := config.ReadDefaultProfile()
	if err != nil {
		log.Fatal(err)
	}
	if *auth != "" {
		ctx.Auth = *auth
	} else if profile != nil {
		ctx.Auth = profile.Auth
	}

	if *url != "" {
		ctx.URL = *url
	} else if profile != nil {
		ctx.URL = profile.URL
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

	switch res {
	case "configure":
		configureCommand(ctx)
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

		syncCommand(ctx, source, dest)
	case "version":
		versionCommand()
	}

}

func configureCommand(ctx *girder.Context) {
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
		ctx.Logger.Fatal(err)
	}

	if err = ctx.CheckMinimumVersion(); err != nil {
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
	ctx.Auth = promptedAuth
	if err = ctx.ValidateAuth(); err != nil {
		log.Fatal(err)
	}
	config.WriteDefaultProfile(promptedAuth, validURL)
}

func syncCommand(ctx *girder.Context, source *string, dest *string) {
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

	ctx.ResourceMap = make(girder.ResourceMap)
	ctx.Destination = strings.TrimPrefix(*dest, "girder://")

	if err := ctx.CheckMinimumVersion(); err != nil {
		log.Fatal(err)
	}
	if err := ctx.ValidateAuth(); err != nil {
		log.Fatal(err)
	}

	if destIsGirder {
		transfer.Upload(ctx, *source, girder.GirderID(*dest))
	} else if sourceIsGirder {
		transfer.Download(ctx, girder.GirderID(strings.TrimPrefix(*source, "girder://")), *dest)
	}
}

func versionCommand() {
	fmt.Printf("rivet       v%s\n", version.Version)
	fmt.Printf("build:      %s\n", version.GitCommit)
	fmt.Printf("built:      %s\n", version.BuildDate)
	fmt.Printf("go version: %s\n", version.GoVersion)
	fmt.Printf("os/arch:    %s\n", version.OsArch)
}
